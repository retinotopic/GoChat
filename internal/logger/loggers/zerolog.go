package zerolog

import (
	"io"

	"github.com/rs/zerolog"
)

type ZerologLogger struct {
	logger *zerolog.Logger
}

func NewZerologLogger(writer io.Writer) *ZerologLogger {
	output := zerolog.ConsoleWriter{Out: writer, TimeFormat: "15:04:05"}
	logger := zerolog.New(output).With().Timestamp().Logger()

	return &ZerologLogger{
		logger: &logger,
	}
}

func (l *ZerologLogger) Error(msg string, err error) {
	l.logger.Error().Err(err).Msg(msg)
}
