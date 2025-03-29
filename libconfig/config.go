package libconfig

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/aliocode/golib/libcloser"
	"github.com/aliocode/golib/liblog"
	"github.com/aliocode/golib/libmode"
	"github.com/caarlos0/env/v11"
	"github.com/pelletier/go-toml/v2"
)

var (
	ErrConfigFile = errors.New("reading TOML config")
	ErrConfigEnv  = errors.New("reading ENV config")
)

type (
	ServiceName           string
	ServiceVersion        string
	CloserGracefulTimeout int
)

type DefaultConfig struct {
	ServiceName           ServiceName               `env:"SERVICE_NAME" toml:"service_name"`
	ServiceVersion        ServiceVersion            `env:"SERVICE_VERSION" toml:"service_version"`
	CloserGracefulTimeout libcloser.GracefulTimeout `env:"CLOSER_GRACEFUL_TIMEOUT" toml:"closer_graceful_timeout"`
	LogLevel              liblog.LogLevel           `env:"LOG_LEVEL" toml:"log_level"`
	Mode                  libmode.Mode              `env:"MODE" toml:"mode"`
}

// NewConfig creates configuration with the following precedence:
// Environment variables take precedence over TOML config values.
func NewConfig[T any](filePath string) (T, error) {
	var err error
	var cfgToml T
	var cfgEnv T
	var cfgMerged T

	if filePath != "" {
		cfgToml, err = loadFromTOML[T](filePath)
		if err != nil {
			return cfgToml, errors.Join(ErrConfigFile, err)
		}
	}

	cfgEnv, err = loadFromEnv[T]()
	if err != nil {
		return cfgEnv, errors.Join(ErrConfigEnv, err)
	}

	// Merge configurations, with environment variables taking precedence.
	// Only override if environment variables are set (non-zero values).
	cfgMerged = mergeConfigWithPriority(cfgToml, cfgEnv)

	return cfgMerged, nil
}

func loadFromEnv[T any]() (T, error) {
	var cfg T

	err := env.Parse(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return cfg, nil
}

func loadFromTOML[T any](filePath string) (T, error) {
	var cfg T

	data, err := os.ReadFile(filePath)
	if err != nil {
		return cfg, fmt.Errorf("failed to read TOML file: %w", err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse TOML file: %w", err)
	}

	return cfg, nil
}

func mergeConfigWithPriority[T any](base T, override T) T {
	result := base
	baseVal := reflect.ValueOf(base)
	overrideVal := reflect.ValueOf(override)
	resultVal := reflect.ValueOf(&result).Elem()

	// Extract concrete values if we have interface values
	if baseVal.Kind() == reflect.Interface && !baseVal.IsNil() {
		baseVal = baseVal.Elem()
	}
	if overrideVal.Kind() == reflect.Interface && !overrideVal.IsNil() {
		overrideVal = overrideVal.Elem()
	}

	if baseVal.Kind() != reflect.Struct {
		return base
	}

	// Iterate over struct fields. Compare and merge.
	for i := 0; i < baseVal.NumField(); i++ {
		fieldType := baseVal.Type().Field(i)
		if !fieldType.IsExported() {
			continue
		}

		overrideField := overrideVal.Field(i)
		resultField := resultVal.Field(i)

		// Skip zero-valued fields in override
		if isZeroValue(overrideField) {
			continue
		}

		// For all fields, just copy the override value if it's not zero
		resultField.Set(overrideField)
	}

	return result
}

// isZeroValue checks arg is of zero value of its type.
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0

	case reflect.Float32, reflect.Float64:
		return v.Float() == 0

	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == complex(0, 0)

	case reflect.String:
		return v.String() == ""

	case reflect.Interface, reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan:
		return v.IsNil()

	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isZeroValue(v.Index(i)) {
				return false
			}
		}
		return true

	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() && !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true

	default:
		return v.IsZero()
	}
}
