package config

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/go-viper/mapstructure/v2"
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

		execPath, err := os.Executable()
		if err == nil {
			execPath := filepath.Dir(execPath)
			execFilePath := filepath.Join(execPath, path)
			filePaths = append(filePaths, execFilePath)
		}

		appPath := filepath.Dir(os.Args[0])
		if appPath != execPath {
			appFilePath := filepath.Join(appPath, path)
			filePaths = append(filePaths, appFilePath)
		}
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

	err := c.Kernel().Unmarshal(dest, func(d *mapstructure.DecoderConfig) {
		d.TagName = string(c.destTagName)
	})
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal config")
	}

	if c.env {
		if err := c.bindEnv(dest); err != nil {
			return errors.Wrap(err, "failed to bind environment variables")
		}
	}

	return nil
}

// bindEnv 读取环境变量
func (c *Config) bindEnv(dest interface{}) error {
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.Elem().Kind() != reflect.Struct {
		return nil
	}

	val := destVal.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		envTag := field.Tag.Get("env")

		switch field.Type.Kind() {
		case reflect.String:
			if envTag != "" {
				envValue := os.Getenv(envTag)
				if envValue != "" {
					val.Field(i).SetString(envValue)
				}
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if envTag != "" {
				envValue := os.Getenv(envTag)
				if envValue != "" {
					intValue, err := strconv.Atoi(envValue)
					if err != nil {
						return errors.Wrap(err, "failed to convert environment variable to int")
					}
					val.Field(i).SetInt(int64(intValue))
				}
			}
		case reflect.Float64, reflect.Float32:
			if envTag != "" {
				envValue := os.Getenv(envTag)
				if envValue != "" {
					floatValue, err := strconv.ParseFloat(envValue, 64)
					if err != nil {
						return errors.Wrap(err, "failed to convert environment variable to float")
					}
					val.Field(i).SetFloat(floatValue)
				}
			}
		case reflect.Bool:
			if envTag != "" {
				envValue := os.Getenv(envTag)
				if envValue != "" {
					boolValue, err := strconv.ParseBool(envValue)
					if err != nil {
						return errors.Wrap(err, "failed to convert environment variable to bool")
					}
					val.Field(i).SetBool(boolValue)
				}
			}
		case reflect.Struct:
			if err := c.bindEnv(val.Field(i).Addr().Interface()); err != nil {
				return errors.Wrapf(err, "failed to bind environment variables in child struct: %s", field.Name)
			}
		}
	}

	return nil
}

func (c *Config) FilePaths() []string {
	return c.filePaths
}
