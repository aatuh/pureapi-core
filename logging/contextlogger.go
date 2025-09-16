package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DebugLogger defines debug and trace level logging.
type DebugLogger interface {
	Debug(messages ...any)
	Debugf(message string, params ...any)
	Trace(messages ...any)
	Tracef(message string, params ...any)
}

// InfoLogger defines info and warning level logging.
type InfoLogger interface {
	Info(messages ...any)
	Infof(message string, params ...any)
	Warn(messages ...any)
	Warnf(message string, params ...any)
}

// ErrorLogger defines error and fatal level logging.
type ErrorLogger interface {
	Error(messages ...any)
	Errorf(message string, params ...any)
	Fatal(messages ...any)
	Fatalf(message string, params ...any)
}

// ILogger represents a logger with different logging levels.
type ILogger interface {
	DebugLogger
	InfoLogger
	ErrorLogger
}

// LoggerFactoryFn is a function that returns a logger.
type LoggerFactoryFn func(params ...any) ILogger

// CtxLoggerFactoryFn is a function that returns a logger with context.
type CtxLoggerFactoryFn func(ctx context.Context) ILogger

// LogLevel represents the level of a log message.
type LogLevel int

// LogSeverity represents the severity of a log message.
type LogSeverity string

// LogMessage represents a log message.
type LogMessage struct {
	Severity LogSeverity `json:"severity"`
	Message  any         `json:"message"`
	Data     any         `json:"data,omitempty"`
	Extra    any         `json:"extra,omitempty"`
}

// NewLogMessage creates a new LogMessage.
//
// Parameters:
//   - severity: The severity of the message.
//   - message: The message to log.
//   - data: The data to log.
//   - extra: The extra data to log.
//
// Returns:
//   - LogMessage: The new LogMessage.
func NewLogMessage(
	severity LogSeverity, message any, data any, extra any,
) LogMessage {
	return LogMessage{
		Severity: severity,
		Message:  message,
		Data:     data,
		Extra:    extra,
	}
}

// LogLevelCfg holds logging level configuration.
type LogLevelCfg struct {
	Level    LogLevel
	Severity LogSeverity
	Color    ANSICode
	Writer   io.Writer
	Callback func(data []byte)
}

// LogLevelOpts holds logging level options.
type LogLevelOpts struct {
	Debug *LogLevelCfg
	Trace *LogLevelCfg
	Info  *LogLevelCfg
	Warn  *LogLevelCfg
	Error *LogLevelCfg
	Fatal *LogLevelCfg
}

// GetExtraDataFunc is a function that returns extra data for logging.
type GetExtraDataFunc func(ctx context.Context) *ExtraData

// LogOpts holds shared logger configuration.
type LogOpts struct {
	LoggingLevel LogLevel
	Compact      bool
	AnsiCodes    bool
	GetExtraData GetExtraDataFunc
	LogLevelOpts *LogLevelOpts
}

// ExtraData contains request metadata.
type ExtraData struct {
	Time      *time.Time `json:"time,omitempty"`
	TimeStart *time.Time `json:"time_start,omitempty"`
	TimeDelta string     `json:"time_delta,omitempty"`
	TraceID   string     `json:"trace_id,omitempty"`
	SpanID    string     `json:"span_id,omitempty"`
}

// defaultLogLevelOpts holds the package-wide fallback log level settings.
var defaultLogLevelOpts = LogLevelOpts{
	Debug: &LogLevelCfg{
		Severity: LogSeverity("DEBUG"),
		Color:    ANSICodeRed,
	},
	Trace: &LogLevelCfg{
		Level:    LogLevel(80),
		Severity: LogSeverity("TRACE"),
		Color:    ANSICodeMagenta,
	},
	Info: &LogLevelCfg{
		Level:    LogLevel(60),
		Severity: LogSeverity("INFO"),
		Color:    ANSICodeCyan,
	},
	Warn: &LogLevelCfg{
		Level:    LogLevel(40),
		Severity: LogSeverity("WARN"),
		Color:    ANSICodeOrange,
	},
	Error: &LogLevelCfg{
		Level:    LogLevel(20),
		Severity: LogSeverity("ERROR"),
		Color:    ANSICodeRed,
		Writer:   os.Stderr,
	},
	Fatal: &LogLevelCfg{
		Level:    LogLevel(0),
		Severity: LogSeverity("FATAL"),
		Color:    ANSICodeBrightYellow,
		Writer:   os.Stderr,
		Callback: func(data []byte) {
			panic(string(data))
		},
	},
}

// defaultLogOpts holds the package-wide fallback settings.
var defaultLogOpts = LogOpts{
	LoggingLevel: defaultLogLevelOpts.Info.Level,
	Compact:      false,
	AnsiCodes:    true,
	GetExtraData: func(ctx context.Context) *ExtraData { return nil },
	LogLevelOpts: &defaultLogLevelOpts,
}

// DefaultLogLevelOpts returns the package-wide fallback log level settings.
func DefaultLogLevelOpts() *LogLevelOpts {
	return &defaultLogLevelOpts
}

// DefaultLogOpts returns the package-wide fallback settings.
func DefaultLogOpts() *LogOpts {
	return &defaultLogOpts
}

// SetDefaultLogOpts overrides the the default logging configuration.
func SetDefaultLogOpts(opts LogOpts) {
	defaultLogOpts = opts
}

// LoggingLevelStrToInt converts a string to a logging level integer. The
// string is trimmed and converted to uppercase. If the string is empty,
// the default logging level is returned. If the string is a valid integer,
// it is returned as the logging level. If the string is one of the
// predefined strings, the corresponding logging level is returned.
//
// Parameters:
//   - level The string to convert.
//
// Returns:
//   - int: The logging level integer.
//   - error: An error if the conversion fails.
func LoggingLevelStrToInt(level string) (LogLevel, error) {
	val := strings.ToUpper(strings.TrimSpace(level))
	if val == "" {
		return defaultLogOpts.LoggingLevel, nil
	}
	if intVal, err := strconv.Atoi(val); err == nil {
		return LogLevel(intVal), nil
	}
	switch val {
	case string(defaultLogOpts.LogLevelOpts.Trace.Severity):
		return defaultLogOpts.LogLevelOpts.Trace.Level, nil
	case string(defaultLogOpts.LogLevelOpts.Info.Severity):
		return defaultLogOpts.LogLevelOpts.Info.Level, nil
	case string(defaultLogOpts.LogLevelOpts.Warn.Severity):
		return defaultLogOpts.LogLevelOpts.Warn.Level, nil
	case string(defaultLogOpts.LogLevelOpts.Error.Severity):
		return defaultLogOpts.LogLevelOpts.Error.Level, nil
	case string(defaultLogOpts.LogLevelOpts.Fatal.Severity):
		return defaultLogOpts.LogLevelOpts.Fatal.Level, nil
	default:
		return 0, fmt.Errorf(
			"DetermineLoggingLevel: invalid string logging level %q", level,
		)
	}
}

// CtxLogger is a logger that takes a context.
type CtxLogger struct {
	ctx  context.Context
	wg   *sync.WaitGroup
	opts LogOpts
}

// ContextLogger implements the ILogger interface.
var _ ILogger = (*CtxLogger)(nil)

// NewCtxLogger constructs a logger, using the package-level default
// options if none are passed in.
//
// Parameters:
//   - ctx The context to use.
//   - opts The optional options to use. If nil, the default options are used.
//
// Returns:
//   - *ContextLogger: The logger.
func NewCtxLogger(ctx context.Context, opts *LogOpts) *CtxLogger {
	if opts == nil {
		opts = &defaultLogOpts
	}
	return &CtxLogger{
		ctx:  ctx,
		wg:   &sync.WaitGroup{},
		opts: *opts,
	}
}

// Log prints a message with custom ANSI code and severity. It will always
// print.
//
// Parameters:
//   - ansicode The ANSI code to use.
//   - severity The severity of the message.
//   - messages The messages to print.
func (cl *CtxLogger) Log(
	ansicode ANSICode, severity LogSeverity, messages ...any,
) {
	cl.wg.Add(1)
	go func() {
		defer cl.wg.Done()
		printLnEncoded(
			ansicode,
			os.Stdout,
			cl.opts.Compact,
			cl.opts.AnsiCodes,
			createLogMessage(
				severity,
				cl.opts.GetExtraData(cl.ctx),
				messages...,
			),
		)
	}()
}

// Debug prints a debug message. It will always print.
//
// Parameters:
//   - messages The messages to print.
func (cl *CtxLogger) Debug(messages ...any) {
	printLnEncoded(
		cl.opts.LogLevelOpts.Debug.Color,
		cl.opts.LogLevelOpts.Debug.Writer,
		cl.opts.Compact,
		cl.opts.AnsiCodes,
		createLogMessage(
			cl.opts.LogLevelOpts.Debug.Severity,
			cl.opts.GetExtraData(cl.ctx),
			messages...,
		),
	)
}

// Trace prints a trace message if the logging level is high enough.
//
// Parameters:
//   - messages The messages to print.
func (cl *CtxLogger) Trace(messages ...any) {
	if cl.opts.LoggingLevel < defaultLogOpts.LogLevelOpts.Trace.Level {
		return
	}
	cl.wg.Add(1)
	go func() {
		defer cl.wg.Done()
		printLnLogLevelCfg(
			cl.opts.LogLevelOpts.Trace,
			cl.opts.Compact,
			cl.opts.AnsiCodes,
			createLogMessage(
				cl.opts.LogLevelOpts.Trace.Severity,
				cl.opts.GetExtraData(cl.ctx),
				messages...,
			),
		)
	}()
}

// Info prints an info message if the logging level is high enough.
//
// Parameters:
//   - messages The messages to print.
func (cl *CtxLogger) Info(messages ...any) {
	if cl.opts.LoggingLevel < defaultLogOpts.LogLevelOpts.Info.Level {
		return
	}
	cl.wg.Add(1)
	go func() {
		defer cl.wg.Done()
		printLnLogLevelCfg(
			cl.opts.LogLevelOpts.Info,
			cl.opts.Compact,
			cl.opts.AnsiCodes,
			createLogMessage(
				cl.opts.LogLevelOpts.Info.Severity,
				cl.opts.GetExtraData(cl.ctx),
				messages...,
			),
		)
	}()
}

// Warn prints a warning message if the logging level is high enough.
//
// Parameters:
//   - messages The messages to print.
func (cl *CtxLogger) Warn(messages ...any) {
	if cl.opts.LoggingLevel < defaultLogOpts.LogLevelOpts.Warn.Level {
		return
	}
	cl.wg.Add(1)
	go func() {
		defer cl.wg.Done()
		printLnLogLevelCfg(
			cl.opts.LogLevelOpts.Warn,
			cl.opts.Compact,
			cl.opts.AnsiCodes,
			createLogMessage(
				cl.opts.LogLevelOpts.Warn.Severity,
				cl.opts.GetExtraData(cl.ctx),
				messages...,
			),
		)
	}()
}

// Error prints an error message if the logging level is high enough.
//
// Parameters:
//   - messages The messages to print.
func (cl *CtxLogger) Error(messages ...any) {
	if cl.opts.LoggingLevel < defaultLogOpts.LogLevelOpts.Error.Level {
		return
	}
	cl.wg.Add(1)
	go func() {
		defer cl.wg.Done()
		printLnLogLevelCfg(
			cl.opts.LogLevelOpts.Error,
			cl.opts.Compact,
			cl.opts.AnsiCodes,
			createLogMessage(
				cl.opts.LogLevelOpts.Error.Severity,
				cl.opts.GetExtraData(cl.ctx),
				messages...,
			),
		)
	}()
}

// Fatal prints a fatal message if the logging level is high enough.
//
// Parameters:
//   - messages The messages to print.
func (cl *CtxLogger) Fatal(messages ...any) {
	if cl.opts.LoggingLevel < defaultLogOpts.LogLevelOpts.Fatal.Level {
		return
	}
	cl.wg.Add(1)
	go func() {
		defer cl.wg.Done()
		printLnLogLevelCfg(
			cl.opts.LogLevelOpts.Fatal,
			cl.opts.Compact,
			cl.opts.AnsiCodes,
			createLogMessage(
				cl.opts.LogLevelOpts.Fatal.Severity,
				cl.opts.GetExtraData(cl.ctx),
				messages...,
			),
		)
	}()
}

// Logf formats and prints a message. It will always print.
func (cl *CtxLogger) Logf(
	ansicode ANSICode,
	severity LogSeverity,
	format string,
	params ...any,
) {
	cl.wg.Add(1)
	go func() {
		defer cl.wg.Done()
		printLnEncoded(
			ansicode,
			os.Stdout,
			cl.opts.Compact,
			cl.opts.AnsiCodes,
			createLogMessage(
				severity,
				cl.opts.GetExtraData(cl.ctx),
				fmt.Sprintf(format, params...),
			),
		)
	}()
}

// Debugf formats and prints a debug message. It will always print.
//
// Parameters:
//   - format The format string.
//   - params The parameters to format.
func (cl *CtxLogger) Debugf(format string, params ...any) {
	cl.wg.Add(1)
	go func() {
		defer cl.wg.Done()
		printLnLogLevelCfg(
			cl.opts.LogLevelOpts.Debug,
			cl.opts.Compact,
			cl.opts.AnsiCodes,
			createLogMessage(
				cl.opts.LogLevelOpts.Debug.Severity,
				cl.opts.GetExtraData(cl.ctx),
				fmt.Sprintf(format, params...),
			),
		)
	}()
}

// Tracef formats and prints a trace message if the logging level is high
// enough.
//
// Parameters:
//   - format The format string.
//   - params The parameters to format.
func (cl *CtxLogger) Tracef(format string, params ...any) {
	if cl.opts.LoggingLevel < defaultLogOpts.LogLevelOpts.Trace.Level {
		return
	}
	cl.wg.Add(1)
	go func() {
		defer cl.wg.Done()
		printLnLogLevelCfg(
			cl.opts.LogLevelOpts.Trace,
			cl.opts.Compact,
			cl.opts.AnsiCodes,
			createLogMessage(
				cl.opts.LogLevelOpts.Trace.Severity,
				cl.opts.GetExtraData(cl.ctx),
				fmt.Sprintf(format, params...),
			),
		)
	}()
}

// Infof formats and prints an info message if the logging level is high
// enough.
//
// Parameters:
//   - format The format string.
//   - params The parameters to format.
func (cl *CtxLogger) Infof(format string, params ...any) {
	if cl.opts.LoggingLevel < defaultLogOpts.LogLevelOpts.Info.Level {
		return
	}
	cl.wg.Add(1)
	go func() {
		defer cl.wg.Done()
		printLnLogLevelCfg(
			cl.opts.LogLevelOpts.Info,
			cl.opts.Compact,
			cl.opts.AnsiCodes,
			createLogMessage(
				cl.opts.LogLevelOpts.Info.Severity,
				cl.opts.GetExtraData(cl.ctx),
				fmt.Sprintf(format, params...),
			),
		)
	}()
}

// Warnf formats and prints a warn message if the logging level is high
// enough.
//
// Parameters:
//   - format The format string.
//   - params The parameters to format.
func (cl *CtxLogger) Warnf(format string, params ...any) {
	if cl.opts.LoggingLevel < defaultLogOpts.LogLevelOpts.Warn.Level {
		return
	}
	cl.wg.Add(1)
	go func() {
		defer cl.wg.Done()
		printLnLogLevelCfg(
			cl.opts.LogLevelOpts.Warn,
			cl.opts.Compact,
			cl.opts.AnsiCodes,
			createLogMessage(
				cl.opts.LogLevelOpts.Warn.Severity,
				cl.opts.GetExtraData(cl.ctx),
				fmt.Sprintf(format, params...),
			),
		)
	}()
}

// Errorf formats and prints an error message if the logging level is high
// enough.
//
// Parameters:
//   - format The format string.
//   - params The parameters to format.
func (cl *CtxLogger) Errorf(format string, params ...any) {
	if cl.opts.LoggingLevel < defaultLogOpts.LogLevelOpts.Error.Level {
		return
	}
	cl.wg.Add(1)
	go func() {
		defer cl.wg.Done()
		printLnLogLevelCfg(
			cl.opts.LogLevelOpts.Error,
			cl.opts.Compact,
			cl.opts.AnsiCodes,
			createLogMessage(
				cl.opts.LogLevelOpts.Error.Severity,
				cl.opts.GetExtraData(cl.ctx),
				fmt.Sprintf(format, params...),
			),
		)
	}()
}

// Fatalf formats and prints a fatal message if the logging level is high
// enough.
//
// Parameters:
//   - format The format string.
//   - params The parameters to format.
func (cl *CtxLogger) Fatalf(format string, params ...any) {
	if cl.opts.LoggingLevel < defaultLogOpts.LogLevelOpts.Fatal.Level {
		return
	}
	cl.wg.Add(1)
	go func() {
		defer cl.wg.Done()
		printLnLogLevelCfg(
			cl.opts.LogLevelOpts.Fatal,
			cl.opts.Compact,
			cl.opts.AnsiCodes,
			createLogMessage(
				cl.opts.LogLevelOpts.Fatal.Severity,
				cl.opts.GetExtraData(cl.ctx),
				fmt.Sprintf(format, params...),
			),
		)
	}()
}

// createLogMessage returns a LogMessage with the given severity and message.
func createLogMessage(
	severity LogSeverity, extra *ExtraData, messages ...any,
) LogMessage {
	if len(messages) == 0 {
		return NewLogMessage(severity, nil, nil, extra)
	}
	return NewLogMessage(
		severity,
		messages[0],
		func() any {
			if len(messages) > 1 {
				return messages[1:]
			}
			return nil
		}(),
		extra,
	)
}

// printLnEncoded formats and prints a message.
func printLnLogLevelCfg(
	logLevelCfg *LogLevelCfg,
	compactLogging bool,
	allowANSICodes bool,
	logData any,
) string {
	data := printLnEncoded(
		logLevelCfg.Color,
		logLevelCfg.Writer,
		compactLogging,
		allowANSICodes,
		logData,
	)
	// Callback if set.
	if logLevelCfg.Callback != nil {
		logLevelCfg.Callback(data)
	}
	return string(data)
}

// printLnEncoded formats and prints a message.
func printLnEncoded(
	ansicode ANSICode,
	logWriter io.Writer,
	compactLogging bool,
	allowANSICodes bool,
	logData any,
) []byte {
	var logBytes []byte
	switch v := logData.(type) {
	case string:
		logBytes = []byte(v)
	case []byte:
		logBytes = []byte(v)
	case bool:
		if v {
			logBytes = []byte("true")
		} else {
			logBytes = []byte("false")
		}
	default:
		var err error
		logBytes, err = encodeJSONLogMessage(compactLogging, logData)
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"log message encoding error: %s",
				err.Error(),
			)
			return nil
		}

	}
	// Print the message.
	if logWriter == nil {
		logWriter = os.Stdout
	}
	printFmt(ansicode, allowANSICodes, logWriter, logBytes)
	return logBytes
}

// printFmt formats and prints a message.
func printFmt(
	ansicode ANSICode,
	allowANSICodes bool,
	logWriter io.Writer,
	jsonBytes []byte,
) {
	if allowANSICodes {
		fmt.Fprintf(logWriter, "%s%s%s\n", ansicode, jsonBytes, ANSICodeReset)
	} else {
		fmt.Fprintf(logWriter, "%s\n", jsonBytes)
	}
}

// encodeJSONLogMessage encodes a log message to JSON.
func encodeJSONLogMessage(compactLogging bool, logData any) ([]byte, error) {
	if compactLogging {
		return json.Marshal(logData)
	} else {
		return json.MarshalIndent(logData, "", " ")
	}
}
