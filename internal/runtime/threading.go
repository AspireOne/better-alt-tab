package runtime

import "runtime"

func LockOSThread() func() {
	runtime.LockOSThread()
	return runtime.UnlockOSThread
}
