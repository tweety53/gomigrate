package exit_code

type ExitCode int

var (
	ExitCodeOK          ExitCode = 0
	ExitCodeUnspecified ExitCode = 1
	ExitCodeIOErr       ExitCode = 74
)
