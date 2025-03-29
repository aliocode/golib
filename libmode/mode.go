package libmode

import (
	"errors"
	"strings"
	"sync"
)

var (
	appMode                  = ServiceModeProd // default application mode
	mu                       = sync.Mutex{}
	ErrWrongModeSyntax error = errors.New("received wrong mode syntax")
)

type Mode string

const (
	ServiceModeUnknown Mode = "UNKNOWN"
	ServiceModeProd    Mode = "PROD"
	ServiceModeDev     Mode = "DEV"
	ServiceModeMock    Mode = "MOCK"
	ServiceModeLocal   Mode = "LOCAL"
)

func NewModeFromString(mode string) (Mode, error) {
	upperMode := strings.ToUpper(mode)

	var newMode Mode
	switch upperMode {
	case string(ServiceModeProd):
		newMode = ServiceModeProd
	case string(ServiceModeDev):
		newMode = ServiceModeDev
	case string(ServiceModeMock):
		newMode = ServiceModeMock
	case string(ServiceModeLocal):
		newMode = ServiceModeLocal
	default:
		return ServiceModeUnknown, ErrWrongModeSyntax
	}

	mu.Lock()
	defer mu.Unlock()
	appMode = newMode
	return newMode, nil
}

func GetMode() Mode {
	mu.Lock()
	defer mu.Unlock()
	return appMode
}

// Stringer for printing
func (m Mode) String() string {
	return string(m)
}

func (m Mode) IsProd() bool {
	return m == ServiceModeProd
}

func (m Mode) IsDev() bool {
	return m == ServiceModeDev
}

func (m Mode) IsMock() bool {
	return m == ServiceModeMock
}

func (m Mode) IsLocal() bool {
	return m == ServiceModeLocal
}
