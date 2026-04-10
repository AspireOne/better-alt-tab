package win32

func GetModuleHandle() (HINSTANCE, error) {
	r, _, err := procGetModuleHandleW.Call(0)
	if r == 0 {
		return 0, err
	}
	return HINSTANCE(r), nil
}
