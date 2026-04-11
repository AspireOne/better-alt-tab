package win32

import (
	"testing"
	"unsafe"
)

func TestVirtualDesktopManagerVtblReadsObjectVTablePointer(t *testing.T) {
	expected := &virtualDesktopManagerVtbl{IsCurrent: 123}
	object := struct {
		vtbl *virtualDesktopManagerVtbl
	}{vtbl: expected}

	// #nosec G103 -- Test fixture intentionally models the COM object memory layout.
	manager := &VirtualDesktopManager{ptr: unsafe.Pointer(&object)}
	if got := manager.vtbl(); got != expected {
		t.Fatal("expected vtable pointer to be read from the COM object")
	}
}
