package logx

/*
创建logger对象，实现trace等功能
*/

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/sysx"
	"github.com/zeromicro/go-zero/core/timex"
)

type fullLogger struct {
	logEntry
	ctx context.Context

	//
	timeFormat   string
	writeConsole bool
	logLevel     uint32
	encoding     uint32

	//
	infoLog   io.WriteCloser
	errorLog  io.WriteCloser
	severeLog io.WriteCloser
	slowLog   io.WriteCloser

	//
	initialized uint32
	options     logOptions
}

func NewFullLogger(c logx.LogConf) FullLogger {
	l := &fullLogger{
		timeFormat: "2006-01-02T15:04:05.000Z07:00",
		encoding:   jsonEncodingType,
	}

	if len(c.TimeFormat) > 0 {
		l.timeFormat = c.TimeFormat
	}
	switch c.Encoding {
	case plainEncoding:
		atomic.StoreUint32(&(l.encoding), plainEncodingType)
	default:
		atomic.StoreUint32(&(l.encoding), jsonEncodingType)
	}

	switch c.Mode {
	case consoleMode:
		l.setupWithConsole(c)
		return l
	case volumeMode:
		err := l.setupWithVolume(c)
		if err != nil {
			return nil
		} else {
			return l
		}
	default:
		err := l.setupWithFiles(c)
		if err != nil {
			return nil
		} else {
			return l
		}
	}
}

func (l *fullLogger) Close() error {
	if l.writeConsole {
		return nil
	}

	if atomic.LoadUint32(&(l.initialized)) == 0 {
		return ErrLogNotInitialized
	}

	atomic.StoreUint32(&(l.initialized), 0)

	if l.infoLog != nil {
		if err := l.infoLog.Close(); err != nil {
			return err
		}
	}

	if l.errorLog != nil {
		if err := l.errorLog.Close(); err != nil {
			return err
		}
	}

	if l.severeLog != nil {
		if err := l.severeLog.Close(); err != nil {
			return err
		}
	}

	if l.slowLog != nil {
		if err := l.slowLog.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (l *fullLogger) WithDuration(duration time.Duration) FullLogger {
	f := &fullLogger{
		logEntry: l.logEntry,
		ctx:      l.ctx,

		//
		timeFormat:   l.timeFormat,
		writeConsole: l.writeConsole,
		logLevel:     atomic.LoadUint32(&l.logLevel),
		encoding:     l.encoding,

		//
		infoLog:   l.infoLog,
		errorLog:  l.errorLog,
		severeLog: l.severeLog,
		slowLog:   l.slowLog,

		//
		initialized: atomic.LoadUint32(&l.initialized),
		options:     l.options,
	}
	f.Duration = timex.ReprOfDuration(duration)
	return f
}

func (l *fullLogger) WithContext(ctx context.Context) FullLogger {
	f := &fullLogger{
		logEntry: l.logEntry,
		ctx:      ctx,

		//
		timeFormat:   l.timeFormat,
		writeConsole: l.writeConsole,
		logLevel:     atomic.LoadUint32(&l.logLevel),
		encoding:     l.encoding,

		//
		infoLog:   l.infoLog,
		errorLog:  l.errorLog,
		severeLog: l.severeLog,
		slowLog:   l.slowLog,

		//
		initialized: atomic.LoadUint32(&l.initialized),
		options:     l.options,
	}
	return f
}

func (l *fullLogger) Severe(v ...interface{}) {
	if l.shallLog(SevereLevel) {
		v = append(v, "\n"+string(debug.Stack()))
		l.write(l.severeLog, levelError, formatWithCaller(fmt.Sprint(v...), durationCallerDepth))
	}
}

func (l *fullLogger) Severef(format string, v ...interface{}) {
	if l.shallLog(SevereLevel) {
		v = append(v, "\n"+string(debug.Stack()))
		l.write(l.severeLog, levelError, formatWithCaller(fmt.Sprintf(format, v...), durationCallerDepth))
	}
}

func (l *fullLogger) Error(v ...interface{}) {
	if l.shallLog(ErrorLevel) {
		l.write(l.errorLog, levelError, formatWithCaller(fmt.Sprint(v...), durationCallerDepth))
	}
}

func (l *fullLogger) Errorf(format string, v ...interface{}) {
	if l.shallLog(ErrorLevel) {
		l.write(l.errorLog, levelError, formatWithCaller(fmt.Sprintf(format, v...), durationCallerDepth))
	}
}

func (l *fullLogger) Info(v ...interface{}) {
	if l.shallLog(InfoLevel) {
		l.write(l.infoLog, levelInfo, fmt.Sprint(v...))
	}
}

func (l *fullLogger) Infof(format string, v ...interface{}) {
	if l.shallLog(InfoLevel) {
		l.write(l.infoLog, levelInfo, fmt.Sprintf(format, v...))
	}
}

func (l *fullLogger) Slow(v ...interface{}) {
	if l.shallLog(ErrorLevel) {
		l.write(l.slowLog, levelSlow, fmt.Sprint(v...))
	}
}

func (l *fullLogger) Slowf(format string, v ...interface{}) {
	if l.shallLog(ErrorLevel) {
		l.write(l.slowLog, levelSlow, fmt.Sprintf(format, v...))
	}
}

func (l *fullLogger) write(writer io.Writer, level string, val interface{}) {
	var (
		traceID string
		spanID  string
	)
	if l.ctx != nil {
		traceID = traceIdFromContext(l.ctx)
		spanID = spanIdFromContext(l.ctx)
	}

	switch atomic.LoadUint32(&(l.encoding)) {
	case plainEncodingType:
		l.writePlainAny(writer, level, val, l.Duration, traceID, spanID)
	default:
		l.outputJson(writer, &traceLogger{
			logEntry: logEntry{
				Timestamp: getTimestamp(),
				Level:     level,
				Duration:  l.Duration,
				Content:   val,
			},
			Trace: traceID,
			Span:  spanID,
		})
	}
}

func (l *fullLogger) shallLog(level uint32) bool {
	return atomic.LoadUint32(&(l.logLevel)) <= level
}

func (l *fullLogger) handleOptions(opts []LogOption) {
	for _, opt := range opts {
		opt(&(l.options))
	}
}

func (l *fullLogger) setupLogLevel(c logx.LogConf) {
	switch c.Level {
	case levelInfo:
		l.SetLevel(InfoLevel)
	case levelError:
		l.SetLevel(ErrorLevel)
	case levelSevere:
		l.SetLevel(SevereLevel)
	}
}

func (l *fullLogger) SetLevel(level uint32) {
	atomic.StoreUint32(&(l.logLevel), level)
}

func (l *fullLogger) setupWithConsole(c logx.LogConf) {
	atomic.StoreUint32(&(l.initialized), 1)
	l.writeConsole = true
	l.setupLogLevel(c)

	l.infoLog = newLogWriter(log.New(os.Stdout, "", flags))
	l.errorLog = newLogWriter(log.New(os.Stderr, "", flags))
	l.severeLog = newLogWriter(log.New(os.Stderr, "", flags))
	l.slowLog = newLogWriter(log.New(os.Stderr, "", flags))
}

func (l *fullLogger) setupWithFiles(c logx.LogConf) error {
	var opts []LogOption
	var err error

	if len(c.Path) == 0 {
		return ErrLogPathNotSet
	}

	opts = append(opts, WithCooldownMillis(c.StackCooldownMillis))
	if c.Compress {
		opts = append(opts, WithGzip())
	}
	if c.KeepDays > 0 {
		opts = append(opts, WithKeepDays(c.KeepDays))
	}

	accessFile := path.Join(c.Path, accessFilename)
	errorFile := path.Join(c.Path, errorFilename)
	severeFile := path.Join(c.Path, severeFilename)
	slowFile := path.Join(c.Path, slowFilename)

	atomic.StoreUint32(&(l.initialized), 1)
	l.handleOptions(opts)
	l.setupLogLevel(c)

	if l.infoLog, err = createOutput(accessFile); err != nil {
		return err
	}

	if l.errorLog, err = createOutput(errorFile); err != nil {
		return err
	}

	if l.severeLog, err = createOutput(severeFile); err != nil {
		return err
	}

	if l.slowLog, err = createOutput(slowFile); err != nil {
		return err
	}

	return err
}

func (l *fullLogger) setupWithVolume(c logx.LogConf) error {
	if len(c.ServiceName) == 0 {
		return ErrLogServiceNameNotSet
	}

	c.Path = path.Join(c.Path, c.ServiceName, sysx.Hostname())
	return setupWithFiles(c)
}

func (l *fullLogger) writePlainAny(writer io.Writer, level string, val interface{}, fields ...string) {
	switch v := val.(type) {
	case string:
		l.writePlainText(writer, level, v, fields...)
	case error:
		l.writePlainText(writer, level, v.Error(), fields...)
	case fmt.Stringer:
		l.writePlainText(writer, level, v.String(), fields...)
	default:
		var buf bytes.Buffer
		buf.WriteString(getTimestamp())
		buf.WriteByte(plainEncodingSep)
		buf.WriteString(level)
		for _, item := range fields {
			buf.WriteByte(plainEncodingSep)
			buf.WriteString(item)
		}
		buf.WriteByte(plainEncodingSep)
		if err := json.NewEncoder(&buf).Encode(val); err != nil {
			log.Println(err.Error())
			return
		}
		buf.WriteByte('\n')
		if atomic.LoadUint32(&(l.initialized)) == 0 || writer == nil {
			log.Println(buf.String())
			return
		}

		if _, err := writer.Write(buf.Bytes()); err != nil {
			log.Println(err.Error())
		}
	}
}

func (l *fullLogger) writePlainText(writer io.Writer, level, msg string, fields ...string) {
	var buf bytes.Buffer
	buf.WriteString(getTimestamp())
	buf.WriteByte(plainEncodingSep)
	buf.WriteString(level)
	for _, item := range fields {
		buf.WriteByte(plainEncodingSep)
		buf.WriteString(item)
	}
	buf.WriteByte(plainEncodingSep)
	buf.WriteString(msg)
	buf.WriteByte('\n')
	if atomic.LoadUint32(&(l.initialized)) == 0 || writer == nil {
		log.Println(buf.String())
		return
	}

	if _, err := writer.Write(buf.Bytes()); err != nil {
		log.Println(err.Error())
	}
}

func (l *fullLogger) outputJson(writer io.Writer, info interface{}) {
	if content, err := json.Marshal(info); err != nil {
		log.Println(err.Error())
	} else if atomic.LoadUint32(&(l.initialized)) == 0 || writer == nil {
		log.Println(string(content))
	} else {
		writer.Write(append(content, '\n'))
	}
}
