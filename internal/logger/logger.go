package logger

type Logger interface {
	Error(string, error)
	Fatal(string, error)
}
