package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	EnglishCourseID string `json:"english_course_id"`
}

func configFile() string {
	dir, _ := os.Getwd()
	// 向上查找包含 go.mod 的项目根目录
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, ".muke", "config.json")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".muke", "config.json")
}

func Load() (*Config, error) {
	data, err := os.ReadFile(configFile())
	if err != nil {
		return &Config{}, nil // 没有配置文件返回空配置
	}
	var c Config
	return &c, json.Unmarshal(data, &c)
}

func Save(c *Config) error {
	path := configFile()
	os.MkdirAll(filepath.Dir(path), 0700)
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
