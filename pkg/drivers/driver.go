package drivers

import "io"

// Driver is the interface that must be implemented by a driver.
type Driver interface {
	LoadEnv(string) error
	LoadFlags() error
	Init() error
	GetWork() (io.Reader, error)
	ClearWork() error
	HandleFailure() error
	Cleanup() error
}
