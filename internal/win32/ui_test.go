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

	manager := &VirtualDesktopManager{ptr: uintptr(unsafe.Pointer(&object))}
	if got := manager.vtbl(); got != expected {
		t.Fatal("expected vtable pointer to be read from the COM object")
	}
}
