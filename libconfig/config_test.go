package libconfig_test

import (
	"os"
	"testing"

	"github.com/aliocode/golib/libconfig"
	"github.com/stretchr/testify/require"
)

func Test_NewConfig(t *testing.T) {
	type Config struct {
		libconfig.DefaultConfig
		Port string `env:"PORT" toml:"port"`
		IDs  []int  `env:"IDS" toml:"ids"`
	}

	t.Run("full env override", func(t *testing.T) {
		nameEnv := "name env"
		modeEnv := "DEV"
		os.Setenv("SERVICE_NAME", nameEnv)
		os.Setenv("MODE", modeEnv)
		os.Setenv("PORT", "55")
		os.Setenv("CLOSER_GRACEFUL_TIMEOUT", "6")
		os.Setenv("IDS", "7,8,9")
		defer os.Unsetenv("SERVICE_NAME")
		defer os.Unsetenv("MODE")
		defer os.Unsetenv("PORT")
		defer os.Unsetenv("CLOSER_GRACEFUL_TIMEOUT")
		defer os.Unsetenv("IDS")

		filePath := ""
		cfg, err := libconfig.NewConfig[Config](filePath)
		require.NoError(t, err)
		require.EqualValues(t, nameEnv, cfg.ServiceName)
		require.EqualValues(t, modeEnv, cfg.Mode)
		require.EqualValues(t, "55", cfg.Port)
		require.EqualValues(t, "6", cfg.DefaultConfig.CloserGracefulTimeout)
		require.EqualValues(t, []int{7, 8, 9}, cfg.IDs)
	},
	)

	t.Run("full toml base", func(t *testing.T) {
		filePath := "./config_test.toml"
		cfg, err := libconfig.NewConfig[Config](filePath)
		require.NoError(t, err)
		require.EqualValues(t, "name toml", cfg.ServiceName)
		require.EqualValues(t, "PROD", cfg.Mode)
		require.EqualValues(t, "99", cfg.Port)
		require.EqualValues(t, "3", cfg.DefaultConfig.CloserGracefulTimeout)
		require.EqualValues(t, []int{1, 2, 3}, cfg.IDs)
	},
	)
}
