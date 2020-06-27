package log

// AConfig Defines From A Config File
type AConfig struct {
	Log *Config
}

// Config log
type Config struct {
	// 日志文件
	Filename string `json:"filename"`
	// 转存大小MB
	MaxSize int `json:"maxsize"`
	// 转存时间days
	MaxAge int `json:"maxage"`
	// 保留最大旧日志文件数
	MaxBackups int `json:"maxbackups"`
	// 使用本地时间,不然文件名就是UTC时间
	LocalTime bool `json:"localtime"`
	// 设置时间格式
	TimeFormat string `json:"timeformat"`
	// 压缩备份gzip
	Compress bool `json:"compress"`
	// 输出位置(选项:file,stdout)
	Writers string `json:"writers"`
	// 日志级别(选项:debug,info,warn,error,fatal,panic,trace)
	Level string `json:"level"`
}
