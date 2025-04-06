package config

type Option func(c *Config)

func WithDataType(dataType DataType) Option {
	return func(c *Config) {
		c.dataType = dataType
	}
}

func WithDestTagName(tagName DestTagName) Option {
	return func(c *Config) {
		c.destTagName = tagName
	}
}

func WithFileName(fileName string) Option {
	return func(c *Config) {
		c.fileName = fileName
	}
}

// WithEnv 使用环境变量, 优先级最高，会覆盖配置文件
func WithEnv() Option {
	return func(c *Config) {
		c.env = true
	}
}
