package stats

import (
	"errors"

	// loggers
	"github.com/sniperkit/logger"
	"github.com/sniperkit/logger/backends/gomol"
	"github.com/sniperkit/logger/backends/logrus"
	"github.com/sniperkit/logger/backends/zap"
)

const (
	defaultBackend    string = "zap"
	defaultEncoding   string = "console"
	defaultLevel      string = "debug"
	defaultOutputFile string = "./logs/activity.log"
)

var (
	// DefaultLogger is the engine used by global helper functions.
	Log                      = NewLogger(defaultBackend, defaultEncoding, defaultLevel, true, nil)
	errConfigNewLogger error = errors.New("Logger was already initialized")
)

// var LogFields logger.Fields

type Logger struct {
	Entry  logger.Logger
	Fields logger.Fields
	Config *LoggerConfig
}

type LoggerConfig struct {
	Backend       string
	Level         string
	Encoding      string
	DisableCaller bool
	InitialFields map[string]interface{}
}

func NewLogger(backend string, encoding string, level string, caller bool, fields *map[string]interface{}) *Logger {
	c := &LoggerConfig{
		Backend:       backend,
		Encoding:      encoding,
		Level:         level,
		DisableCaller: caller,
		// InitialFields: make(map[string]interface{}),
	}
	if fields != nil {
		c.InitialFields = *fields
	}
	l, err := NewLoggerWith(c)
	if err != nil {
		return nil
	}
	return l
}

func NewLoggerWith(cfg *LoggerConfig) (*Logger, error) {
	// f := logger.Fields{}

	if cfg.Backend == "" {
		cfg.Backend = defaultBackend
	}

	if cfg.Encoding == "" {
		cfg.Encoding = defaultEncoding
	}

	if cfg.Level == "" {
		cfg.Level = defaultLevel
	}

	c := &logger.Config{
		Backend:       cfg.Backend,
		Level:         cfg.Level,
		Encoding:      cfg.Encoding,
		DisableCaller: cfg.DisableCaller,
		InitialFields: cfg.InitialFields,
	}

	switch cfg.Backend {
	case "gomol":
		if l, err := gomol.New(c); err == nil {
			return &Logger{Entry: l, Config: cfg}, nil
		}
	case "logrus":
		if l, err := logrus.New(c); err == nil {
			return &Logger{Entry: l, Config: cfg}, nil
		}
	case "zap":
		if l, err := zap.New(c); err == nil {
			return &Logger{Entry: l, Config: cfg}, nil
		}
	}

	return nil, errors.New("unkown backend for logger")
}

func (l *Logger) Register(factory logger.Logger) {
	//if l == nil {
	l.Entry = factory
	//}
}

/*
func (l *Logger) Fields(fields map[string]interface{}) logger.Fields {
	//if len(fields) <= 0 {
	//	return nil
	//}
	lf := make(logger.Fields)
	for k, v := range fields {
		lf[k] = v
	}
	return lf
}
*/

// Register adds handler to the default engine.
func RegisterLogger(factory logger.Logger) {
	Log.Register(factory)
	Log.Entry.DebugWithFields(logger.Fields{
		"factory": factory != nil,
	}, "stats.RegisterLogger()")
}
