package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Field struct {
		Value1 string `mapstructure:"val1,omitempty"`
		Value2 string `mapstructure:"val2"`
	} `mapstructure:"fiel"`
}

func TestGettingEnvVars(t *testing.T) {
	c := testConfig{}
	acc := make([]string, 0)
	acc = getConfigFieldsAsEnvNames(c, "", "_", acc)
	require.Equal(t, acc, []string{"fiel_val1", "fiel_val2"})
}
