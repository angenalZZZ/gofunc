package logger

import (
	"bytes"
	"fmt"
	"github.com/angenalZZZ/gofunc/configfile"
	"github.com/angenalZZZ/gofunc/f"
	"github.com/angenalZZZ/gofunc/http/fast"
	"github.com/angenalZZZ/gofunc/log"
	"github.com/valyala/fasttemplate"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config defines the config for logger middleware
type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fast.Ctx) bool
	// Format defines the logging format with defined variables
	// Optional. Default: "${time} - ${method} ${path} - ${ip}\t${ua}\n"
	// Possible values: ip, url, host, method, path, protocol
	// referer, ua, header:<key>, query:<key>, form:<key>, cookie:<key>
	Format string
	// Format json defines
	// Optional Default: ip,method,path,status,latency,query,data
	JsonFormat string
	// Output is a writer where logs are written
	// Default: log.Log
	Output log.Logger
	// Output json defines with JsonFormat
	JsonOutput bool
	// ConfigFile log.yaml
	// Optional if cfg.Output equals nil.
	ConfigFile string
	// Tag extension field.
	Tag string
}

// LogConfig Defines Config File
type LogConfig struct {
	Log log.Config
}

// New middleware.
//  cfg := logger.Config{
//    JsonOutput: true,
//    ConfigFile: "log.yaml",
//    Tag: "1",
//  }
// app.Use(logger.New(cfg))
func New(config ...Config) func(*fast.Ctx) {
	// Init config
	var cfg Config
	// Set config if provided
	if len(config) > 0 {
		cfg = config[0]
	}
	// Set config default values
	if cfg.Format == "" {
		cfg.Format = "${ip} ${method} ${path} -> ${status} - ${latency} <- ${query} -d ${data}"
	}
	if cfg.JsonFormat == "" {
		cfg.JsonFormat = "ip,method,path,status,latency,query,data"
	}
	tags := strings.Split(cfg.JsonFormat, ",")
	if cfg.Output == nil {
		if cfg.ConfigFile == "" {
			cfg.Output = log.Log
		} else {
			logCfg := new(LogConfig)
			if err := configfile.YamlTo(cfg.ConfigFile, logCfg); err != nil {
				_ = fmt.Errorf("%s", err.Error())
			}
			cfg.Output = log.Init(logCfg.Log)
			if log.Log == nil {
				log.Log = cfg.Output
			}
		}
	}
	// Middleware settings
	tmpl := fasttemplate.New(cfg.Format, "${", "}")
	pool := &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 256))
		},
	}
	// Middleware function
	return func(c *fast.Ctx) {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(c) {
			c.Next()
			return
		}
		start := time.Now()
		// handle request
		c.Next()
		// build log
		if cfg.Output == nil {
			return
		}
		stop := time.Now()
		if cfg.JsonOutput == false {
			buf := pool.Get().(*bytes.Buffer)
			buf.Reset()
			defer pool.Put(buf)
			_, err := tmpl.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
				switch tag {
				case "latency":
					return buf.WriteString(stop.Sub(start).String())
				default:
					if b := formatTag(c, tag); b != nil {
						return buf.Write(b)
					} else {
						return 0, nil
					}
				}
			})
			if err != nil {
				buf.WriteString(err.Error())
			}
			l := cfg.Output.Log()
			if cfg.Tag != "" {
				l.Str("tag", cfg.Tag)
			}
			l.Msg(fast.GetString(buf.Bytes()))
		} else {
			l := cfg.Output.Log()
			if cfg.Tag != "" {
				l.Str("tag", cfg.Tag)
			}
			for _, tag := range tags {
				switch tag {
				case "latency":
					l.Str(tag, stop.Sub(start).String())
				default:
					if val := formatTag(c, tag); val != nil {
						l.Bytes(tag, val)
					}
				}
			}
			l.Send()
		}
	}
}

func formatTag(c *fast.Ctx, tag string) []byte {
	switch tag {
	case "referer":
		return c.C.Request.Header.Peek("Referer")
	case "protocol":
		return f.Bytes(c.Protocol())
	case "ip":
		return c.C.RemoteIP()
	case "host":
		return c.C.URI().Host()
	case "method":
		return f.Bytes(c.Method())
	case "path":
		return f.Bytes(c.Path())
	case "query":
		return c.C.QueryArgs().QueryString()
	case "url":
		return c.C.Request.Header.RequestURI()
	case "header":
		return c.C.Request.Header.Header()
	case "data":
		return c.C.Request.Body()
	case "ua":
		return c.C.Request.Header.Peek("User-Agent")
	case "status":
		return f.Bytes(strconv.FormatInt(int64(c.C.Response.StatusCode()), 10))
	default:
		switch {
		case strings.HasPrefix(tag, "header:"):
			return c.C.Request.Header.Peek(tag[7:])
		case strings.HasPrefix(tag, "query:"):
			return f.Bytes(c.Query(tag[6:]))
		case strings.HasPrefix(tag, "form:"):
			return f.Bytes(c.FormValue(tag[5:]))
		case strings.HasPrefix(tag, "cookie:"):
			return f.Bytes(c.Cookies(tag[7:]))
		}
	}
	return nil
}
