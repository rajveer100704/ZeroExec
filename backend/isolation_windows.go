package main

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	kernel32                      = syscall.NewLazyDLL("kernel32.dll")
	procCreateJobObject           = kernel32.NewProc("CreateJobObjectW")
	procSetInformationJobObject   = kernel32.NewProc("SetInformationJobObject")
	procAssignProcessToJobObject  = kernel32.NewProc("AssignProcessToJobObject")
)

const (
	JobObjectExtendedLimitInformation = 9
	JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE = 0x00002000
)

type JOBOBJECT_BASIC_LIMIT_INFORMATION struct {
	PerProcessUserTimeLimit uint64
	PerJobUserTimeLimit     uint64
	LimitFlags              uint32
	MinimumWorkingSetSize   uintptr
	MaximumWorkingSetSize   uintptr
	ActiveProcessLimit      uint32
	Affinity                uintptr
	PriorityClass           uint32
	SchedulingClass         uint32
}

type IO_COUNTERS struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

type JOBOBJECT_EXTENDED_LIMIT_INFORMATION struct {
	BasicLimitInformation JOBOBJECT_BASIC_LIMIT_INFORMATION
	IoCounters             IO_COUNTERS
	ProcessMemoryLimit    uintptr
	JobMemoryLimit        uintptr
	PeakProcessMemoryLimit uintptr
	PeakJobMemoryLimit    uintptr
}

type JobObject struct {
	handle syscall.Handle
}

func CreateJobObject(name string) (*JobObject, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("job objects are only supported on Windows")
	}

	var namePtr *uint16
	if name != "" {
		var err error
		namePtr, err = syscall.UTF16PtrFromString(name)
		if err != nil {
			return nil, err
		}
	}

	handle, _, err := procCreateJobObject.Call(0, uintptr(unsafe.Pointer(namePtr)))
	if handle == 0 {
		return nil, err
	}

	jo := &JobObject{handle: syscall.Handle(handle)}

	// Set limit: kill on close
	info := JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	info.BasicLimitInformation.LimitFlags = JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE

	ret, _, err := procSetInformationJobObject.Call(
		uintptr(jo.handle),
		JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uintptr(unsafe.Sizeof(info)),
	)
	if ret == 0 {
		jo.Close()
		return nil, err
	}

	return jo, nil
}

func (jo *JobObject) AssignProcess(pid int) error {
	const PROCESS_SET_QUOTA = 0x0100
	hProcess, err := syscall.OpenProcess(PROCESS_SET_QUOTA|syscall.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(hProcess)

	ret, _, err := procAssignProcessToJobObject.Call(uintptr(jo.handle), uintptr(hProcess))
	if ret == 0 {
		return err
	}

	return nil
}

func (jo *JobObject) Close() error {
	return syscall.CloseHandle(jo.handle)
}
