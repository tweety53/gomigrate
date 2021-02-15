package exitcode

type ExitCode int

//nolint:gochecknoglobals // because its like https://github.com/leighmcculloch/gochecknoglobals#exceptions
const (
	OK          ExitCode = 0
	Unspecified ExitCode = 1
	IoErr       ExitCode = 74
)
