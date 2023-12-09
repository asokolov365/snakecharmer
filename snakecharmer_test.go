// Copyright 2013-2023 The SnakeCharmer Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package snakecharmer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

type testLogLimitsConfig struct {
	WarnsLimit  *uint `snakecharmer:"warn,omitempty" env:"TEST_LOG_LIMIT_WARN" usage:"Limit warn messages per sec"`
	ErrorsLimit *uint `snakecharmer:"error,omitempty" env:"TEST_LOG_LIMIT_ERROR" usage:"Limit error messages per sec"`
	ignoreMe    bool
}

type testLoggingConfig struct {
	Level           *string              `snakecharmer:"level,omitempty" env:"TEST_LOG_LEVEL" usage:"Log level"`
	LogJSON         *bool                `snakecharmer:"json,omitempty" env:"TEST_LOG_JSON" usage:"Log in JSON format"`
	LogLimits       *testLogLimitsConfig `snakecharmer:"limit,omitempty"`
	LogDestinations *map[string]string   `snakecharmer:"dst,omitempty" usage:"Log to multiple destinations"`
	ignoreMe        bool
}

type testStruct struct {
	Workers      *int               `snakecharmer:"workers,omitempty" env:"TEST_WORKERS" usage:"Number of workers to run"`
	MaxBurst     *float64           `snakecharmer:"max-burst,omitempty" env:"TEST_MAX_BURST" usage:"Max burst allowed, e.g 1.25"`
	BindAddr     *string            `snakecharmer:"bind-addr,omitempty" env:"TEST_BIND_ADDR" usage:"Addr to bind"`
	UpstreamURls *[]string          `snakecharmer:"upstreams,omitempty" usage:"List of upstream urls"`
	Logging      *testLoggingConfig `snakecharmer:"log,omitempty"`
	ignoreMe     bool
}

var (
	defaultWorkers      = 128
	defaultMaxBurst     = 1.25
	defaultBindAddr     = "0.0.0.0"
	defaultUpstreamURls = []string{
		"http://www.foo.com/",
		"http://www.bar.com/",
	}
	defaultLogLevel             = "info"
	defaultLogJSON              = false
	defaultLogWarnsLimit   uint = 100
	defaultLogErrorsLimit  uint = 100
	defaultLogDestinations      = map[string]string{
		"error": "/var/log/error.log",
		"debug": "/var/log/debug.log",
	}
	expectedFlagNames = []string{
		"workers",
		"max-burst",
		"bind-addr",
		"upstreams",
		"log.level",
		"log.json",
		"log.dst",
		"log.limit.warn",
		"log.limit.error",
	}
)

func initTestStruct() *testStruct {
	return &testStruct{
		ignoreMe:     true,
		Workers:      &defaultWorkers,
		MaxBurst:     &defaultMaxBurst,
		BindAddr:     &defaultBindAddr,
		UpstreamURls: &defaultUpstreamURls,
		Logging: &testLoggingConfig{
			ignoreMe: true,
			Level:    &defaultLogLevel,
			LogJSON:  &defaultLogJSON,
			LogLimits: &testLogLimitsConfig{
				ignoreMe:    true,
				WarnsLimit:  &defaultLogWarnsLimit,
				ErrorsLimit: &defaultLogErrorsLimit,
			},
			LogDestinations: &defaultLogDestinations,
		},
	}
}

func testFlagsDefined(cmd *cobra.Command, vpr *viper.Viper) error {
	for _, flagName := range expectedFlagNames {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			return fmt.Errorf("flag %q is not defined in Cobra command", flagName)
		}
		value := vpr.Get(flagName)
		if value == nil {
			return fmt.Errorf("value of flag %q is not set in Viper settings", flagName)
		}
	}
	return nil
}

func passConfigFlag(sch *SnakeCharmer, s string) (string, error) {
	if err := sch.Set(WithConfigFilePath(s)); err != nil {
		return "", fmt.Errorf("unexpected error in charmer.Set(): %s", err.Error())
	}
	return fmt.Sprintf("--config=%s", s), nil
}

func Test_NewSnakeCharmerError(t *testing.T) {
	f := func(opts ...CharmingOption) {
		t.Helper()
		_, err := NewSnakeCharmer(opts...)
		if err == nil {
			t.Fatalf("expecting non-nil error in NewSnakeCharmer()")
		}
		// fmt.Println(err.Error())
	}
	s := "snake charmer"
	cmd := &cobra.Command{}
	result := initTestStruct()

	f()
	f(WithResultStruct(s))
	f(WithResultStruct(&s))
	f(WithResultStruct(&time.Time{}))
	f(WithResultStruct(*result))
	f(WithResultStruct(result))
	f(
		WithResultStruct(result),
		WithCobraCommand(cmd),
		WithFieldTagName(" "),
	)
	f(
		WithResultStruct(result),
		WithCobraCommand(cmd),
		WithEnvTagName(" "),
	)
	f(
		WithResultStruct(result),
		WithCobraCommand(cmd),
		WithFlagHelpTagName(" "),
	)
	f(
		WithResultStruct(result),
		WithCobraCommand(cmd),
		WithConfigFileType(" "),
	)
	f(
		WithResultStruct(result),
		WithCobraCommand(cmd),
		WithConfigFileType("jpg"),
	)
	f(
		WithResultStruct(result),
		WithCobraCommand(cmd),
		WithConfigFileBaseName(" "),
	)
}

func Test_NewSnakeCharmerOkay(t *testing.T) {
	f := func(opts ...CharmingOption) {
		t.Helper()
		_, err := NewSnakeCharmer(opts...)
		if err != nil {
			t.Fatalf("unexpected error in NewSnakeCharmer(opts ...Option): %s", err.Error())
		}
	}
	cmd := &cobra.Command{}
	result := initTestStruct()

	f(
		WithResultStruct(result),
		WithCobraCommand(cmd),
	)
	f(
		WithResultStruct(result),
		WithCobraCommand(cmd),
		WithFieldTagName("snakecharmer"),
	)
	f(
		WithResultStruct(result),
		WithCobraCommand(cmd),
		WithEnvTagName("environment"),
	)
	f(
		WithResultStruct(result),
		WithCobraCommand(cmd),
		WithFlagHelpTagName("help"),
	)
	f(
		WithResultStruct(result),
		WithCobraCommand(cmd),
		WithConfigFileType("json"),
	)
	f(
		WithResultStruct(result),
		WithCobraCommand(cmd),
		WithConfigFilePath("/etc/snakecharmer"),
	)
	f(
		WithResultStruct(result),
		WithCobraCommand(cmd),
		WithConfigFileBaseName("conf"),
	)
	f(
		WithResultStruct(result),
		WithCobraCommand(cmd),
		WithIgnoreUntaggedFields(true),
	)
	f(
		WithResultStruct(result),
		WithCobraCommand(cmd),
		WithDecoderConfigOption(func(dc *mapstructure.DecoderConfig) { dc.WeaklyTypedInput = true }),
	)
}

func Test_SetError(t *testing.T) {
	cmd := &cobra.Command{}
	result := initTestStruct()
	charmer, err := NewSnakeCharmer(
		WithResultStruct(result),
		WithCobraCommand(cmd),
	)
	if err != nil {
		t.Fatalf("unexpected error in NewSnakeCharmer(): %s", err.Error())
	}
	f := func(sch *SnakeCharmer, option CharmingOption) {
		t.Helper()
		err := sch.Set(option)
		if err == nil {
			t.Fatalf("expecting non-nil error in Set(opts ...Option)")
		}
	}
	f(charmer, WithResultStruct(""))
	f(charmer, WithFieldTagName("    "))
	f(charmer, WithEnvTagName(""))
	f(charmer, WithConfigFileType("xml"))
}

func Test_configFilePathError(t *testing.T) {
	cmd := &cobra.Command{}
	result := initTestStruct()
	charmer, err := NewSnakeCharmer(
		WithResultStruct(result),
		WithCobraCommand(cmd),
	)
	if err != nil {
		t.Fatalf("unexpected error in NewSnakeCharmer(opts ...Option): %s", err.Error())
	}
	_, err = charmer.findConfigFile()
	if err == nil {
		t.Fatalf("expecting non-nil error in findConfigFile()")
	}
	err = charmer.mergeInConfigFile()
	if err == nil {
		t.Fatalf("expecting non-nil error in mergeInConfigFile()")
	}

	err = charmer.Set(WithConfigFilePath("./nonexiting"))
	if err != nil {
		t.Fatalf("unexpected error in (*SnakeCharmer).Set(opts ...Option): %s", err.Error())
	}
	_, err = charmer.findConfigFile()
	if err == nil {
		t.Fatalf("expecting non-nil error in (*SnakeCharmer).findConfigFile()")
	}
	err = charmer.mergeInConfigFile()
	if err == nil {
		t.Fatalf("expecting non-nil error in (*SnakeCharmer).mergeInConfigFile()")
	}
}

func Test_SetGetOkay(t *testing.T) {
	cmd := &cobra.Command{}
	result := initTestStruct()
	charmer, err := NewSnakeCharmer(
		WithResultStruct(result),
		WithCobraCommand(cmd),
	)
	if err != nil {
		t.Fatalf("unexpected error in NewSnakeCharmer(opts ...Option): %s", err.Error())
	}
	err = charmer.Set(
		WithFieldTagName("snakecharmer"),
		WithEnvTagName("environment"),
		WithFlagHelpTagName("help"),
		WithConfigFileType("json"),
		WithConfigFilePath("/etc/snakecharmer"),
		WithConfigFileBaseName("conf"),
		WithDecoderConfigOption(
			func(dc *mapstructure.DecoderConfig) { dc.WeaklyTypedInput = true },
		),
		WithIgnoreUntaggedFields(true),
	)
	if err != nil {
		t.Fatalf("unexpected error in Set(opts ...Option): %s", err.Error())
	}
	require.Equal(t, "snakecharmer", charmer.FieldTagName())
	require.Equal(t, "environment", charmer.EnvTagName())
	require.Equal(t, "help", charmer.FlagHelpTagName())
	require.Equal(t, "json", charmer.ConfigFileType())
	require.Equal(t, "/etc/snakecharmer", charmer.ConfigFilePath())
	require.Equal(t, "conf", charmer.ConfigFileBaseName())
	require.Equal(t, result, charmer.ResultStruct())
	require.Equal(t, true, charmer.IgnoreUntaggedFields())
	require.Equal(t, 2, len(charmer.DecoderConfigOptions()))
}

func Test_WithoutConfigFile(t *testing.T) {
	var charmer *SnakeCharmer
	var err error

	result := initTestStruct()
	vpr := viper.New()
	cmd := &cobra.Command{
		PreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if err = charmer.UnmarshalExact(); err != nil {
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Happy programming!")
		},
	}

	charmer, err = NewSnakeCharmer(
		WithResultStruct(result),
		WithFieldTagName("snakecharmer"),
		WithViper(vpr),
		WithCobraCommand(cmd),
		WithIgnoreUntaggedFields(true),
	)
	if err != nil {
		t.Fatalf("unexpected error in NewSnakeCharmer(opts ...Option): %s", err.Error())
	}

	charmer.AddFlags()

	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("unexpected error in (*cobra.Command).ParseFlags(): %s", err.Error())
	}

	// cmd.DebugFlags()
	// fmt.Printf("viper.AllSettings() = %v\n", viper.AllSettings())

	if err = testFlagsDefined(cmd, vpr); err != nil {
		t.Fatalf(err.Error())
	}

	var (
		expectedWorkers  = 512             // this will be set via env TEST_WORKERS=512
		expectedMaxBurst = 1.5             // this will be set via flag --max-burst=1.5
		expectedBindAddr = defaultBindAddr // this will be the unchaged
	)

	// Pretend a user sets TEST_WORKERS=512 env var
	os.Setenv("TEST_WORKERS", fmt.Sprintf("%d", expectedWorkers))

	// Pretend a user set a few flags
	cmd.SetArgs([]string{
		fmt.Sprintf("--max-burst=%f", expectedMaxBurst),
	})

	if err = cmd.Execute(); err != nil {
		t.Fatalf("unexpected error in cmd.Execute(): %s", err.Error())
	}

	// Checking result one by one
	require.Equal(t, expectedWorkers, *result.Workers)
	require.Equal(t, expectedMaxBurst, *result.MaxBurst)
	require.Equal(t, expectedBindAddr, *result.BindAddr)

	require.Equal(t, defaultUpstreamURls, *result.UpstreamURls)
	require.Equal(t, defaultLogLevel, *result.Logging.Level)
	require.Equal(t, defaultLogJSON, *result.Logging.LogJSON)
	require.Equal(t, defaultLogWarnsLimit, *result.Logging.LogLimits.WarnsLimit)
	require.Equal(t, defaultLogErrorsLimit, *result.Logging.LogLimits.ErrorsLimit)
	require.Equal(t, defaultLogDestinations, *result.Logging.LogDestinations)

}

func Test_WithDefaultConfigFile(t *testing.T) {
	var charmer *SnakeCharmer
	var err error

	result := initTestStruct()
	vpr := viper.New()
	cmd := &cobra.Command{
		PreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if err = charmer.UnmarshalExact(); err != nil {
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// fmt.Printf("all settings: %+v\n", vpr.AllSettings())
			fmt.Println("Happy programming!")
		},
	}
	defaultConfigFile := "./test-config"

	charmer, err = NewSnakeCharmer(
		WithResultStruct(result),
		WithFieldTagName("snakecharmer"),
		WithViper(vpr),
		WithCobraCommand(cmd),
		WithConfigFilePath(defaultConfigFile),
		WithConfigFileType("yaml"),
		WithIgnoreUntaggedFields(true),
		WithDecoderConfigOption(
			func(dc *mapstructure.DecoderConfig) { dc.WeaklyTypedInput = true },
		),
	)
	if err != nil {
		t.Fatalf("unexpected error in NewSnakeCharmer(opts ...Option): %s", err.Error())
	}

	charmer.AddFlags()

	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("unexpected error in (*cobra.Command).ParseFlags(): %s", err.Error())
	}

	if err = testFlagsDefined(cmd, vpr); err != nil {
		t.Fatalf(err.Error())
	}

	var (
		expectedConfigFile   = defaultConfigFile
		expectedWorkers      = 512
		expectedMaxBurst     = 1.5
		expectedBindAddr     = "127.0.0.255"
		expectedUpstreamURls = []string{
			"http://www.example1.com/",
			"http://www.example2.com/",
			"http://www.example3.com/",
		}
		expectedLogLevel             = "debug"
		expectedLogJSON              = true
		expectedLogWarnsLimit   uint = 20
		expectedLogErrorsLimit  uint = 5
		expectedLogDestinations      = map[string]string{
			"error": "/var/log/error.log",
			"debug": "/var/log/debug.log",
		}
	)

	// Pretend a user sets TEST_WORKERS=512 env var
	os.Setenv("TEST_WORKERS", fmt.Sprintf("%d", expectedWorkers))
	// Pretend a user sets TEST_LOG_LEVEL=debug env var
	os.Setenv("TEST_LOG_LEVEL", expectedLogLevel)

	// Pretend a user set a few flags
	cmd.SetArgs([]string{
		fmt.Sprintf("--log.limit.error=%d", expectedLogErrorsLimit),
	})

	if err = cmd.Execute(); err != nil {
		t.Fatalf("unexpected error in (*cobra.Command).Execute(): %s", err.Error())
	}

	require.Equal(t, expectedConfigFile, vpr.ConfigFileUsed())
	require.Equal(t, expectedWorkers, *result.Workers)
	require.Equal(t, expectedMaxBurst, *result.MaxBurst)
	require.Equal(t, expectedBindAddr, *result.BindAddr)
	require.Equal(t, expectedUpstreamURls, *result.UpstreamURls)
	require.Equal(t, expectedLogLevel, *result.Logging.Level)
	require.Equal(t, expectedLogJSON, *result.Logging.LogJSON)
	require.Equal(t, expectedLogWarnsLimit, *result.Logging.LogLimits.WarnsLimit)
	require.Equal(t, expectedLogErrorsLimit, *result.Logging.LogLimits.ErrorsLimit)
	require.Equal(t, expectedLogDestinations, *result.Logging.LogDestinations)
}

func Test_WithConfigFileUnknownExtension(t *testing.T) {
	var charmer *SnakeCharmer
	var err error

	result := initTestStruct()
	vpr := viper.New()
	cmd := &cobra.Command{
		PreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if err = charmer.UnmarshalExact(); err != nil {
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// fmt.Printf("all settings: %+v\n", vpr.AllSettings())
			fmt.Println("Happy programming!")
		},
	}
	defaultConfigFile := "./test.conf"

	charmer, err = NewSnakeCharmer(
		WithResultStruct(result),
		WithFieldTagName("snakecharmer"),
		WithViper(vpr),
		WithCobraCommand(cmd),
		WithConfigFilePath(defaultConfigFile),
		WithConfigFileType("yaml"),
		WithIgnoreUntaggedFields(true),
		WithDecoderConfigOption(
			func(dc *mapstructure.DecoderConfig) { dc.WeaklyTypedInput = true },
		),
	)
	if err != nil {
		t.Fatalf("unexpected error in NewSnakeCharmer(opts ...Option): %s", err.Error())
	}

	charmer.AddFlags()

	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("unexpected error in (*cobra.Command).ParseFlags(): %s", err.Error())
	}

	if err = testFlagsDefined(cmd, vpr); err != nil {
		t.Fatalf(err.Error())
	}

	var (
		expectedConfigFile   = defaultConfigFile
		expectedWorkers      = 512
		expectedMaxBurst     = 1.5
		expectedBindAddr     = "127.0.0.255"
		expectedUpstreamURls = []string{
			"http://www.example1.com/",
			"http://www.example2.com/",
			"http://www.example3.com/",
		}
		expectedLogLevel             = "debug"
		expectedLogJSON              = true
		expectedLogWarnsLimit   uint = 20
		expectedLogErrorsLimit  uint = 5
		expectedLogDestinations      = map[string]string{
			"error": "/var/log/error.log",
			"debug": "/var/log/debug.log",
		}
	)

	// Pretend a user sets TEST_WORKERS=512 env var
	os.Setenv("TEST_WORKERS", fmt.Sprintf("%d", expectedWorkers))
	// Pretend a user sets TEST_LOG_LEVEL=debug env var
	os.Setenv("TEST_LOG_LEVEL", expectedLogLevel)

	// Pretend a user set a few flags
	cmd.SetArgs([]string{
		fmt.Sprintf("--log.limit.error=%d", expectedLogErrorsLimit),
	})

	if err = cmd.Execute(); err != nil {
		t.Fatalf("unexpected error in (*cobra.Command).Execute(): %s", err.Error())
	}

	require.Equal(t, expectedConfigFile, vpr.ConfigFileUsed())
	require.Equal(t, expectedWorkers, *result.Workers)
	require.Equal(t, expectedMaxBurst, *result.MaxBurst)
	require.Equal(t, expectedBindAddr, *result.BindAddr)
	require.Equal(t, expectedUpstreamURls, *result.UpstreamURls)
	require.Equal(t, expectedLogLevel, *result.Logging.Level)
	require.Equal(t, expectedLogJSON, *result.Logging.LogJSON)
	require.Equal(t, expectedLogWarnsLimit, *result.Logging.LogLimits.WarnsLimit)
	require.Equal(t, expectedLogErrorsLimit, *result.Logging.LogLimits.ErrorsLimit)
	require.Equal(t, expectedLogDestinations, *result.Logging.LogDestinations)
}

func Test_WithConfigFileArg(t *testing.T) {
	var charmer *SnakeCharmer
	var err error

	result := initTestStruct()
	vpr := viper.New()
	cmd := &cobra.Command{
		PreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if err = charmer.UnmarshalExact(); err != nil {
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// fmt.Printf("config file used: %s\n", vpr.ConfigFileUsed())
			// fmt.Printf("all settings: %+v\n", vpr.AllSettings())
			fmt.Println("Happy programming!")
		},
	}
	defaultConfigFile := ""
	cmd.PersistentFlags().StringVarP(&defaultConfigFile, "config", "c", defaultConfigFile, "Path to a config file")

	charmer, err = NewSnakeCharmer(
		WithResultStruct(result),
		WithFieldTagName("snakecharmer"),
		WithViper(vpr),
		WithCobraCommand(cmd),
		WithConfigFilePath(defaultConfigFile),
		WithIgnoreUntaggedFields(true),
	)
	if err != nil {
		t.Fatalf("unexpected error in NewSnakeCharmer(opts ...Option): %s", err.Error())
	}

	charmer.AddFlags()

	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("unexpected error in (*cobra.Command).ParseFlags(): %s", err.Error())
	}

	if err = testFlagsDefined(cmd, vpr); err != nil {
		t.Fatalf(err.Error())
	}

	var (
		expectedConfigFile   = "./test-config.json"
		expectedWorkers      = 512
		expectedMaxBurst     = 1.5
		expectedBindAddr     = "127.0.0.1"
		expectedUpstreamURls = []string{
			"http://www.example1.com/",
			"http://www.example2.com/",
			"http://www.example3.com/",
		}
		expectedLogLevel             = "debug"
		expectedLogJSON              = true
		expectedLogWarnsLimit   uint = 5
		expectedLogErrorsLimit  uint = 10
		expectedLogDestinations      = map[string]string{
			"error": "/var/log/error.log",
			"debug": "/var/log/debug.log",
		}
	)

	// Pretend a user sets TEST_WORKERS=512 env var
	os.Setenv("TEST_WORKERS", fmt.Sprintf("%d", expectedWorkers))
	// Pretend a user sets TEST_LOG_LEVEL=debug env var
	os.Setenv("TEST_LOG_LEVEL", expectedLogLevel)

	confFlag, err := passConfigFlag(charmer, expectedConfigFile)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Pretend a user set a few flags
	cmd.SetArgs([]string{
		confFlag,
		"--log.json",
		fmt.Sprintf("--log.limit.warn=%d", expectedLogWarnsLimit),
		fmt.Sprintf("--bind-addr=%s", expectedBindAddr),
	})

	if err = cmd.Execute(); err != nil {
		t.Fatalf("unexpected error in cmd.Execute(): %s", err.Error())
	}

	require.Equal(t, expectedConfigFile, vpr.ConfigFileUsed())
	require.Equal(t, expectedWorkers, *result.Workers)
	require.Equal(t, expectedMaxBurst, *result.MaxBurst)
	require.Equal(t, expectedBindAddr, *result.BindAddr)
	require.Equal(t, expectedUpstreamURls, *result.UpstreamURls)
	require.Equal(t, expectedLogLevel, *result.Logging.Level)
	require.Equal(t, expectedLogJSON, *result.Logging.LogJSON)
	require.Equal(t, expectedLogWarnsLimit, *result.Logging.LogLimits.WarnsLimit)
	require.Equal(t, expectedLogErrorsLimit, *result.Logging.LogLimits.ErrorsLimit)
	require.Equal(t, expectedLogDestinations, *result.Logging.LogDestinations)
}

func Test_WithConfigDir(t *testing.T) {
	var charmer *SnakeCharmer
	var err error

	result := initTestStruct()
	vpr := viper.New()
	cmd := &cobra.Command{
		PreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if err = charmer.UnmarshalExact(); err != nil {
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// fmt.Printf("config file used: %s\n", vpr.ConfigFileUsed())
			// fmt.Printf("all settings: %+v\n", vpr.AllSettings())
			fmt.Println("Happy programming!")
		},
	}
	defaultConfigDir := ""
	cmd.PersistentFlags().StringVarP(&defaultConfigDir, "config", "c", defaultConfigDir, "Path to a config dir")

	charmer, err = NewSnakeCharmer(
		WithResultStruct(result),
		WithFieldTagName("snakecharmer"),
		WithViper(vpr),
		WithCobraCommand(cmd),
		WithConfigFilePath(defaultConfigDir),
		WithConfigFileBaseName("test-config"),
		WithIgnoreUntaggedFields(true),
	)
	if err != nil {
		t.Fatalf("unexpected error in NewSnakeCharmer(opts ...Option): %s", err.Error())
	}

	charmer.AddFlags()

	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("unexpected error in (*cobra.Command).ParseFlags(): %s", err.Error())
	}

	if err = testFlagsDefined(cmd, vpr); err != nil {
		t.Fatalf(err.Error())
	}

	var (
		expectedConfigFile   = "./test-config.json"
		expectedWorkers      = 512
		expectedMaxBurst     = 1.5
		expectedBindAddr     = "127.0.0.1"
		expectedUpstreamURls = []string{
			"http://www.example1.com/",
			"http://www.example2.com/",
			"http://www.example3.com/",
		}
		expectedLogLevel             = "debug"
		expectedLogJSON              = true
		expectedLogWarnsLimit   uint = 5
		expectedLogErrorsLimit  uint = 10
		expectedLogDestinations      = map[string]string{
			"error": "/var/log/error.log",
			"debug": "/var/log/debug.log",
		}
	)

	// Pretend a user sets TEST_WORKERS=512 env var
	os.Setenv("TEST_WORKERS", fmt.Sprintf("%d", expectedWorkers))
	// Pretend a user sets TEST_LOG_LEVEL=debug env var
	os.Setenv("TEST_LOG_LEVEL", expectedLogLevel)

	confFlag, err := passConfigFlag(charmer, "./")
	if err != nil {
		t.Fatalf(err.Error())
	}
	// Pretend a user set a few flags
	cmd.SetArgs([]string{
		confFlag,
		"--log.json",
		fmt.Sprintf("--log.limit.warn=%d", expectedLogWarnsLimit),
		fmt.Sprintf("--bind-addr=%s", expectedBindAddr),
	})

	if err = cmd.Execute(); err != nil {
		t.Fatalf("unexpected error in cmd.Execute(): %s", err.Error())
	}

	expectedConfigFileAbsPath, err := filepath.Abs(expectedConfigFile)
	if err != nil {
		t.Fatalf("unexpected error in filepath.Abs(%q): %s", expectedConfigFile, err.Error())
	}
	require.Equal(t, expectedConfigFileAbsPath, vpr.ConfigFileUsed())
	require.Equal(t, expectedWorkers, *result.Workers)
	require.Equal(t, expectedMaxBurst, *result.MaxBurst)
	require.Equal(t, expectedBindAddr, *result.BindAddr)
	require.Equal(t, expectedUpstreamURls, *result.UpstreamURls)
	require.Equal(t, expectedLogLevel, *result.Logging.Level)
	require.Equal(t, expectedLogJSON, *result.Logging.LogJSON)
	require.Equal(t, expectedLogWarnsLimit, *result.Logging.LogLimits.WarnsLimit)
	require.Equal(t, expectedLogErrorsLimit, *result.Logging.LogLimits.ErrorsLimit)
	require.Equal(t, expectedLogDestinations, *result.Logging.LogDestinations)
}
