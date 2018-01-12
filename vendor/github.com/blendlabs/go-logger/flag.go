package logger

const (
	// FlagAll is a special flag that allows all events to fire.
	FlagAll Flag = "all"
	// FlagNone is a special flag that allows no events to fire.
	FlagNone Flag = "none"

	// Fatal fires for fatal errors and is an alias to `Fatal`.
	Fatal Flag = "fatal"
	// Error fires for errors that are severe enough to log but not so severe as to abort a process.
	Error Flag = "error"
	// Warning fires for warnings.
	Warning Flag = "warning"
	// Debug fires for debug messages.
	Debug Flag = "debug"
	// Info fires for informational messages (app startup etc.)
	Info Flag = "info"
	// Silly is for when you just need to log something weird.
	Silly Flag = "silly"

	// WebRequestStart is an event flag.
	WebRequestStart Flag = "web.request.start"
	// WebRequest is an event flag.
	WebRequest Flag = "web.request"
)

// Flag represents an event type that can be enabled or disabled.
type Flag string
