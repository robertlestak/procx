package drivers

// Driver is the interface that must be implemented by a driver.
type Driver interface {
	LoadEnv(string) error
	LoadFlags() error
	Init() error
	GetWork() (*string, error)
	ClearWork() error
	HandleFailure() error
}
