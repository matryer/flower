package flower

// Logger represents types capable of writing
// logs.
type Logger interface {
	New(name string) Logger
	Info(a ...interface{}) bool
	Debug(a ...interface{}) bool
	Err(a ...interface{}) bool
}

type nilLogger struct{}

func (n *nilLogger) New(name string) Logger {
	return n
}
func (n *nilLogger) Info(a ...interface{}) bool {
	return false
}
func (n *nilLogger) Debug(a ...interface{}) bool {
	return false
}
func (n *nilLogger) Err(a ...interface{}) bool {
	return false
}
