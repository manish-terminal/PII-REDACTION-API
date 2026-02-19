package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port            string `envconfig:"PORT" default:"8080"`
	LogLevel        string `envconfig:"LOG_LEVEL" default:"info"`
	AWSRegion       string `envconfig:"AWS_REGION" default:"us-east-1"`
	DynamoTableName string `envconfig:"DYNAMO_TABLE_NAME" default:"pii-tokens"`
	APIKey          string `envconfig:"API_KEY" default:"sk_test_123"`
	EnableNER       bool   `envconfig:"ENABLE_NER" default:"false"`
}

func Load() (Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	return cfg, err
}
