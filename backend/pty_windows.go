package main

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"unsafe"
)

var (
	procCreatePseudoConsole           = kernel32.NewProc("CreatePseudoConsole")
	procClosePseudoConsole            = kernel32.NewProc("ClosePseudoConsole")
	procResizePseudoConsole           = kernel32.NewProc("ResizePseudoConsole")
	procInitializeProcThreadAttributeList = kernel32.NewProc("InitializeProcThreadAttributeList")
	procUpdateProcThreadAttribute     = kernel32.NewProc("UpdateProcThreadAttribute")
	procDeleteProcThreadAttributeList = kernel32.NewProc("DeleteProcThreadAttributeList")
)

const (
	PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE = 0x00020016
	EXTENDED_STARTUPINFO_PRESENT        = 0x00080000
)

type HPCON syscall.Handle

type COORD struct {
	X int16
	Y int16
}

type STARTUPINFOEX struct {
	StartupInfo syscall.StartupInfo
	AttributeList uintptr
}

type PTY struct {
	hPC      HPCON
	inPipe   *os.File
	outPipe  *os.File
	process  *os.Process
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
		return nil, fmt.Errorf("failed to create pseudo console: %v", err)
	}

	pty := &PTY{
		hPC:     hPC,
		inPipe:  pwIn,
		outPipe: prOut,
	}

	prIn.Close()
	pwOut.Close()

	// Prepare Attribute List for CreateProcess
	var sizeList uintptr
	procInitializeProcThreadAttributeList.Call(0, 1, 0, uintptr(unsafe.Pointer(&sizeList)))
	
	attrList := make([]byte, sizeList)
	procInitializeProcThreadAttributeList.Call(uintptr(unsafe.Pointer(&attrList[0])), 1, 0, uintptr(unsafe.Pointer(&sizeList)))
	defer procDeleteProcThreadAttributeList.Call(uintptr(unsafe.Pointer(&attrList[0])))

	procUpdateProcThreadAttribute.Call(
		uintptr(unsafe.Pointer(&attrList[0])),
		0,
		PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE,
		uintptr(hPC),
		uintptr(unsafe.Sizeof(hPC)),
		0,
		0,
	)

	siEx := STARTUPINFOEX{
		StartupInfo: syscall.StartupInfo{
			Cb: uint32(unsafe.Sizeof(STARTUPINFOEX{})),
		},
		AttributeList: uintptr(unsafe.Pointer(&attrList[0])),
	}

	pi := syscall.ProcessInformation{}
	cmdPtr, _ := syscall.UTF16PtrFromString(command)

	err = syscall.CreateProcess(
		nil,
		cmdPtr,
		nil,
		nil,
		false,
		EXTENDED_STARTUPINFO_PRESENT,
		nil,
		nil,
		(*syscall.StartupInfo)(unsafe.Pointer(&siEx)),
		&pi,
	)
	if err != nil {
		pty.Close()
		return nil, fmt.Errorf("failed to create process: %v", err)
	}

	// Wait, we should also assign to Job Object here!
	// But we'll do it in the session manager since we have the JobObject there.
	
	pty.process, _ = os.FindProcess(int(pi.ProcessId))
	
	syscall.CloseHandle(pi.Thread)
	syscall.CloseHandle(pi.Process)

	return pty, nil
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
		return err
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
	if p.process != nil {
		p.process.Kill()
	}
	return nil
}
