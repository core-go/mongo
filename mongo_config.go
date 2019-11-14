package mongo

type MongoConfig struct {
	Uri      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}
