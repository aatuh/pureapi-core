package logging

// ANSICode represents a color and formatting code.
type ANSICode string

// ANSICodes for colored and formatted output.
var (
	ANSICodeCyan         = ANSICode("\033[36m")
	ANSICodeLightGreen   = ANSICode("\033[92m")
	ANSICodeOrange       = ANSICode("\033[33m")
	ANSICodeRed          = ANSICode("\033[31m")
	ANSICodeLightBlue    = ANSICode("\033[94m")
	ANSICodeGray         = ANSICode("\033[90m")
	ANSICodeBrightYellow = ANSICode("\033[93m")
	ANSICodeMagenta      = ANSICode("\033[35m")
	ANSICodeReset        = ANSICode("\033[0m")
)
