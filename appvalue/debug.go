package appvalue

// debug is whether or not debug logging is enabled.
var debug = false

// Debug enables or disables debug logging.
//
// When enabled, slog.Debug messages will be logged.
func Debug(value bool) {
	debug = value
}
