package mongo

type MongoConfig struct {
	Uri                    string `mapstructure:"uri"`
	Database               string `mapstructure:"database"`
	AuthSource             string `mapstructure:"auth_source"`
	MaxPoolSize            uint64 `mapstructure:"max_pool_size"`
	MinPoolSize            uint64 `mapstructure:"min_pool_size"`
	ConnectTimeout         int64  `mapstructure:"connect_timeout"`
	SocketTimeout          int64  `mapstructure:"socket_timeout"`
	ServerSelectionTimeout int64  `mapstructure:"server_selection_timeout"`
	LocalThreshold         int64  `mapstructure:"local_threshold"`
	HeartbeatInterval      int64  `mapstructure:"heartbeat_interval"`
	ZlibLevel              int    `mapstructure:"zlibLevel"`
}
