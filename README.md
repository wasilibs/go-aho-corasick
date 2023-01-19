# go-aho-corasick

_Note: Go applications should use [petar-dambovaliev/aho-corasick][1] instead for better performance.
Go support using [wazero][3] is provided here mostly as a performance analysis of wrapped vs rewritten libraries._

go-aho-corasick is a library for aho-corasick which wraps the Rust [BurntSushi/aho-corasick][2] library.
It provides the same API as [petar-dambovaliev/aho-corasick][1] and is a drop-in replacement with full
API and behavior compatibility.

For TinyGo applications being built for WASM, this library will perform significantly better. For Go applications,
it performs slightly worse.

## Usage

go-aho-corasick is a standard Go library package and can be added to a go.mod file. It will work fine in
Go or TinyGo projects.

```
go get github.com/wasilibs/go-aho-corasick
```

Because the library is a drop-in replacement for [petar-dambovaliev/aho-corasick][1], an import rewrite can
make migrating code to use it simple.

```go
import "github.com/peter-dambovaliev/aho-corasick"
```

can be changed to

```go
import "github.com/wasilibs/go-aho-corasick"
```

### cgo

This library also supports opting into using cgo to wrap [BurntSushi/aho-corasick][2] instead
of using WebAssembly. This requires having a built version of the library available - because
Rust libraries are not easily installed for use with `pkg-config`, environment variables like
`CGO_LDFLAGS` and `LD_LIBRARY_PATH` may be needed to find the library at build and runtime.
The build tag `aho_corasick_cgo` can be used to enable cgo support.

[1]: https://github.com/petar-dambovaliev/aho-corasick
[2]: https://github.com/BurntSushi/aho-corasick
[3]: https://wazero.io
