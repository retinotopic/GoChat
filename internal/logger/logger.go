package logger

type Logger interface {
	Error(string, error)
}
