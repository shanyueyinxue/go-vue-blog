package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"blog/pkg/config"
)

type appConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
}

func TestNewConfig(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "app.yaml"), []byte("name: newconfig\nport: 1"), 0644)
	os.WriteFile(filepath.Join(dir, ".env"), nil, 0644)

	cfg := config.NewConfig[appConfig](config.ConfigOptions{
		ConfigPath:  dir,
		ConfigType:  "yaml",
		ConfigName:  "app",
		DotEnvPaths: []string{filepath.Join(dir, ".env")},
	})

	result, err := cfg.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "newconfig", result.Name)
	assert.Equal(t, 1, result.Port)

	v := cfg.Viper()
	assert.NotNil(t, v)
}

func TestConfig_AddConfigPath(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.yaml"), []byte("name: addpath\nport: 1"), 0644)
	os.WriteFile(filepath.Join(dir, ".env"), nil, 0644)

	cfg := config.NewConfig[appConfig](config.ConfigOptions{
		ConfigPath:  "nonexistent_dir",
		ConfigType:  "yaml",
		ConfigName:  "test",
		DotEnvPaths: []string{filepath.Join(dir, ".env")},
	})

	cfg.AddConfigPath(dir)

	result, err := cfg.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "addpath", result.Name)
}

func TestConfig_SetDefault(t *testing.T) {
	cfg := config.NewConfig[string](config.ConfigOptions{})
	cfg.SetDefault("key", "default_value")
	assert.Equal(t, "default_value", cfg.GetString("key"))
}

func TestConfig_SetDefaultMap(t *testing.T) {
	cfg := config.NewConfig[string](config.ConfigOptions{})
	cfg.SetDefaultMap(map[string]any{
		"key1": "val1",
		"key2": 42,
		"key3": true,
	})
	assert.Equal(t, "val1", cfg.GetString("key1"))
	assert.Equal(t, 42, cfg.GetInt("key2"))
	assert.Equal(t, true, cfg.GetBool("key3"))
}

func TestConfig_SetAndGet(t *testing.T) {
	cfg := config.NewConfig[string](config.ConfigOptions{})
	cfg.Set("key", "value")
	assert.Equal(t, "value", cfg.Get("key"))
}

func TestConfig_GetTypedGetters(t *testing.T) {
	cfg := config.NewConfig[string](config.ConfigOptions{})
	cfg.Set("str", "hello")
	cfg.Set("bool", true)
	cfg.Set("int", 42)
	cfg.Set("float", 3.14)

	assert.Equal(t, "hello", cfg.GetString("str"))
	assert.Equal(t, true, cfg.GetBool("bool"))
	assert.Equal(t, 42, cfg.GetInt("int"))
	assert.Equal(t, 3.14, cfg.GetFloat64("float"))
}

func TestConfig_GetDefaults(t *testing.T) {
	cfg := config.NewConfig[string](config.ConfigOptions{})

	assert.Equal(t, "", cfg.GetString("nonexistent"))
	assert.Equal(t, false, cfg.GetBool("nonexistent"))
	assert.Equal(t, 0, cfg.GetInt("nonexistent"))
	assert.Equal(t, float64(0), cfg.GetFloat64("nonexistent"))
}

func TestConfig_Viper(t *testing.T) {
	cfg := config.NewConfig[string](config.ConfigOptions{})
	v := cfg.Viper()
	require.NotNil(t, v)

	cfg.Set("key", "direct_value")
	assert.Equal(t, "direct_value", v.GetString("key"))
}

func TestConfig_LoadConfig_Success_YAML(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("name: hello\nport: 8080"), 0644)
	os.WriteFile(filepath.Join(dir, ".env"), nil, 0644)

	cfg := config.NewConfig[appConfig](config.ConfigOptions{
		ConfigPath:  dir,
		ConfigType:  "yaml",
		ConfigName:  "config",
		DotEnvPaths: []string{filepath.Join(dir, ".env")},
	})
	cfg.SetDefault("name", "default_name")
	cfg.SetDefault("port", 3000)

	result, err := cfg.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "hello", result.Name)
	assert.Equal(t, 8080, result.Port)
}

func TestConfig_LoadConfig_Success_JSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"name":"from_json","port":9090}`), 0644)
	os.WriteFile(filepath.Join(dir, ".env"), nil, 0644)

	cfg := config.NewConfig[appConfig](config.ConfigOptions{
		ConfigPath:  dir,
		ConfigType:  "json",
		ConfigName:  "config",
		DotEnvPaths: []string{filepath.Join(dir, ".env")},
	})

	result, err := cfg.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "from_json", result.Name)
	assert.Equal(t, 9090, result.Port)
}

func TestConfig_LoadConfig_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env"), nil, 0644)

	cfg := config.NewConfig[appConfig](config.ConfigOptions{
		ConfigPath:  dir,
		ConfigType:  "yaml",
		ConfigName:  "nonexistent",
		DotEnvPaths: []string{filepath.Join(dir, ".env")},
	})

	_, err := cfg.LoadConfig()
	assert.Error(t, err)
}

func TestConfig_LoadConfig_BadFormat(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("not: valid: yaml: ["), 0644)
	os.WriteFile(filepath.Join(dir, ".env"), nil, 0644)

	cfg := config.NewConfig[appConfig](config.ConfigOptions{
		ConfigPath:  dir,
		ConfigType:  "yaml",
		ConfigName:  "config",
		DotEnvPaths: []string{filepath.Join(dir, ".env")},
	})

	_, err := cfg.LoadConfig()
	assert.Error(t, err)
}

func TestConfig_LoadConfig_WithEnvPrefix(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("name: from_file\nport: 0"), 0644)
	os.WriteFile(filepath.Join(dir, ".env"), nil, 0644)

	const envKey = "MYAPP_NAME"
	os.Setenv(envKey, "from_env")
	t.Cleanup(func() { os.Unsetenv(envKey) })

	cfg := config.NewConfig[appConfig](config.ConfigOptions{
		ConfigPath:  dir,
		ConfigType:  "yaml",
		ConfigName:  "config",
		EnvPrefix:   "MYAPP",
		DotEnvPaths: []string{filepath.Join(dir, ".env")},
	})

	result, err := cfg.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "from_env", result.Name)
}

func TestConfig_LoadConfig_WithDotEnv(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("name: unchanged\nport: 0"), 0644)
	os.WriteFile(filepath.Join(dir, ".env"), []byte("NAME=from_dotenv"), 0644)
	t.Cleanup(func() { os.Unsetenv("NAME") })

	cfg := config.NewConfig[appConfig](config.ConfigOptions{
		ConfigPath:  dir,
		ConfigType:  "yaml",
		ConfigName:  "config",
		DotEnvPaths: []string{filepath.Join(dir, ".env")},
	})

	result, err := cfg.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "from_dotenv", result.Name)
}

func TestConfig_LoadConfig_DotEnvNotFound(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("name: hello\nport: 0"), 0644)

	cfg := config.NewConfig[appConfig](config.ConfigOptions{
		ConfigPath:  dir,
		ConfigType:  "yaml",
		ConfigName:  "config",
		DotEnvPaths: []string{filepath.Join(dir, "nonexistent.env")},
	})

	_, err := cfg.LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LoadDotEnv error")
}

func TestConfig_LoadConfig_SetOverrides(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("name: file_value\nport: 8080"), 0644)
	os.WriteFile(filepath.Join(dir, ".env"), nil, 0644)

	cfg := config.NewConfig[appConfig](config.ConfigOptions{
		ConfigPath:  dir,
		ConfigType:  "yaml",
		ConfigName:  "config",
		DotEnvPaths: []string{filepath.Join(dir, ".env")},
	})
	cfg.Set("name", "set_value")
	cfg.Set("port", 9999)

	result, err := cfg.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "set_value", result.Name)
	assert.Equal(t, 9999, result.Port)
}

func TestConfig_LoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, []byte("METHOD_TEST_VAR=from_method"), 0644)
	t.Cleanup(func() { os.Unsetenv("METHOD_TEST_VAR") })

	cfg := config.NewConfig[string](config.ConfigOptions{})
	err := cfg.LoadDotEnv(envFile)
	require.NoError(t, err)
	assert.Equal(t, "from_method", os.Getenv("METHOD_TEST_VAR"))
}

func TestLoadDotEnv_CustomPath(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, "custom.env")
	os.WriteFile(envFile, []byte("CUSTOM_VAR=hello_custom"), 0644)
	t.Cleanup(func() { os.Unsetenv("CUSTOM_VAR") })

	err := config.LoadDotEnv(envFile)
	require.NoError(t, err)
	assert.Equal(t, "hello_custom", os.Getenv("CUSTOM_VAR"))
}

func TestLoadDotEnv_DefaultPath(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".env"), []byte("DEFAULT_VAR=hello_default"), 0644)
	t.Cleanup(func() { os.Unsetenv("DEFAULT_VAR") })
	t.Chdir(dir)

	err := config.LoadDotEnv()
	require.NoError(t, err)
	assert.Equal(t, "hello_default", os.Getenv("DEFAULT_VAR"))
}

func TestLoadDotEnv_MultiplePaths(t *testing.T) {
	dir := t.TempDir()
	env1 := filepath.Join(dir, ".env")
	env2 := filepath.Join(dir, ".env.local")
	os.WriteFile(env1, []byte("VAR1=from_env1"), 0644)
	os.WriteFile(env2, []byte("VAR2=from_env2"), 0644)
	t.Cleanup(func() {
		os.Unsetenv("VAR1")
		os.Unsetenv("VAR2")
	})

	err := config.LoadDotEnv(env1, env2)
	require.NoError(t, err)
	assert.Equal(t, "from_env1", os.Getenv("VAR1"))
	assert.Equal(t, "from_env2", os.Getenv("VAR2"))
}

func TestLoadDotEnv_FileNotFound(t *testing.T) {
	err := config.LoadDotEnv("nonexistent_file_12345.env")
	assert.Error(t, err)
}

func TestLoadDotEnv_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, nil, 0644)

	err := config.LoadDotEnv(envFile)
	assert.NoError(t, err)
}
