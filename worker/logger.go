package main

// NullLogger : Null logger for debug output
type NullLogger struct {
}

// NewNullLogger : Create Nulllogger
func NewNullLogger() *NullLogger {
	var newNullLogger NullLogger
	return &newNullLogger
}

func (l *NullLogger) Print(v ...interface{}) {
}
func (l *NullLogger) Printf(format string, v ...interface{}) {
}
func (l *NullLogger) Println(v ...interface{}) {
}
func (l *NullLogger) Fatal(v ...interface{}) {
}
func (l *NullLogger) Fatalf(format string, v ...interface{}) {
}
func (l *NullLogger) Fatalln(v ...interface{}) {
}
func (l *NullLogger) Panic(v ...interface{}) {
}
func (l *NullLogger) Panicf(format string, v ...interface{}) {
}
func (l *NullLogger) Panicln(v ...interface{}) {
}
