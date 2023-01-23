//go:build !tinygo.wasm && !aho_corasick_cgo

package aho_corasick

import (
	"context"
	_ "embed"
	"errors"
	"strconv"
	"sync"
	"sync/atomic"

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

	malloc api.Function
	free   api.Function

	wasmMemory api.Memory

	mod api.Module

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

var moduleIdx = uint64(0)

func newABI() *ahoCorasickABI {
	ctx := context.Background()
	modIdx := atomic.AddUint64(&moduleIdx, 1)
	mod, err := wasmRT.InstantiateModule(ctx, wasmCompiled, wazero.NewModuleConfig().WithName(strconv.FormatUint(modIdx, 10)))
	if err != nil {
		panic(err)
	}

	return &ahoCorasickABI{
		new_matcher:             mod.ExportedFunction("new_matcher"),
		delete_matcher:          mod.ExportedFunction("delete_matcher"),
		find_iter:               mod.ExportedFunction("find_iter"),
		find_iter_next:          mod.ExportedFunction("find_iter_next"),
		find_iter_delete:        mod.ExportedFunction("find_iter_delete"),
		overlapping_iter:        mod.ExportedFunction("overlapping_iter"),
		overlapping_iter_next:   mod.ExportedFunction("overlapping_iter_next"),
		overlapping_iter_delete: mod.ExportedFunction("overlapping_iter_delete"),

		malloc: mod.ExportedFunction("malloc"),
		free:   mod.ExportedFunction("free"),

		wasmMemory: mod.Memory(),
		mod:        mod,
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

	res, err := abi.new_matcher.Call(context.Background(), uint64(patternsPtr), uint64(lensPtr), uint64(len(patterns)), uint64(aci), uint64(d), uint64(matchKind))
	if err != nil {
		panic(err)
	}

	return uintptr(res[0])
}

func (abi *ahoCorasickABI) deleteMatcher(ptr uintptr) {
	_, err := abi.delete_matcher.Call(context.Background(), uint64(ptr))
	if err != nil {
		panic(err)
	}
}

func (abi *ahoCorasickABI) findIter(acPtr uintptr, value cString) uintptr {
	res, err := abi.find_iter.Call(context.Background(), uint64(acPtr), uint64(value.ptr), uint64(value.length))
	if err != nil {
		panic(err)
	}

	return uintptr(res[0])
}

func (abi *ahoCorasickABI) findIterNext(iter uintptr) (int, int, int, bool) {
	patternPtr := abi.memory.allocate(4)
	startPtr := abi.memory.allocate(4)
	endPtr := abi.memory.allocate(4)

	res, err := abi.find_iter_next.Call(context.Background(), uint64(iter), uint64(patternPtr), uint64(startPtr), uint64(endPtr))
	if err != nil {
		panic(err)
	}

	if res[0] == 0 {
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
	_, err := abi.find_iter_delete.Call(context.Background(), uint64(iter))
	if err != nil {
		panic(err)
	}
}

func (abi *ahoCorasickABI) overlappingIter(acPtr uintptr, value cString) uintptr {
	res, err := abi.overlapping_iter.Call(context.Background(), uint64(acPtr), uint64(value.ptr), uint64(value.length))
	if err != nil {
		panic(err)
	}

	return uintptr(res[0])
}

func (abi *ahoCorasickABI) overlappingIterNext(iter uintptr) (int, int, int, bool) {
	patternPtr := abi.memory.allocate(4)
	startPtr := abi.memory.allocate(4)
	endPtr := abi.memory.allocate(4)

	res, err := abi.overlapping_iter_next.Call(context.Background(), uint64(iter), uint64(patternPtr), uint64(startPtr), uint64(endPtr))
	if err != nil {
		panic(err)
	}

	if res[0] == 0 {
		return 0, 0, 0, false
	}

	pattern, _ := abi.wasmMemory.ReadUint32Le(uint32(patternPtr))
	start, _ := abi.wasmMemory.ReadUint32Le(uint32(startPtr))
	end, _ := abi.wasmMemory.ReadUint32Le(uint32(endPtr))

	return int(pattern), int(start), int(end), true
}

func (abi *ahoCorasickABI) overlappingIterDelete(iter uintptr) {
	_, err := abi.overlapping_iter_delete.Call(context.Background(), uint64(iter))
	if err != nil {
		panic(err)
	}
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
	if m.bufPtr != 0 {
		_, err := abi.free.Call(ctx, uint64(m.bufPtr))
		if err != nil {
			panic(err)
		}
	}

	res, err := abi.malloc.Call(ctx, uint64(size))
	if err != nil {
		panic(err)
	}

	m.size = size
	m.bufPtr = uint32(res[0])
}

func (m *sharedMemory) allocate(size uint32) uintptr {
	if m.nextIdx+size > m.size {
		panic("not enough reserved shared memory")
	}

	ptr := m.bufPtr + m.nextIdx
	m.nextIdx += size
	return uintptr(ptr)
}

func (m *sharedMemory) write(abi *ahoCorasickABI, b []byte) uintptr {
	ptr := m.allocate(uint32(len(b)))
	abi.wasmMemory.Write(uint32(ptr), b)
	return ptr
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
	_, err := abi.free.Call(context.Background(), uint64(ptr))
	if err != nil {
		panic(err)
	}
}
