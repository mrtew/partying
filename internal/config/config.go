package config

type Config struct {
	Port      string
	JWTSecret string
	MySQLDSN  string
	RedisAddr string
}

func New() *Config {
	return &Config{
		Port:      ":8080",
		JWTSecret: "partying-super-secret-key-2024",
		MySQLDSN:  "root:password@tcp(localhost:3306)/partying?parseTime=true",
		RedisAddr: "localhost:6379",
	}
}
