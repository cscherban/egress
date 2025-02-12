package config

import (
	"os"
	"time"

	"github.com/livekit/protocol/logger"
	"github.com/livekit/protocol/redis"
	lksdk "github.com/livekit/server-sdk-go"
)

type BaseConfig struct {
	NodeID string // do not supply - will be overwritten

	// required
	Redis     *redis.RedisConfig `yaml:"redis"`      // redis config
	ApiKey    string             `yaml:"api_key"`    // (env LIVEKIT_API_KEY)
	ApiSecret string             `yaml:"api_secret"` // (env LIVEKIT_API_SECRET)
	WsUrl     string             `yaml:"ws_url"`     // (env LIVEKIT_WS_URL)

	// optional
	Logging       logger.Config           `yaml:"logging"`        // logging config
	TemplateBase  string                  `yaml:"template_base"`  // custom template base url
	BackupStorage string                  `yaml:"backup_storage"` // backup file location for failed uploads
	ClusterID     string                  `yaml:"cluster_id"`     // cluster this instance belongs to
	StorageConfig `yaml:",inline"`        // upload config (S3, Azure, GCP, or AliOSS)
	SessionLimits `yaml:"session_limits"` // session duration limits

	// dev/debugging
	Insecure bool        `yaml:"insecure"` // allow chrome to connect to an insecure websocket
	Debug    DebugConfig `yaml:"debug"`    // create dot file on internal error

	// deprecated
	LogLevel string `yaml:"log_level"` // Use Logging instead
}

type DebugConfig struct {
	EnableProfiling bool             `yaml:"enable_profiling"` // create dot file and pprof on internal error
	StorageConfig   `yaml:",inline"` // upload config (S3, Azure, GCP, or AliOSS)
}

type StorageConfig struct {
	S3     *S3Config    `yaml:"s3"`
	Azure  *AzureConfig `yaml:"azure"`
	GCP    *GCPConfig   `yaml:"gcp"`
	AliOSS *S3Config    `yaml:"alioss"`
}

type S3Config struct {
	AccessKey      string `yaml:"access_key"` // (env AWS_ACCESS_KEY_ID)
	Secret         string `yaml:"secret"`     // (env AWS_SECRET_ACCESS_KEY)
	Region         string `yaml:"region"`     // (env AWS_DEFAULT_REGION)
	Endpoint       string `yaml:"endpoint"`
	Bucket         string `yaml:"bucket"`
	ForcePathStyle bool   `yaml:"force_path_style"`
}

type AzureConfig struct {
	AccountName   string `yaml:"account_name"` // (env AZURE_STORAGE_ACCOUNT)
	AccountKey    string `yaml:"account_key"`  // (env AZURE_STORAGE_KEY)
	ContainerName string `yaml:"container_name"`
}

type GCPConfig struct {
	CredentialsJSON string `yaml:"credentials_json"` // (env GOOGLE_APPLICATION_CREDENTIALS)
	Bucket          string `yaml:"bucket"`
}

type SessionLimits struct {
	FileOutputMaxDuration    time.Duration `yaml:"file_output_max_duration"`
	StreamOutputMaxDuration  time.Duration `yaml:"stream_output_max_duration"`
	SegmentOutputMaxDuration time.Duration `yaml:"segment_output_max_duration"`
}

func (c *BaseConfig) initLogger(values ...interface{}) error {
	if c.LogLevel != "" {
		logger.Warnw("log_level deprecated. use logging instead", nil)
		c.Logging.Level = c.LogLevel
	}

	var gstDebug string
	switch c.Logging.Level {
	case "debug":
		gstDebug = "3"
	case "info", "warn":
		gstDebug = "2"
	case "error":
		gstDebug = "1"
	}
	if err := os.Setenv("GST_DEBUG", gstDebug); err != nil {
		return err
	}

	zl, err := logger.NewZapLogger(&c.Logging)
	if err != nil {
		return err
	}

	l := zl.WithValues(values...)

	logger.SetLogger(l, "egress")
	lksdk.SetLogger(l)
	return nil
}
