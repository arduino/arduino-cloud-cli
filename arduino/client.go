package arduino

// Client of arduino package allows to call
// the arduino-cli commands in a programmatic way
type Client interface {
	BoardList() error
	Compile() error
}
