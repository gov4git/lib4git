package base

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var verbose bool

func init() {
	LogQuietly()
}

func IsVerbose() bool {
	return verbose
}

func LogQuietly() {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:        os.Stderr,
			NoColor:    true,
			TimeFormat: time.RFC3339,
		},
	)
	verbose = false
}

func LogVerbosely() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:        os.Stderr,
			NoColor:    true,
			TimeFormat: time.RFC3339,
		},
	)
	verbose = true
}

func AssertNoErr(err error) {
	if err == nil {
		return
	}
	Fatalf("encountered %v", err)
}

func Debugf(template string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	log.Debug().Msg(src + msg)
}

func Infof(template string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	log.Info().Msg(src + msg)
}

func Fatalf(template string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	log.Fatal().Msg(src + msg)
}

func Errorf(template string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	src := fmt.Sprintf("%s:%d ", file, line)
	msg := fmt.Sprintf(template, args...)
	log.Error().Msg(src + msg)
}

func Sync() error {
	return nil
}
