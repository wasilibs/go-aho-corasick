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
	errFailedWrite = errors.New("failed to write to wasm memory")
	errFailedRead  = errors.New("failed to read from wasm memory")
)

//go:embed internal/wasm/aho_corasick.wasm
var ahocorasickWasm []byte

var (
	wasmRT       wazero.Runtime
	wasmCompiled wazero.CompiledModule
)

// ahoCorasickABI holds the handles to the wasm module's exports. The Fn suffix
// distinguishes them from the wrapper methods of the same name defined below.
type ahoCorasickABI struct {
	newMatcherFn            api.Function
	deleteMatcherFn         api.Function
	findIterFn              api.Function
	findIterNextFn          api.Function
	findIterDeleteFn        api.Function
	overlappingIterFn       api.Function
	overlappingIterNextFn   api.Function
	overlappingIterDeleteFn api.Function
	matchesFn               api.Function
	matchesDeleteFn         api.Function

	mallocFn api.Function
	freeFn   api.Function

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
		newMatcherFn:            mod.ExportedFunction("new_matcher"),
		deleteMatcherFn:         mod.ExportedFunction("delete_matcher"),
		findIterFn:              mod.ExportedFunction("find_iter"),
		findIterNextFn:          mod.ExportedFunction("find_iter_next"),
		findIterDeleteFn:        mod.ExportedFunction("find_iter_delete"),
		overlappingIterFn:       mod.ExportedFunction("overlapping_iter"),
		overlappingIterNextFn:   mod.ExportedFunction("overlapping_iter_next"),
		overlappingIterDeleteFn: mod.ExportedFunction("overlapping_iter_delete"),
		matchesFn:               mod.ExportedFunction("matches"),
		matchesDeleteFn:         mod.ExportedFunction("matches_delete"),

		mallocFn: mod.ExportedFunction("malloc"),
		freeFn:   mod.ExportedFunction("free"),

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
	if err := abi.newMatcherFn.CallWithStack(context.Background(), callStack); err != nil {
		panic(err)
	}

	return uintptr(callStack[0])
}

func (abi *ahoCorasickABI) deleteMatcher(ptr uintptr) {
	callStack := abi.callStack
	callStack[0] = uint64(ptr)
	if err := abi.deleteMatcherFn.CallWithStack(context.Background(), callStack); err != nil {
		panic(err)
	}
}

func (abi *ahoCorasickABI) findIter(acPtr uintptr, value cString) uintptr {
	callStack := abi.callStack
	callStack[0] = uint64(acPtr)
	callStack[1] = uint64(value.ptr)
	callStack[2] = uint64(value.length)
	if err := abi.findIterFn.CallWithStack(context.Background(), callStack); err != nil {
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
	if err := abi.findIterNextFn.CallWithStack(context.Background(), callStack); err != nil {
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
	if err := abi.findIterDeleteFn.CallWithStack(context.Background(), callStack); err != nil {
		panic(err)
	}
}

func (abi *ahoCorasickABI) overlappingIter(acPtr uintptr, value cString) uintptr {
	callStack := abi.callStack
	callStack[0] = uint64(acPtr)
	callStack[1] = uint64(value.ptr)
	callStack[2] = uint64(value.length)
	if err := abi.overlappingIterFn.CallWithStack(context.Background(), callStack); err != nil {
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
	if err := abi.overlappingIterNextFn.CallWithStack(context.Background(), callStack); err != nil {
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
	if err := abi.overlappingIterDeleteFn.CallWithStack(context.Background(), callStack); err != nil {
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
	if err := abi.matchesFn.CallWithStack(context.Background(), callStack); err != nil {
		panic(err)
	}

	resLen, ok := abi.wasmMemory.ReadUint32Le(uint32(lenPtr))
	if !ok {
		panic(errFailedRead)
	}

	resPtr := callStack[0]
	defer func() {
		callStack[0] = resPtr
		callStack[1] = uint64(resLen)
		if err := abi.matchesDeleteFn.CallWithStack(context.Background(), callStack); err != nil {
			panic(err)
		}
	}()

	res, ok := abi.wasmMemory.Read(uint32(resPtr), resLen*4)
	if !ok {
		panic(errFailedRead)
	}

	num := resLen / 3
	matches := make([]Match, 0, num)
	for i := range int(num) {
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
		if err := abi.freeFn.CallWithStack(ctx, callStack); err != nil {
			panic(err)
		}
	}

	callStack[0] = uint64(size)
	if err := abi.mallocFn.CallWithStack(ctx, callStack); err != nil {
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
	res, err := abi.mallocFn.Call(context.Background(), uint64(len(s)))
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
	if err := abi.freeFn.CallWithStack(context.Background(), callStack); err != nil {
		panic(err)
	}
}
