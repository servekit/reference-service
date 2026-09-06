// Package config defines the reference-service configuration shape and loads it
// via go-common's configx.
//
// serviceName and envPrefix are the two anchors external tooling (systemd
// unit, Docker env, k8s configmap) must agree on with this binary.
package config

import (
	"github.com/servekit/go-common/configx"

	"github.com/servekit/go-common/logging"
)

// serviceName identifies this binary in config file lookup (/etc/<name>) and
// the <NAME>_CONFIG env var. envPrefix scopes all env overrides under
// REFERENCE_SERVICE_.
const (
	serviceName = "reference-service"
	envPrefix   = "REFERENCE_SERVICE"
)

// Config holds all configuration for the reference service.
//
// Sub-config fields are pointers per golang-development skill §14: keeps
// style consistent with the outer *Config return + functional options, and
// avoids large-struct copies. Note: configx (viper) ALWAYS allocates nil
// pointer fields during unmarshal, so cfg.X == nil is never true — express
// "optional dependency" via an inner Enabled bool, not pointer nil-check.
type Config struct {
	Server *ServerConfig
	Cron   *CronConfig
	Log    *logging.Config
}

// ServerConfig holds the gRPC server address.
type ServerConfig struct {
	// GRPCAddr defaults to ":19094" — next free port in the 1909x family for the gRPC port.
	GRPCAddr string `default:":19094"`
}

// CronConfig configures the internal cronx instance used by jobs.Scheduler.
// Empty by default — jobs.Scheduler is wired but registers no jobs until
// setupJobs adds them.
type CronConfig struct {
	// Timezone for cron expression evaluation. Defaults to Asia/Shanghai.
	Timezone string `default:"Asia/Shanghai"`
}

// Load reads config from the standard configx locations:
//   - /etc/reference-service/config.yaml
//   - ./config.yaml
//   - $REFERENCE_SERVICE_CONFIG
//
// config.example.yaml is fully placeholder-driven: every value is a ${VAR}
// reference expanded from the process environment (WithExpandEnv), so the file
// holds structure only — all actual values live in .env.example / the runtime
// env. Env vars under $REFERENCE_SERVICE_ also override file values via
// viper's automatic binding; struct `default:` tags apply last.
func Load() (*Config, error) {
	var cfg Config
	if err := configx.Load(&cfg,
		configx.WithServiceName(serviceName),
		configx.WithEnvPrefix(envPrefix),
		// Expand ${VAR} placeholders in config values from the process env.
		configx.WithExpandEnv(),
	); err != nil {
		return nil, err
	}
	return &cfg, nil
}
