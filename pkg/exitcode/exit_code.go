package exitcode

type ExitCode int

//nolint:gochecknoglobals // because its like https://github.com/leighmcculloch/gochecknoglobals#exceptions
var (
	ExitCodeOK          ExitCode = 0
	ExitCodeUnspecified ExitCode = 1
	ExitCodeIOErr       ExitCode = 74
)
