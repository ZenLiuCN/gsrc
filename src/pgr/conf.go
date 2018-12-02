package pgr

import (
	"encoding/json"
	"github.com/jackc/pgx"
	"logger"
	"time"
)

var (
	log *loggerWarp
)

func SetLogger(l *logger.Logger) {
	log = &loggerWarp{l}
}

type Conf struct {
	Host                 string // host (e.g. localhost) or path to unix domain socket directory (e.g. /private/tmp)
	Port                 uint16 // default: 5432
	Database             string
	User                 string // default: OS user name
	Password             string //TLSConfig         *tls.Config // config for TLS connection -- nil disables TLS
	UseFallbackTLS       bool   // Try FallbackTLSConfig if connecting with TLSConfig fails. Used for preferring TLS, but allowing unencrypted, or vice-versa 	//FallbackTLSConfig *tls.Config // config for fallback TLS connection (only used if UseFallBackTLS is true)-- nil disables TLS
	LogLevel             int
	RuntimeParams        map[string]string // Run-time parameters to set on connection as session default values (e.g. search_path or application_name)
	PreferSimpleProtocol bool
	MaxConnections       int
	AcquireTimeout       time.Duration
}

func (c *Conf) CreatePool() (p *pgx.ConnPool, e error) {
	return pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:                 c.Host,
			Port:                 c.Port,
			Database:             c.Database,
			User:                 c.User,
			Password:             c.Password,
			UseFallbackTLS:       c.UseFallbackTLS,
			LogLevel:             c.LogLevel,
			RuntimeParams:        c.RuntimeParams,
			PreferSimpleProtocol: c.PreferSimpleProtocol,
			Logger:               log,
		},
		MaxConnections: c.MaxConnections,
		AcquireTimeout: c.AcquireTimeout,
	})
}

type loggerWarp struct {
	*logger.Logger
}

func (l loggerWarp) Log(level pgx.LogLevel, msg string, data map[string]interface{}) {
	if level == pgx.LogLevelNone {
		return
	}
	d,_:=json.MarshalIndent(data, "", " ")
	l.Printf(warpLogLevel(level), msg+"\n %s", string(d))
}
func warpLogLevel(l pgx.LogLevel) logger.LogLevel {
	switch l {
	case pgx.LogLevelTrace:
		return logger.TRACE
	case pgx.LogLevelDebug:
		return logger.DEBUG
	case pgx.LogLevelInfo:
		return logger.DEBUG
	case pgx.LogLevelWarn:
		return logger.WARN
	case pgx.LogLevelError:
		return logger.ERROR
	case pgx.LogLevelNone:
		return logger.FATAL
	default:
		return logger.FATAL
	}
}
