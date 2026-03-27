package agent

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
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
	handle windows.Handle
}

func CreateJobObject(name string) (*JobObject, error) {
	var namePtr *uint16
	if name != "" {
		var err error
		namePtr, err = windows.UTF16PtrFromString(name)
		if err != nil {
			return nil, err
		}
	}

	handle, err := windows.CreateJobObject(nil, namePtr)
	if err != nil {
		return nil, fmt.Errorf("CreateJobObject failed: %v", err)
	}

	jo := &JobObject{handle: handle}

	// Set limit: kill on close
	info := JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	info.BasicLimitInformation.LimitFlags = JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE

	_, err = windows.SetInformationJobObject(
		jo.handle,
		JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	)
	if err != nil {
		jo.Close()
		return nil, fmt.Errorf("SetInformationJobObject failed: %v", err)
	}

	return jo, nil
}

func (jo *JobObject) AssignProcess(hProcess windows.Handle) error {
	err := windows.AssignProcessToJobObject(jo.handle, hProcess)
	if err != nil {
		return fmt.Errorf("AssignProcessToJobObject failed: %v", err)
	}
	return nil
}

func (jo *JobObject) Close() error {
	return windows.CloseHandle(jo.handle)
}
