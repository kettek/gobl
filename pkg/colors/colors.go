package colors

// Our name to color map.
var (
	Info    = Purple
	Notice  = Teal
	Warn    = Yellow
	Error   = Red
	Success = Green
)

// Our colors to escape codes map.
var (
	Black   = "\033[1;30m"
	Red     = "\033[1;31m"
	Green   = "\033[1;32m"
	Yellow  = "\033[1;33m"
	Purple  = "\033[1;34m"
	Magenta = "\033[1;35m"
	Teal    = "\033[1;36m"
	White   = "\033[1;37m"
	Clear   = "\033[0m"
)
