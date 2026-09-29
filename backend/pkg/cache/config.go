package cache

type CacheConfig struct {
	Driver string       `mapstructure:"driver"` // redis or memory
	Redis  RedisConfig  `mapstructure:"redis"`
	Memory MemoryConfig `mapstructure:"memory"`
}

type RedisConfig struct {
	Host               string `mapstructure:"host"`
	Port               string `mapstructure:"port"`
	Password           string `mapstructure:"password"`
	DB                 int    `mapstructure:"db"`
	MaxIdleConnections int    `mapstructure:"maxIdleConnections"`
	MaxOpenConnections int    `mapstructure:"maxOpenConnections"`
	ConnMaxLifetime    int    `mapstructure:"connMaxLifetime"`
	ReadTimeout        int    `mapstructure:"readTimeout"`
	WriteTimeout       int    `mapstructure:"writeTimeout"`
	IdleTimeout        int    `mapstructure:"idleTimeout"`
	WaitTimeout        int    `mapstructure:"waitTimeout"`
}

type MemoryConfig struct {
	CleanupInterval int `mapstructure:"cleanupInterval"`
}
