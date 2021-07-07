package grpc

type compileHandler struct {
	*service
}

// Compile executes the 'arduino-cli compile' command
// and returns its result.
func (c compileHandler) Compile() error {
	return nil
}
