package events

import (
	"sync/atomic"
	"syscall"

	"quick_app_switcher/internal/win32"
)

type ForegroundWatcher struct {
	hook      win32.Handle
	target    win32.HWND
	callback  uintptr
	lastHWND  atomic.Uintptr
	messageID uint32
}

func NewForegroundWatcher(target win32.HWND, messageID uint32) (*ForegroundWatcher, error) {
	w := &ForegroundWatcher{target: target, messageID: messageID}
	w.callback = syscall.NewCallback(w.handle)
	hook, err := win32.SetForegroundEventHook(w.callback)
	if err != nil {
		return nil, err
	}
	w.hook = hook
	return w, nil
}

func (w *ForegroundWatcher) Close() error {
	if w == nil || w.hook == 0 {
		return nil
	}
	err := win32.UnhookWinEvent(w.hook)
	w.hook = 0
	return err
}

func (w *ForegroundWatcher) handle(_ win32.Handle, _ uint32, hwnd win32.HWND, _ int32, _ int32, _ uint32, _ uint32) uintptr {
	if hwnd == 0 {
		return 0
	}
	last := w.lastHWND.Swap(uintptr(hwnd))
	if last == uintptr(hwnd) {
		return 0
	}
	win32.PostMessage(w.target, w.messageID, uintptr(hwnd), 0)
	return 0
}
