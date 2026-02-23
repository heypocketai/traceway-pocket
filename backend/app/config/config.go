package config

import "os"

type Cfg struct {
	JWTSecret string

	DBType           string
	PostgresHost     string
	PostgresPort     string
	PostgresDatabase string
	PostgresUsername  string
	PostgresPassword string
	PostgresSSLMode  string
	SQLitePath       string

	ClickhouseType     string
	ClickhouseServer   string
	ClickhouseDatabase string
	ClickhouseUsername string
	ClickhousePassword string
	ClickhouseTLS      string
	ClickhousePath     string

	StorageType string
	StoragePath string
	S3Bucket    string
	S3Region    string
	S3AccessKey string
	S3SecretKey string
	S3Endpoint  string

	SMTPEnabled  string
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	AppBaseURL            string
	CloudMode             string
	MonitoringTracewayURL string
	APIOnly               string
	Ports                 string
	TurnstileSecretKey    string
}

var Config *Cfg

func Init(c *Cfg) { Config = c }

func LoadFromEnv() *Cfg {
	return &Cfg{
		JWTSecret: os.Getenv("JWT_SECRET"),

		DBType:           os.Getenv("DB_TYPE"),
		PostgresHost:     os.Getenv("POSTGRES_HOST"),
		PostgresPort:     os.Getenv("POSTGRES_PORT"),
		PostgresDatabase: os.Getenv("POSTGRES_DATABASE"),
		PostgresUsername:  os.Getenv("POSTGRES_USERNAME"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresSSLMode:  os.Getenv("POSTGRES_SSLMODE"),
		SQLitePath:       os.Getenv("SQLITE_PATH"),

		ClickhouseType:     os.Getenv("CLICKHOUSE_TYPE"),
		ClickhouseServer:   os.Getenv("CLICKHOUSE_SERVER"),
		ClickhouseDatabase: os.Getenv("CLICKHOUSE_DATABASE"),
		ClickhouseUsername: os.Getenv("CLICKHOUSE_USERNAME"),
		ClickhousePassword: os.Getenv("CLICKHOUSE_PASSWORD"),
		ClickhouseTLS:      os.Getenv("CLICKHOUSE_TLS"),
		ClickhousePath:     os.Getenv("CLICKHOUSE_PATH"),

		StorageType: os.Getenv("STORAGE_TYPE"),
		StoragePath: os.Getenv("STORAGE_PATH"),
		S3Bucket:    os.Getenv("S3_BUCKET"),
		S3Region:    os.Getenv("S3_REGION"),
		S3AccessKey: os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey: os.Getenv("S3_SECRET_KEY"),
		S3Endpoint:  os.Getenv("S3_ENDPOINT"),

		SMTPEnabled:  os.Getenv("SMTP_ENABLED"),
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:     os.Getenv("SMTP_FROM"),

		AppBaseURL:            os.Getenv("APP_BASE_URL"),
		CloudMode:             os.Getenv("CLOUD_MODE"),
		MonitoringTracewayURL: os.Getenv("MONITORING_TRACEWAY_URL"),
		APIOnly:               os.Getenv("API_ONLY"),
		Ports:                 os.Getenv("PORTS"),
		TurnstileSecretKey:    os.Getenv("TURNSTILE_SECRET_KEY"),
	}
}
