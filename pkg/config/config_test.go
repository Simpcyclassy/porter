package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/experimental"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_GetHomeDir(t *testing.T) {
	c := NewTestConfig(t)

	home, err := c.GetHomeDir()
	require.NoError(t, err)

	assert.Equal(t, "/home/myuser/.porter", home)
}

func TestConfig_GetHomeDirFromSymlink(t *testing.T) {
	c := NewTestConfig(t)

	// Set up no PORTER_HOME, and /usr/local/bin/porter -> ~/.porter/porter
	c.Unsetenv(EnvHOME)
	getExecutable = func() (string, error) {
		return "/usr/local/bin/porter", nil
	}
	evalSymlinks = func(path string) (string, error) {
		return "/home/myuser/.porter/porter", nil
	}

	home, err := c.GetHomeDir()
	require.NoError(t, err)

	// The reason why we do filepath.join here and not above is because resolving symlinks gets the OS involved
	// and on Windows, that means flipping the afero `/` to `\`.
	assert.Equal(t, filepath.Join("/home/myuser", ".porter"), home)
}

func TestConfig_GetFeatureFlags(t *testing.T) {
	t.Parallel()

	t.Run("structured logs defaulted to disabled", func(t *testing.T) {
		c := Config{}
		assert.False(t, c.IsFeatureEnabled(experimental.FlagStructuredLogs))
	})

	t.Run("structured logs enabled", func(t *testing.T) {
		c := Config{}
		c.Data.ExperimentalFlags = []string{experimental.StructuredLogs}
		assert.True(t, c.IsFeatureEnabled(experimental.FlagStructuredLogs))
	})
}

func TestConfigExperimentalFlags(t *testing.T) {
	// Do not run in parallel since we are using os.Setenv

	t.Run("default off", func(t *testing.T) {
		c := NewTestConfig(t)
		assert.False(t, c.IsFeatureEnabled(experimental.FlagStructuredLogs))
	})

	t.Run("user configuration", func(t *testing.T) {
		os.Setenv("PORTER_EXPERIMENTAL", experimental.StructuredLogs)
		defer os.Unsetenv("PORTER_EXPERIMENTAL")

		c := New()
		require.NoError(t, c.Load(context.Background(), nil), "Load failed")
		assert.True(t, c.IsFeatureEnabled(experimental.FlagStructuredLogs))
	})

	t.Run("programmatically enabled", func(t *testing.T) {
		c := NewTestConfig(t)
		c.Data.ExperimentalFlags = nil
		c.SetExperimentalFlags(experimental.FlagStructuredLogs)
		assert.True(t, c.IsFeatureEnabled(experimental.FlagStructuredLogs))
	})
}

func TestConfig_GetBuildDriver(t *testing.T) {
	c := NewTestConfig(t)
	c.Data.BuildDriver = "special"
	require.Equal(t, BuildDriverBuildkit, c.GetBuildDriver(), "Default to docker when experimental is false, even when a build driver is set")

	c.SetExperimentalFlags(experimental.FlagStructuredLogs)
	require.Equal(t, BuildDriverBuildkit, c.GetBuildDriver(), "Use the specified driver when the build driver feature is enabled")
}
