//go:build !tinygo.wasm && !aho_corasick_cgo

package aho_corasick

import (
	"context"
	_ "embed"
	"encoding/binary"
	"errors"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

var (
	errFailedWrite = errors.New("failed to read from wasm memory")
	errFailedRead  = errors.New("failed to read from wasm memory")
)

//go:embed wasm/aho_corasick.wasm
var ahocorasickWasm []byte

var (
	wasmRT       wazero.Runtime
	wasmCompiled wazero.CompiledModule
)

type ahoCorasickABI struct {
	new_matcher             api.Function
	delete_matcher          api.Function
	find_iter               api.Function
	find_iter_next          api.Function
	find_iter_delete        api.Function
	overlapping_iter        api.Function
	overlapping_iter_next   api.Function
	overlapping_iter_delete api.Function
	matches                 api.Function
	matches_delete          api.Function

	malloc api.Function
	free   api.Function

	wasmMemory api.Memory

	mod       api.Module
	callStack []uint64

	memory sharedMemory
	mu     sync.Mutex
}

func init() {
	ctx := context.Background()
	rt := wazero.NewRuntime(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	code, err := rt.CompileModule(ctx, ahocorasickWasm)
	if err != nil {
		panic(err)
	}
	wasmCompiled = code

	wasmRT = rt
}

func newABI() *ahoCorasickABI {
	ctx := context.Background()
	mod, err := wasmRT.InstantiateModule(ctx, wasmCompiled, wazero.NewModuleConfig().WithName(""))
	if err != nil {
		panic(err)
	}

	callStack := make([]uint64, 6)

	return &ahoCorasickABI{
		new_matcher:             mod.ExportedFunction("new_matcher"),
		delete_matcher:          mod.ExportedFunction("delete_matcher"),
		find_iter:               mod.ExportedFunction("find_iter"),
		find_iter_next:          mod.ExportedFunction("find_iter_next"),
		find_iter_delete:        mod.ExportedFunction("find_iter_delete"),
		overlapping_iter:        mod.ExportedFunction("overlapping_iter"),
		overlapping_iter_next:   mod.ExportedFunction("overlapping_iter_next"),
		overlapping_iter_delete: mod.ExportedFunction("overlapping_iter_delete"),
		matches:                 mod.ExportedFunction("matches"),
		matches_delete:          mod.ExportedFunction("matches_delete"),

		malloc: mod.ExportedFunction("malloc"),
		free:   mod.ExportedFunction("free"),

		wasmMemory: mod.Memory(),
		mod:        mod,
		callStack:  callStack,
	}
}

func (abi *ahoCorasickABI) startOperation(memorySize int) {
	abi.mu.Lock()
	abi.memory.reserve(abi, uint32(memorySize))
}

func (abi *ahoCorasickABI) endOperation() {
	abi.mu.Unlock()
}

func (abi *ahoCorasickABI) newMatcher(patterns []string, patternBytes int, asciiCaseInsensitive bool, dfa bool, matchKind int) uintptr {
	patternsPtr := abi.memory.allocate(uint32(patternBytes))
	lensPtr := abi.memory.allocate(uint32(len(patterns) * 4))

	buf, ok := abi.wasmMemory.Read(uint32(patternsPtr), uint32(patternBytes))
	if !ok {
		panic(errFailedRead)
	}
	off := 0
	for i, p := range patterns {
		copy(buf[off:], p)
		off += len(p)
		abi.wasmMemory.WriteUint32Le(uint32(lensPtr)+uint32(i*4), uint32(len(p)))
	}

	aci := 0
	if asciiCaseInsensitive {
		aci = 1
	}

	d := 0
	if dfa {
		d = 1
	}

	callStack := abi.callStack
	callStack[0] = uint64(patternsPtr)
	callStack[1] = uint64(lensPtr)
	callStack[2] = uint64(len(patterns))
	callStack[3] = uint64(aci)
	callStack[4] = uint64(d)
	callStack[5] = uint64(matchKind)
	if err := abi.new_matcher.CallWithStack(context.Background(), callStack); err != nil {
		panic(err)
	}

	return uintptr(callStack[0])
}

func (abi *ahoCorasickABI) deleteMatcher(ptr uintptr) {
	callStack := abi.callStack
	callStack[0] = uint64(ptr)
	if err := abi.delete_matcher.CallWithStack(context.Background(), callStack); err != nil {
		panic(err)
	}
}

func (abi *ahoCorasickABI) findIter(acPtr uintptr, value cString) uintptr {
	callStack := abi.callStack
	callStack[0] = uint64(acPtr)
	callStack[1] = uint64(value.ptr)
	callStack[2] = uint64(value.length)
	if err := abi.find_iter.CallWithStack(context.Background(), callStack); err != nil {
		panic(err)
	}

	return uintptr(callStack[0])
}

func (abi *ahoCorasickABI) findIterNext(iter uintptr) (int, int, int, bool) {
	patternPtr := abi.memory.allocate(4)
	startPtr := abi.memory.allocate(4)
	endPtr := abi.memory.allocate(4)

	callStack := abi.callStack
	callStack[0] = uint64(iter)
	callStack[1] = uint64(patternPtr)
	callStack[2] = uint64(startPtr)
	callStack[3] = uint64(endPtr)
	if err := abi.find_iter_next.CallWithStack(context.Background(), callStack); err != nil {
		panic(err)
	}

	if callStack[0] == 0 {
		return 0, 0, 0, false
	}

	pattern, ok := abi.wasmMemory.ReadUint32Le(uint32(patternPtr))
	if !ok {
		panic(errFailedRead)
	}
	start, ok := abi.wasmMemory.ReadUint32Le(uint32(startPtr))
	if !ok {
		panic(errFailedRead)
	}
	end, ok := abi.wasmMemory.ReadUint32Le(uint32(endPtr))
	if !ok {
		panic(errFailedRead)
	}

	return int(pattern), int(start), int(end), true
}

func (abi *ahoCorasickABI) findIterDelete(iter uintptr) {
	callStack := abi.callStack
	callStack[0] = uint64(iter)
	if err := abi.find_iter_delete.CallWithStack(context.Background(), callStack); err != nil {
		panic(err)
	}
}

func (abi *ahoCorasickABI) overlappingIter(acPtr uintptr, value cString) uintptr {
	callStack := abi.callStack
	callStack[0] = uint64(acPtr)
	callStack[1] = uint64(value.ptr)
	callStack[2] = uint64(value.length)
	if err := abi.overlapping_iter.CallWithStack(context.Background(), callStack); err != nil {
		panic(err)
	}

	return uintptr(callStack[0])
}

func (abi *ahoCorasickABI) overlappingIterNext(iter uintptr) (int, int, int, bool) {
	patternPtr := abi.memory.allocate(4)
	startPtr := abi.memory.allocate(4)
	endPtr := abi.memory.allocate(4)

	callStack := abi.callStack
	callStack[0] = uint64(iter)
	callStack[1] = uint64(patternPtr)
	callStack[2] = uint64(startPtr)
	callStack[3] = uint64(endPtr)
	if err := abi.overlapping_iter_next.CallWithStack(context.Background(), callStack); err != nil {
		panic(err)
	}

	if callStack[0] == 0 {
		return 0, 0, 0, false
	}

	pattern, _ := abi.wasmMemory.ReadUint32Le(uint32(patternPtr))
	start, _ := abi.wasmMemory.ReadUint32Le(uint32(startPtr))
	end, _ := abi.wasmMemory.ReadUint32Le(uint32(endPtr))

	return int(pattern), int(start), int(end), true
}

func (abi *ahoCorasickABI) overlappingIterDelete(iter uintptr) {
	callStack := abi.callStack
	callStack[0] = uint64(iter)
	if err := abi.overlapping_iter_delete.CallWithStack(context.Background(), callStack); err != nil {
		panic(err)
	}
}

func (abi *ahoCorasickABI) findN(iter uintptr, valueStr string, value cString, n int, matchWholeWords bool) []Match {
	lenPtr := abi.memory.allocate(4)

	callStack := abi.callStack
	callStack[0] = uint64(iter)
	callStack[1] = uint64(value.ptr)
	callStack[2] = uint64(value.length)
	callStack[3] = uint64(n)
	callStack[4] = uint64(lenPtr)
	if err := abi.matches.CallWithStack(context.Background(), callStack); err != nil {
		panic(err)
	}

	resLen, ok := abi.wasmMemory.ReadUint32Le(uint32(lenPtr))
	if !ok {
		panic(errFailedRead)
	}

	resPtr := callStack[0]
	defer func() {
		callStack[0] = uint64(resPtr)
		callStack[1] = uint64(resLen)
		if err := abi.matches_delete.CallWithStack(context.Background(), callStack); err != nil {
			panic(err)
		}
	}()

	res, ok := abi.wasmMemory.Read(uint32(resPtr), resLen*4)
	if !ok {
		panic(errFailedRead)
	}

	num := resLen / 3
	matches := make([]Match, 0, num)
	for i := 0; i < int(num); i++ {
		start := int(binary.LittleEndian.Uint32(res[i*12+4:]))
		end := int(binary.LittleEndian.Uint32(res[i*12+8:]))
		if matchWholeWords && isNotWholeWord(valueStr, start, end) {
			continue
		}
		var m Match
		m.pattern = int(binary.LittleEndian.Uint32(res[i*12:]))
		m.start = start
		m.end = end
		matches = append(matches, m)
	}

	return matches
}

type sharedMemory struct {
	size    uint32
	bufPtr  uint32
	nextIdx uint32
}

func (m *sharedMemory) reserve(abi *ahoCorasickABI, size uint32) {
	m.nextIdx = 0
	if m.size >= size {
		return
	}

	ctx := context.Background()
	callStack := abi.callStack
	if m.bufPtr != 0 {
		callStack[0] = uint64(m.bufPtr)
		if err := abi.free.CallWithStack(ctx, callStack); err != nil {
			panic(err)
		}
	}

	callStack[0] = uint64(size)
	if err := abi.malloc.CallWithStack(ctx, callStack); err != nil {
		panic(err)
	}

	m.size = size
	m.bufPtr = uint32(callStack[0])
}

func (m *sharedMemory) allocate(size uint32) uintptr {
	if m.nextIdx+size > m.size {
		panic("not enough reserved shared memory")
	}

	ptr := m.bufPtr + m.nextIdx
	m.nextIdx += size
	return uintptr(ptr)
}

type cString struct {
	ptr    uintptr
	length int
}

func (abi *ahoCorasickABI) newOwnedCString(s string) cString {
	res, err := abi.malloc.Call(context.Background(), uint64(len(s)))
	if err != nil {
		panic(err)
	}
	ptr := res[0]
	if !abi.wasmMemory.WriteString(uint32(ptr), s) {
		panic(errFailedWrite)
	}
	return cString{
		ptr:    uintptr(ptr),
		length: len(s),
	}
}

func (abi *ahoCorasickABI) freeOwnedCStringPtr(ptr uintptr) {
	callStack := abi.callStack
	callStack[0] = uint64(ptr)
	if err := abi.free.CallWithStack(context.Background(), callStack); err != nil {
		panic(err)
	}
}
