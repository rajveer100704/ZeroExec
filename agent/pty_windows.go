package agent

import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	procCreatePseudoConsole = kernel32.NewProc("CreatePseudoConsole")
	procClosePseudoConsole  = kernel32.NewProc("ClosePseudoConsole")
	procResizePseudoConsole = kernel32.NewProc("ResizePseudoConsole")
)

type HPCON windows.Handle

type COORD struct {
	X int16
	Y int16
}

type PTY struct {
	hPC      HPCON
	inPipe   *os.File
	outPipe  *os.File
	process  windows.Handle
	pid      uint32
	mu       sync.Mutex
	closed   bool
}

func StartPTY(command string, rows, cols uint16) (*PTY, error) {
	prIn, pwIn, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	prOut, pwOut, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	size := COORD{X: int16(cols), Y: int16(rows)}
	var hPC HPCON

	ret, _, err := procCreatePseudoConsole.Call(
		uintptr(unsafe.Pointer(&size)),
		uintptr(prIn.Fd()),
		uintptr(pwOut.Fd()),
		0,
		uintptr(unsafe.Pointer(&hPC)),
	)
	if ret != 0 {
		return nil, fmt.Errorf("CreatePseudoConsole failed (HRESULT: 0x%X): %v", ret, err)
	}

	pty := &PTY{
		hPC:     hPC,
		inPipe:  pwIn,
		outPipe: prOut,
	}

	prIn.Close()
	pwOut.Close()

	// Launch process with attribute list
	attrList, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		pty.Close()
		return nil, err
	}
	defer attrList.Delete()

	err = attrList.Update(
		windows.PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE,
		unsafe.Pointer(uintptr(hPC)),
		unsafe.Sizeof(hPC),
	)
	if err != nil {
		pty.Close()
		return nil, err
	}

	si := windows.StartupInfoEx{
		StartupInfo: windows.StartupInfo{
			Cb: uint32(unsafe.Sizeof(windows.StartupInfoEx{})),
		},
	}
	si.ProcThreadAttributeList = attrList.List()

	pi := windows.ProcessInformation{}
	cmdPtr, _ := windows.UTF16PtrFromString(command)

	err = windows.CreateProcess(
		nil,
		cmdPtr,
		nil,
		nil,
		false,
		windows.EXTENDED_STARTUPINFO_PRESENT,
		nil,
		nil,
		&si.StartupInfo,
		&pi,
	)
	if err != nil {
		pty.Close()
		return nil, fmt.Errorf("CreateProcess failed: %v", err)
	}

	pty.process = pi.Process
	pty.pid = pi.ProcessId
	windows.CloseHandle(pi.Thread)

	return pty, nil
}

func (p *PTY) PID() uint32 {
	return p.pid
}

func (p *PTY) Write(b []byte) (int, error) {
	return p.inPipe.Write(b)
}

func (p *PTY) Read(b []byte) (int, error) {
	return p.outPipe.Read(b)
}

func (p *PTY) Resize(rows, cols uint16) error {
	size := COORD{X: int16(cols), Y: int16(rows)}
	ret, _, err := procResizePseudoConsole.Call(uintptr(p.hPC), uintptr(unsafe.Pointer(&size)))
	if ret != 0 {
		return fmt.Errorf("ResizePseudoConsole failed (HRESULT: 0x%X): %v", ret, err)
	}
	return nil
}

func (p *PTY) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil
	}
	p.closed = true
	procClosePseudoConsole.Call(uintptr(p.hPC))
	p.inPipe.Close()
	p.outPipe.Close()
	if p.process != 0 {
		windows.TerminateProcess(p.process, 0)
		windows.CloseHandle(p.process)
	}
	return nil
}
