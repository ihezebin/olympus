package config

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	kernel      *viper.Viper
	destTagName DestTagName
	dataType    DataType
	fileName    string
	filePaths   []string
	reader      io.Reader
	env         bool
}

func NewWithFilePath(path string, opts ...Option) *Config {
	config := &Config{
		destTagName: defaultDestTagName,
	}
	for _, opt := range opts {
		opt(config)
	}

	// if not sure file type, use json
	if config.dataType == "" {
		config.dataType = DataTypeJson
	}

	filePaths := make([]string, 0)
	// if path is relative path, find from current working directory and app executable path
	if !filepath.IsAbs(path) {
		workPath, err := os.Getwd()
		if err == nil {
			workFilePath := filepath.Join(workPath, path)
			filePaths = append(filePaths, workFilePath)
		}
		appPath := filepath.Dir(os.Args[0])
		appFilePath := filepath.Join(appPath, path)
		filePaths = append(filePaths, appFilePath)
	} else {
		filePaths = append(filePaths, path)
	}

	allDir := true
	kernel := viper.New()
	for _, filePath := range filePaths {
		stat, err := os.Stat(filePath)
		if err == nil && stat != nil {
			if !stat.IsDir() {
				kernel.SetConfigFile(filePath)
				allDir = false
				break
			}
			kernel.AddConfigPath(filePath)
		}
	}
	// if all paths are directories, need set config name and type
	if allDir {
		if config.fileName == "" {
			config.fileName = defaultFileName
		}
		kernel.SetConfigName(config.fileName)
		kernel.SetConfigType(string(config.dataType))
	}

	config.filePaths = filePaths
	config.kernel = kernel

	return config
}

func NewWithReader(reader io.Reader, opts ...Option) *Config {
	config := &Config{
		reader:      reader,
		destTagName: defaultDestTagName,
	}

	for _, opt := range opts {
		opt(config)
	}

	if config.dataType == "" {
		config.dataType = DataTypeJson
	}

	kernel := viper.New()
	// reader need know data type
	kernel.SetConfigType(string(config.dataType))

	config.kernel = kernel

	return config
}

func (c *Config) Kernel() *viper.Viper {
	return c.kernel
}

func (c *Config) Load(dest interface{}) error {
	// 读取配置文件
	if c.kernel.ConfigFileUsed() != "" || len(c.filePaths) > 0 {
		if err := c.kernel.ReadInConfig(); err != nil {
			return errors.Wrap(err, "failed to load config file path")
		}
	}
	if c.reader != nil {
		if err := c.kernel.MergeConfig(c.reader); err != nil {
			return errors.Wrap(err, "failed to load config reader")
		}
	}

	// 处理环境变量
	if c.env {
		if err := c.bindEnvVars(dest); err != nil {
			return errors.Wrap(err, "failed to bind environment variables")
		}
	}

	// 使用 viper 的 Unmarshal 功能
	return c.Kernel().Unmarshal(dest, func(d *mapstructure.DecoderConfig) {
		d.TagName = string(c.destTagName)
	})
}

// bindEnvVars 处理环境变量绑定
func (c *Config) bindEnvVars(dest interface{}) error {
	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return nil
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		envTag := field.Tag.Get("env")
		if envTag == "" {
			continue
		}

		// 获取环境变量值
		envValue := os.Getenv(envTag)
		if envValue == "" {
			continue
		}

		// 设置 viper 中的值
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			fieldName = strings.Split(jsonTag, ",")[0]
		}
		c.kernel.Set(fieldName, envValue)
	}

	return nil
}

func (c *Config) FilePaths() []string {
	return c.filePaths
}
