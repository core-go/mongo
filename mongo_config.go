package mongo

import "time"

type MongoConfig struct {
	Uri                    string        `mapstructure:"uri"`
	Database               string        `mapstructure:"database"`
	AuthSource             string        `mapstructure:"auth_source"`
	MaxPoolSize            uint64        `mapstructure:"max_pool_size"`
	MinPoolSize            uint64        `mapstructure:"min_pool_size"`
	ConnectTimeout         time.Duration `mapstructure:"connect_timeout"`
	SocketTimeout          time.Duration `mapstructure:"socket_timeout"`
	ServerSelectionTimeout time.Duration `mapstructure:"server_selection_timeout"`
	LocalThreshold         time.Duration `mapstructure:"local_threshold"`
	HeartbeatInterval      time.Duration `mapstructure:"heartbeat_interval"`
	ZlibLevel              int           `mapstructure:"zlibLevel"`
}
