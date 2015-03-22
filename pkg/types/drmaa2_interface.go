package types

/* DRMAA2 types. Copied from github.com/dgruber/drmaa2.
   The DRMAA2 Go implementation is a cgo project which
   complicates compilation and makes drmaa2.h and a libdrmaa2.so
   neccessary. */

import (
	"fmt"
	"time"
	"unsafe"
)

// Job States
type JobState int

const (
	Unset JobState = iota
	Undetermined
	Queued
	QueuedHeld
	Running
	Suspended
	Requeued
	RequeuedHeld
	Done
	Failed
)

// Implements the Stringer interface
func (js JobState) String() string {
	switch js {
	case Undetermined:
		return "Undetermined"
	case Queued:
		return "Queued"
	case QueuedHeld:
		return "QueuedHeld"
	case Running:
		return "Running"
	case Suspended:
		return "Suspended"
	case Requeued:
		return "Requeued"
	case RequeuedHeld:
		return "RequeuedHeld"
	case Done:
		return "Done"
	case Failed:
		return "Failed"
	}
	return "Unset"
}

type StructType int

const (
	JobTemplateType = iota
	JobInfoType
	ReservationTemplateType
	ReservationInfoType
	QueueInfoType
	MachineInfoType
	NotificationType
)

// Extension struct which is embedded in DRMAA2 objects
// which are extensible.
type Extension struct {
	SType         StructType        // Stores the type of the struct
	Internal      unsafe.Pointer    // Enhancmement of C struct
	ExtensionList map[string]string // stores the extension requests as string
}

type JobInfo struct {
	Extension         `xml:"-" json:"-"`
	Id                string        `json:"id"`
	ExitStatus        int           `json:"exitStatus"`
	TerminatingSignal string        `json:"terminationSignal"`
	Annotation        string        `json:"annotation"`
	State             JobState      `json:"state"`
	SubState          string        `json:"subState"`
	AllocatedMachines []string      `json:"allocatedMachines"`
	SubmissionMachine string        `json:"submissionMachine"`
	JobOwner          string        `json:"jobOwner"`
	Slots             int64         `json:"slots"`
	QueueName         string        `json:"queueName"`
	WallclockTime     time.Duration `json:"wallockTime"`
	CPUTime           int64         `json:"cpuTime"`
	SubmissionTime    time.Time     `json:"submissionTime"`
	DispatchTime      time.Time     `json:"dispatchTime"`
	FinishTime        time.Time     `json:"finishTime"`
}

type JobTemplate struct {
	Extension         `xml:"-" json:"-"`
	RemoteCommand     string            `json:"remoteCommand"`
	Args              []string          `json:"args"`
	SubmitAsHold      bool              `json:"submitAsHold"`
	ReRunnable        bool              `json:"reRunnable"`
	JobEnvironment    map[string]string `json:"jobEnvironment"`
	WorkingDirectory  string            `json:"workingDirectory"`
	JobCategory       string            `json:"jobCategory"`
	Email             []string          `json:"email"`
	EmailOnStarted    bool              `json:"emailOnStarted"`
	EmailOnTerminated bool              `json:"emailOnTerminated"`
	JobName           string            `json:"jobName"`
	InputPath         string            `json:"inputPath"`
	OutputPath        string            `json:"outputPath"`
	ErrorPath         string            `json:"errorPath"`
	JoinFiles         bool              `json:"joinFiles"`
	ReservationId     string            `json:"reservationId"`
	QueueName         string            `json:"queueName"`
	MinSlots          int64             `json:"minSlots"`
	MaxSlots          int64             `json:"maxSlots"`
	Priority          int64             `json:"priority"`
	CandidateMachines []string          `json:"candidateMachines"`
	MinPhysMemory     int64             `json:"minPhysMemory"`
	MachineOs         string            `json:"machineOs"`
	MachineArch       string            `json:"machineArch"`
	StartTime         time.Time         `json:"startTime"`
	DeadlineTime      time.Time         `json:"deadlineTime"`
	StageInFiles      map[string]string `json:"stageInFiles"`
	StageOutFiles     map[string]string `json:"stageOutFiles"`
	ResourceLimits    map[string]string `json:"resourceLimits"`
	AccountingId      string            `json:"accountingString"`
}

// CPU architecture types
type CPU int

const (
	OtherCPU CPU = iota
	Alpha
	ARM
	ARM64
	Cell
	PA_RISC
	PA_RISC64
	x86
	x64
	IA_64
	MIPS
	MIPS64
	PowerPC
	PowerPC64
	SPARC
	SPARC64
)

func (cpu CPU) String() string {
	switch cpu {
	case OtherCPU:
		return "OtherCPU"
	case Alpha:
		return "Alpha"
	case ARM:
		return "ARM"
	case ARM64:
		return "ARM64"
	case Cell:
		return "Cell"
	case PA_RISC:
		return "PA_RISC"
	case PA_RISC64:
		return "PA_RISC64"
	case x86:
		return "x86"
	case x64:
		return "x64"
	case IA_64:
		return "IA_64"
	case MIPS:
		return "MIPS"
	case MIPS64:
		return "MIPS64"
	case PowerPC:
		return "PowerPC"
	case SPARC:
		return "SPARC"
	case SPARC64:
		return "SPARC64"
	}
	return "Unknown"
}

// Operating System type
type OS int

const (
	OtherOS OS = iota
	AIX
	BSD
	Linux
	HPUX
	IRIX
	MacOS
	SunOS
	TRU64
	UnixWare
	Win
	WinNT
)

// An OS struct needs to be printable.
func (os OS) String() string {
	switch os {
	case OtherOS:
		return "OtherOS"
	case AIX:
		return "AIX"
	case BSD:
		return "BSD"
	case Linux:
		return "Linux"
	case HPUX:
		return "HPUX"
	case IRIX:
		return "IRIX"
	case MacOS:
		return "MacOS"
	case SunOS:
		return "SunOS"
	case TRU64:
		return "TRU64"
	case UnixWare:
		return "UnixWare"
	case Win:
		return "Win"
	case WinNT:
		return "WinNT"
	}
	return "Unknown"
}

type Version struct {
	Major string
	Minor string
}

func (v *Version) String() string {
	return fmt.Sprintf("%s.%s", v.Major, v.Minor)
}

type Machine struct {
	Extension      `xml:"-" json:"-"`
	Name           string  `json:"name"`
	Available      bool    `json:"available"`
	Sockets        int64   `json:"sockets"`
	CoresPerSocket int64   `json:"coresPerSocket"`
	ThreadsPerCore int64   `json:"threadsPerCore"`
	Load           float64 `json:"load"`
	PhysicalMemory int64   `json:"physicalMemory"`
	VirtualMemory  int64   `json:"virtualMemory"`
	Architecture   CPU     `json:"architecture"`
	OSVersion      Version `json:"osVersion"`
	OS             OS      `json:"os"`
}

type Queue struct {
	Extension `xml:"-" json:"-"`
	Name      string `xml:"name"`
}

// Special timeout value: Don't wait
const ZeroTime = int64(0)

// Special timeout value: Wait probably infinitly
const InfiniteTime = int64(-1)

// Special time value: Time or date not set
const UnsetTime = int64(-2)
