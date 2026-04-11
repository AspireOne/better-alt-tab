package input

import (
	"fmt"
	"sync/atomic"
	"syscall"
	"unsafe"

	"quick_app_switcher/internal/win32"

	coderuntime "quick_app_switcher/internal/runtime"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	procPeekMessageW = user32.NewProc("PeekMessageW")
)

const pmNoRemove = 0x0000

type Hook struct {
	target          win32.HWND
	hook            win32.Handle
	callback        uintptr
	threadID        atomic.Uint32
	altDown         bool
	tabDown         bool
	ownedSession    bool
	messageTab      uint32
	messageAltUp    uint32
	messageCancel   uint32
	started         chan error
	stopped         chan struct{}
	shutdownRequest chan struct{}
}

type keyDecision struct {
	suppress   bool
	postTab    bool
	postAltUp  bool
	postCancel bool
}

func New(target win32.HWND, messageTab, messageAltUp, messageCancel uint32) *Hook {
	return &Hook{
		target:          target,
		messageTab:      messageTab,
		messageAltUp:    messageAltUp,
		messageCancel:   messageCancel,
		started:         make(chan error, 1),
		stopped:         make(chan struct{}),
		shutdownRequest: make(chan struct{}),
	}
}

func (h *Hook) Start() error {
	go h.run()
	return <-h.started
}

func (h *Hook) Close() error {
	select {
	case <-h.shutdownRequest:
	default:
		close(h.shutdownRequest)
	}
	if threadID := h.threadID.Load(); threadID != 0 {
		win32.PostThreadMessage(threadID, win32.WM_QUIT, 0, 0)
	}
	<-h.stopped
	return nil
}

func (h *Hook) run() {
	defer close(h.stopped)
	unlock := coderuntime.LockOSThread()
	defer unlock()

	h.callback = syscall.NewCallback(h.proc)
	h.threadID.Store(win32.GetCurrentThreadID())
	ensureMessageQueue()
	hook, err := win32.SetKeyboardHook(h.callback)
	if err != nil {
		h.started <- fmt.Errorf("set keyboard hook: %w", err)
		return
	}
	h.hook = hook
	h.started <- nil

	var msg win32.MSG
	for {
		select {
		case <-h.shutdownRequest:
			_ = win32.UnhookWindowsHook(h.hook)
			return
		default:
		}
		ok, getErr := win32.GetMessage(&msg, 0, 0, 0)
		if getErr != nil || !ok {
			_ = win32.UnhookWindowsHook(h.hook)
			return
		}
		win32.TranslateMessage(&msg)
		win32.DispatchMessage(&msg)
	}
}

func (h *Hook) proc(code int32, wParam uintptr, lParam uintptr) uintptr {
	if code < 0 {
		return win32.CallNextHook(h.hook, code, wParam, lParam)
	}
	// #nosec G103 -- Low-level keyboard hooks receive a KBDLLHOOKSTRUCT pointer in lParam.
	//nolint:govet // Windows passes lParam as a live hook-struct pointer for the callback duration.
	data := *(*win32.KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
	decision := h.handleKey(data.VKCode, data.Flags)
	if decision.postTab {
		win32.PostMessage(h.target, h.messageTab, 0, 0)
	}
	if decision.postAltUp {
		win32.PostMessage(h.target, h.messageAltUp, 0, 0)
	}
	if decision.postCancel {
		win32.PostMessage(h.target, h.messageCancel, 0, 0)
	}
	if decision.suppress {
		return 1
	}
	return win32.CallNextHook(h.hook, code, wParam, lParam)
}

func (h *Hook) handleKey(vkCode, flags uint32) keyDecision {
	if flags&win32.LLKHF_INJECTED != 0 {
		return keyDecision{}
	}
	keyUp := flags&win32.LLKHF_UP != 0
	switch vkCode {
	case win32.VK_LMENU, win32.VK_RMENU, win32.VK_MENU:
		h.altDown = !keyUp
		if !keyUp {
			return keyDecision{}
		}
		h.tabDown = false
		if h.ownedSession {
			h.ownedSession = false
			return keyDecision{suppress: true, postAltUp: true}
		}
	case win32.VK_TAB:
		if !h.altDown {
			h.tabDown = false
			return keyDecision{}
		}
		if keyUp {
			h.tabDown = false
			return keyDecision{suppress: true}
		}
		if h.tabDown {
			return keyDecision{suppress: true}
		}
		h.tabDown = true
		h.ownedSession = true
		return keyDecision{suppress: true, postTab: true}
	case win32.VK_ESCAPE:
		if h.ownedSession && !keyUp {
			h.ownedSession = false
			return keyDecision{suppress: true, postCancel: true}
		}
	}
	return keyDecision{}
}

func ensureMessageQueue() {
	var msg win32.MSG
	// #nosec G103 -- PeekMessageW initializes the thread's message queue via the MSG buffer.
	_, _, _ = procPeekMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0, pmNoRemove)
}
