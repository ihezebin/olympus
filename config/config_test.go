package config

import (
	"os"
	"strings"
	"testing"
)

func TestConfigFile(t *testing.T) {
	config := NewWithFilePath("config.json")
	t.Log(config.FilePaths())
	dest := make(map[string]interface{})
	if err := config.Load(&dest); err != nil {
		t.Fatal(err)
	}

	t.Log(dest)
}

func TestConfigReader(t *testing.T) {
	data := `{
  "name": "olympus",
  "version": "1.0.0"
}`
	config := NewWithReader(strings.NewReader(data))
	dest := make(map[string]interface{})
	if err := config.Load(&dest); err != nil {
		t.Fatal(err)
	}

	t.Log(dest)
}

func TestConfigEnvVars(t *testing.T) {
	// 设置环境变量
	os.Setenv("APP_NAME", "MyViperApp")
	os.Setenv("APP_VERSION", "1.0.0")
	os.Setenv("SERVER_PORT", "8080")

	// 定义配置结构体
	type AppConfig struct {
		AppName    string `env:"APP_NAME" json:"app_name"`
		AppVersion string `env:"APP_VERSION" json:"app_version"`
		Name       string `json:"name" env:"NAME"`
		Version    string `json:"version"`
		Server     struct {
			Port int `json:"port" env:"SERVER_PORT"`
		} `json:"server"`
	}

	// 创建配置实例
	config := NewWithFilePath("config.json", WithEnv())

	// 加载配置
	var dest AppConfig
	if err := config.Load(&dest); err != nil {
		t.Fatal(err)
	}

	// 验证环境变量是否正确绑定
	if dest.AppName != "MyViperApp" {
		t.Errorf("Expected AppName to be 'MyViperApp', got '%s'", dest.AppName)
	}
	if dest.AppVersion != "1.0.0" {
		t.Errorf("Expected AppVersion to be '1.0.0', got '%s'", dest.AppVersion)
	}

	t.Logf("%+v", dest)
}
