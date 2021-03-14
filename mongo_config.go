package mongo

type MongoConfig struct {
	Uri                      string            `mapstructure:"uri"`
	Database                 string            `mapstructure:"database"`
	AuthSource               string            `mapstructure:"auth_source"`
	ReplicaSet               string            `mapstructure:"replica_set"`
	Credential               *CredentialConfig `mapstructure:"credential"`
	Compressors              []string          `mapstructure:"compressors"`
	Hosts                    []string          `mapstructure:"hosts"`
	RetryReads               *bool             `mapstructure:"retry_reads"`
	RetryWrites              *bool             `mapstructure:"retry_writes"`
	AppName                  string            `mapstructure:"app_name"`
	MaxPoolSize              uint64            `mapstructure:"max_pool_size"`
	MinPoolSize              uint64            `mapstructure:"min_pool_size"`
	ConnectTimeout           int64             `mapstructure:"connect_timeout"`
	SocketTimeout            int64             `mapstructure:"socket_timeout"`
	ServerSelectionTimeout   int64             `mapstructure:"server_selection_timeout"`
	LocalThreshold           int64             `mapstructure:"local_threshold"`
	HeartbeatInterval        int64             `mapstructure:"heartbeat_interval"`
	ZlibLevel                int               `mapstructure:"zlibLevel"`
	MaxConnIdleTime          int64             `mapstructure:"max_conn_idle_time"`
	DisableOCSPEndpointCheck *bool             `mapstructure:"disable_ocsp_endpoint_check"`
	Direct                   *bool             `mapstructure:"direct"`
}

type CredentialConfig struct {
	AuthMechanism           *string           `mapstructure:"auth_mechanisms"`
	AuthMechanismProperties map[string]string `mapstructure:"auth_mechanism_properties"`
	AuthSource              *string           `mapstructure:"auth_source"`
	Username                string            `mapstructure:"username"`
	Password                string            `mapstructure:"password"`
	PasswordSet             *bool             `mapstructure:"password_set"`
}
