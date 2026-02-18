package config

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/platform-mesh/golang-commons/context/keys"
	"github.com/platform-mesh/golang-commons/traces"
)

func SetConfigInContext(ctx context.Context, config any) context.Context {
	return context.WithValue(ctx, keys.ConfigCtxKey, config)
}

func LoadConfigFromContext(ctx context.Context) any {
	return ctx.Value(keys.ConfigCtxKey)
}

type ImageConfig struct {
	Name string
	Tag  string
}

type LogConfig struct {
	Level  string
	NoJson bool
}

type MetricsConfig struct {
	BindAddress string
	Secure      bool
}

type TracingConfig struct {
	Enabled   bool
	Collector traces.Config
}

type LeaderElectionConfig struct {
	Enabled bool
}

type SentryConfig struct {
	Dsn string
}

type CommonServiceConfig struct {
	DebugLabelValue         string
	MaxConcurrentReconciles int
	Environment             string
	Region                  string
	Kubeconfig              string
	IsLocal                 bool

	Image ImageConfig

	Log LogConfig

	ShutdownTimeout        time.Duration
	Metrics                MetricsConfig
	Tracing                TracingConfig
	EnableHTTP2            bool
	HealthProbeBindAddress string

	LeaderElectionEnabled bool

	Sentry SentryConfig
}

// generateFlagSet generates a pflag.FlagSet from a struct based on its `mapstructure` tags.
func generateFlagSet(config any) (*pflag.FlagSet, error) {
	flagSet := pflag.NewFlagSet("generated", pflag.ContinueOnError)
	err := traverseStruct(reflect.ValueOf(config), flagSet, "")
	return flagSet, err
}

// traverseStruct recursively traverses a struct and adds flags to the FlagSet.
func traverseStruct(value reflect.Value, flagSet *pflag.FlagSet, prefix string) error {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return errors.New("value must be a struct")
	}

	typ := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := value.Field(i)

		// Get the `mapstructure` tag
		tag := field.Tag.Get("mapstructure")
		if tag == "" || tag == "-" {
			continue
		}

		defaultValueTag := field.Tag.Get("default")
		defaultStrValue := ""
		if defaultValueTag != "" {
			defaultStrValue = defaultValueTag
		}

		descriptionValueTag := field.Tag.Get("description")
		descriptionStrValue := ""
		if descriptionValueTag != "" {
			descriptionStrValue = descriptionValueTag
		}

		// Handle nested structs
		if fieldValue.Kind() == reflect.Struct {
			if tag == ",squash" {
				err := traverseStruct(fieldValue, flagSet, "")
				if err != nil {
					return err
				}
			} else {
				err := traverseStruct(fieldValue, flagSet, prefix+tag+".")
				if err != nil {
					return err
				}
			}
			continue
		}

		description := fmt.Sprintf("Set the %s", tag)
		if descriptionStrValue != "" {
			description = descriptionStrValue
		}

		// Add flags based on the field type
		switch fieldValue.Kind() {
		case reflect.String:
			flagSet.String(prefix+tag, defaultStrValue, description)
		case reflect.Int, reflect.Int64:
			if fieldValue.Type() == reflect.TypeOf(time.Duration(0)) {
				var durVal time.Duration
				if defaultStrValue != "" {
					parsedDurVal, err := time.ParseDuration(defaultStrValue)
					if err != nil {
						return fmt.Errorf("invalid duration value for field %s: %w", field.Name, err)
					}
					durVal = parsedDurVal
				}

				durDescription := fmt.Sprintf("Set the %s in seconds", tag)
				if descriptionStrValue != "" {
					durDescription = descriptionStrValue
				}
				flagSet.Duration(prefix+tag, durVal, durDescription)
			} else {
				i := 0
				if defaultStrValue != "" {
					parsedInt, err := strconv.Atoi(defaultStrValue)
					if err != nil {
						return err
					}
					i = parsedInt
				}
				flagSet.Int(prefix+tag, i, description)
			}
		case reflect.Bool:
			var defaultBoolValue bool
			if defaultStrValue != "" {
				b, err := strconv.ParseBool(defaultStrValue)
				if err != nil {
					return err
				}
				defaultBoolValue = b
			}
			flagSet.Bool(prefix+tag, defaultBoolValue, description)
		case reflect.Slice:
			var defaultSliceValue []string
			if defaultStrValue != "" {
				defaultSliceValue = strings.Split(defaultStrValue, ",")
			}
			if fieldValue.Type().Elem().Kind() != reflect.String {
				return fmt.Errorf("unsupported slice element type %s for field %s", fieldValue.Type().Elem().Kind(), field.Name)
			}
			flagSet.StringSlice(prefix+tag, defaultSliceValue, description)
		default:
			return fmt.Errorf("unsupported field type %s for field %s", fieldValue.Kind(), field.Name)
		}

	}
	return nil
}

func NewDefaultConfig() *CommonServiceConfig {

	config := &CommonServiceConfig{
		DebugLabelValue:         "",
		MaxConcurrentReconciles: 10,
		Environment:             "",
		Region:                  "local",
		Kubeconfig:              "",
		IsLocal:                 false,

		Image: ImageConfig{
			Name: "",
			Tag:  "",
		},

		Log: LogConfig{
			Level:  "info",
			NoJson: false,
		},

		ShutdownTimeout: time.Minute,

		Metrics: MetricsConfig{
			BindAddress: ":9090",
			Secure:      false,
		},

		Tracing: TracingConfig{
			Enabled:   false,
			Collector: traces.Config{},
		},

		EnableHTTP2:            true,
		HealthProbeBindAddress: ":8090",

		LeaderElectionEnabled: false,

		Sentry: SentryConfig{
			Dsn: "",
		},
	}

	return config
}

func (c *CommonServiceConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.DebugLabelValue, "debug-label-value", c.DebugLabelValue, "Set the debug label value")
	fs.IntVar(&c.MaxConcurrentReconciles, "max-concurrent-reconciles", c.MaxConcurrentReconciles, "Set the max concurrent reconciles")
	fs.StringVar(&c.Environment, "environment", c.Environment, "Set the environment of the service")
	fs.StringVar(&c.Region, "region", c.Region, "Set the region of the service, e.g. local, dev, staging, prod")
	fs.StringVar(&c.Kubeconfig, "kubeconfig", c.Kubeconfig, "Set the kubeconfig path")
	fs.BoolVar(&c.IsLocal, "is-local", c.IsLocal, "Flagging execution to be local")

	fs.StringVar(&c.Image.Name, "image-name", c.Image.Name, "Set the image name")
	fs.StringVar(&c.Image.Tag, "image-tag", c.Image.Tag, "Set the image tag")

	fs.StringVar(&c.Log.Level, "log-level", c.Log.Level, "Set the log level")
	fs.BoolVar(&c.Log.NoJson, "no-json", c.Log.NoJson, "Disable JSON logging")

	fs.DurationVar(&c.ShutdownTimeout, "shutdown-timeout", c.ShutdownTimeout, "Set the shutdown timeout as duration in seconds, e.g. 30s, 1m, 2h")
	fs.StringVar(&c.Metrics.BindAddress, "metrics-bind-address", c.Metrics.BindAddress, "Set the metrics bind address")
	fs.BoolVar(&c.Metrics.Secure, "metrics-secure", c.Metrics.Secure, "Set if metrics should be exposed via https")

	fs.BoolVar(&c.Tracing.Enabled, "tracing-enabled", c.Tracing.Enabled, "Enable tracing for the service")
	fs.StringVar(&c.Tracing.Collector.ServiceName, "tracing-config-service-name", c.Tracing.Collector.ServiceName, "Set the tracing service name used in traces")
	fs.StringVar(&c.Tracing.Collector.ServiceVersion, "tracing-config-service-version", c.Tracing.Collector.ServiceVersion, "Set the tracing service version used in traces")
	fs.StringVar(&c.Tracing.Collector.CollectorEndpoint, "tracing-config-collector-endpoint", c.Tracing.Collector.CollectorEndpoint, "Set the tracing collector endpoint used to send traces to the collector")

	fs.BoolVar(&c.EnableHTTP2, "enable-http2", c.EnableHTTP2, "Toggle to disable metrics/webhook serving using http2")
	fs.StringVar(&c.HealthProbeBindAddress, "health-probe-bind-address", c.HealthProbeBindAddress, "Set the health probe bind address")

	fs.BoolVar(&c.LeaderElectionEnabled, "leader-elect", c.LeaderElectionEnabled, "Enable leader election for the controller manager")

	fs.StringVar(&c.Sentry.Dsn, "sentry-dsn", c.Sentry.Dsn, "Set the Sentry DSN for error reporting")
}

func BindConfigToFlags(v *viper.Viper, cmd *cobra.Command, config any) error {
	flagSet, err := generateFlagSet(config)
	if err != nil {
		return fmt.Errorf("failed to generate flag set: %w", err)
	}
	err = v.BindPFlags(flagSet)
	if err != nil {
		return err
	}

	cmd.Flags().AddFlagSet(flagSet)

	cobra.OnInitialize(unmarshalIntoStruct(v, config))

	return nil
}

// unmarshalIntoStruct returns a function that unmarshal viper config into cfg and panics on error.
func unmarshalIntoStruct(v *viper.Viper, cfg any) func() {
	return func() {
		if err := v.Unmarshal(cfg); err != nil {
			panic(fmt.Errorf("failed to unmarshal config: %w", err))
		}
	}
}
