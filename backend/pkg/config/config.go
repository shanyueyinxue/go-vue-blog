package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type ConfigOptions struct {
	ConfigPath  string
	ConfigType  string
	ConfigName  string
	EnvPrefix   string
	DotEnvPaths []string
}
type Config[T any] struct {
	viper   *viper.Viper
	options ConfigOptions
}

func NewConfig[T any](options ConfigOptions) *Config[T] {
	v := viper.New()
	v.SetConfigName(options.ConfigName)
	v.AddConfigPath(options.ConfigPath)
	v.SetConfigType(options.ConfigType)
	return &Config[T]{viper: v, options: options}
}

func (c *Config[T]) AddConfigPath(path string) {
	c.viper.AddConfigPath(path)
}

func (c *Config[T]) LoadConfig() (*T, error) {
	if err := c.LoadDotEnv(c.options.DotEnvPaths...); err != nil {
		return nil, errors.New("LoadDotEnv error: " + err.Error())
	}

	if c.options.EnvPrefix != "" {
		c.viper.SetEnvPrefix(c.options.EnvPrefix) // 设置环境变量前缀
	}
	c.viper.AutomaticEnv()                                   // 从环境变量中读取配置文件中的变量
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // 环境变量名替换点为下划线

	if err := c.viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg T
	err := c.viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config[T]) SetDefault(key string, value any) {
	c.viper.SetDefault(key, value)
}

func (c *Config[T]) SetDefaultMap(m map[string]any) {
	for k, v := range m {
		c.viper.SetDefault(k, v)
	}
}

func (c *Config[T]) Set(key string, value any) {
	c.viper.Set(key, value)
}
func (c *Config[T]) Get(key string) any {
	return c.viper.Get(key)
}
func (c *Config[T]) GetString(key string) string {
	return c.viper.GetString(key)
}
func (c *Config[T]) GetBool(key string) bool {
	return c.viper.GetBool(key)
}
func (c *Config[T]) GetInt(key string) int {
	return c.viper.GetInt(key)
}
func (c *Config[T]) GetFloat64(key string) float64 {
	return c.viper.GetFloat64(key)
}

func (c *Config[T]) Viper() *viper.Viper {
	return c.viper
}

func (c *Config[T]) LoadDotEnv(dotenvPaths ...string) error {
	return LoadDotEnv(dotenvPaths...)
}

// LoadDotEnv 加载 .env 文件中的环境变量。
// 未显式传入路径时，默认读取当前目录的 .env，缺失则静默跳过；
// 显式传入的路径必须存在，否则返回错误，便于尽早发现配置缺失。
func LoadDotEnv(filenames ...string) error {
	paths := make([]string, 0)
	if len(filenames) == 0 {
		if _, err := os.Stat(".env"); err != nil {
			return nil
		}
		paths = append(paths, ".env")
	} else {
		paths = filenames
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("dotenv file %q not found: %w", p, err)
		}
		if err := godotenv.Load(p); err != nil {
			return err
		}
	}
	return nil
}
