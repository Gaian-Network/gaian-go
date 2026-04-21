package gaian

// Logger is the interface for optional diagnostic output.
// Implement it with any logging library (slog, zap, logrus, etc.).
type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
}

// nopLogger discards all log output (default when no logger is provided).
type nopLogger struct{}

func (nopLogger) Debugf(string, ...any) {}
func (nopLogger) Infof(string, ...any)  {}
func (nopLogger) Errorf(string, ...any) {}
