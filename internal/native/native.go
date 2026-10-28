package native

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/Sn0wo2/caelum/internal/loader"
	"github.com/ebitengine/purego"
)

// ABIVersion is the acta FFI ABI this binding implements.
const ABIVersion = 2

type API struct {
	ABIVersion func() uint32
	Create     func(unsafe.Pointer, uintptr, uintptr, unsafe.Pointer) int32
	Log        func(uint64, uint32, unsafe.Pointer, uintptr) int32
	Event      func(uint64, unsafe.Pointer, uintptr) int32
	CreateSpan func(uint64, unsafe.Pointer, uintptr, unsafe.Pointer) int32
	RecordSpan func(uint64, uint32, unsafe.Pointer, uintptr) int32
	CloseSpan  func(uint64, uint32) int32
	Drain      func(uint64, unsafe.Pointer, uintptr, unsafe.Pointer, unsafe.Pointer) int32
	SetFilter  func(uint64, unsafe.Pointer, uintptr) int32
	Destroy    func(uint64) int32

	version   func() unsafe.Pointer
	lastError func() unsafe.Pointer
}

// cString reads a NUL-terminated string at pointer, or "" when null.
func cString(pointer unsafe.Pointer) string {
	if pointer == nil {
		return ""
	}

	length := 0
	for *(*byte)(unsafe.Add(pointer, length)) != 0 {
		length++
	}

	return unsafe.String((*byte)(pointer), length)
}

// ActaVersion returns the acta version string reported by the loaded library.
func (api *API) ActaVersion() string {
	return cString(api.version())
}

// LastError returns the detail message acta recorded for the most recent
// failed call on this thread, or "" when the last call succeeded. It must be
// read before any other acta call, which may overwrite the message.
func (api *API) LastError() string {
	return cString(api.lastError())
}

var embeddedAPI = sync.OnceValues(func() (*API, error) {
	filename := "libacta_ffi.so"

	switch runtime.GOOS {
	case "windows":
		filename = "acta_ffi.dll"
	case "darwin":
		filename = "libacta_ffi.dylib"
	}

	path, err := loader.Extract(embeddedNativeAssets, runtime.GOOS+"_"+runtime.GOARCH+"/"+filename)
	if err != nil {
		return nil, fmt.Errorf("caelum: load embedded acta library: %w", err)
	}

	library, err := loader.Open(path)
	if err != nil {
		return nil, fmt.Errorf("caelum: open embedded acta library: %w", err)
	}

	return bind(library)
})

func Load(library loader.Loader) (*API, error) {
	if library == nil {
		return embeddedAPI()
	}

	return bind(library)
}

func bind(library loader.Loader) (*API, error) {
	api := &API{}

	for _, binding := range []struct {
		name string
		fn   any
	}{
		{"acta_abi_version", &api.ABIVersion},
		{"acta_version", &api.version},
		{"acta_last_error", &api.lastError},
		{"acta_logger_create", &api.Create},
		{"acta_logger_log", &api.Log},
		{"acta_logger_event", &api.Event},
		{"acta_logger_create_span", &api.CreateSpan},
		{"acta_logger_record_span", &api.RecordSpan},
		{"acta_logger_close_span", &api.CloseSpan},
		{"acta_logger_drain", &api.Drain},
		{"acta_logger_set_filter", &api.SetFilter},
		{"acta_logger_destroy", &api.Destroy},
	} {
		address, err := library.Lookup(binding.name)
		if err != nil {
			return nil, fmt.Errorf("caelum: resolve %s: %w", binding.name, err)
		}

		purego.RegisterFunc(binding.fn, address)
	}

	if version := api.ABIVersion(); version != ABIVersion {
		return nil, fmt.Errorf("caelum: unsupported acta ABI version %d (want %d)", version, ABIVersion)
	}

	return api, nil
}
