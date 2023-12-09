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
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CharmingOption represents Functional Options Pattern.
// See this article for details -
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
// CharmingOption is a type for a function that accepts a pointer to
// an empty or minimal SnakeCharmer struct created in the constructor.
type CharmingOption func(*SnakeCharmer) error

// WithResultStruct sets the pointer to the struct that will contain
// the decoded config parameters.
// REQUIRED
// NOTE: the struct must be initialized with default values
// which will be used as flags
func WithResultStruct(rs interface{}) CharmingOption {
	timeType := reflect.TypeOf(time.Time{})
	v := reflect.ValueOf(rs)
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	} else {
		return func(sch *SnakeCharmer) error {
			return fmt.Errorf("result struct must be a pointer to a struct. Got <%s>", v.Type().String())
		}
	}

	if v.Kind() != reflect.Struct || v.Type().ConvertibleTo(timeType) {
		return func(sch *SnakeCharmer) error {
			return fmt.Errorf("result struct must be a pointer to a struct. Got <*%s>", v.Type().String())
		}
	}

	return func(sch *SnakeCharmer) error {
		sch.resultStruct = rs
		return nil
	}
}

// WithFieldTagName sets the tag name that snakecharmer reads for field names.
// This defaults to "mapstructure"
func WithFieldTagName(s string) CharmingOption {
	tag := strings.TrimSpace(s)
	if len(tag) == 0 {
		return func(sch *SnakeCharmer) error {
			return fmt.Errorf("invalid field tag name: %q", s)
		}
	}
	return func(sch *SnakeCharmer) error {
		sch.fieldTagName = tag
		return nil
	}
}

// WithEnvTagName sets the tag name that snakecharmer reads for setting ENV var name.
// This defaults to "env"
func WithEnvTagName(s string) CharmingOption {
	tag := strings.TrimSpace(s)
	if len(tag) == 0 {
		return func(sch *SnakeCharmer) error {
			return fmt.Errorf("invalid envTagName: %q", s)
		}
	}
	return func(sch *SnakeCharmer) error {
		sch.envTagName = tag
		return nil
	}
}

// WithFlagHelpTagName sets the tag name that snakecharmer reads for flag usage help.
// This defaults to "usage"
func WithFlagHelpTagName(s string) CharmingOption {
	tag := strings.TrimSpace(s)
	if len(tag) == 0 {
		return func(sch *SnakeCharmer) error {
			return fmt.Errorf("invalid flag help tag name: %q", s)
		}
	}
	return func(sch *SnakeCharmer) error {
		sch.flagHelpTagName = tag
		return nil
	}
}

// WithConfigFileType sets the type that will be passed to viper.SetConfigType().
// REQUIRED in case if the config file does not have the extension or
// if the config file extension is not in the list of supported extensions.
// See viper.SupportedExts for full list of supported extensions.
// This defaults to "yaml"
func WithConfigFileType(s string) CharmingOption {
	ext := strings.TrimSpace(s)
	if !fileExtSupported(ext) {
		return func(sch *SnakeCharmer) error {
			return fmt.Errorf("invalid config file type: %q", s)
		}
	}
	return func(sch *SnakeCharmer) error {
		sch.configFileType = ext
		return nil
	}
}

// WithConfigFilePath sets the config file path that will be passed to
// viper.AddConfigPath() if path is a directory,
// or to viper.SetConfigFile() if path is a file.
// This defaults to "", which means config file won't be used.
func WithConfigFilePath(s string) CharmingOption {
	return func(sch *SnakeCharmer) error {
		sch.configFilePath = strings.TrimSpace(s)
		return nil
	}
}

// WithConfigFileBaseName sets the base name of the config file (without extension)
// that will be passed to viper.SetConfigName().
// REQUIRED in case of the config file path is a directory, otherwise ignored.
// This defaults to "config"
func WithConfigFileBaseName(s string) CharmingOption {
	name := strings.TrimSpace(s)
	if len(name) == 0 {
		return func(sch *SnakeCharmer) error {
			return fmt.Errorf("invalid config file base name: %q", s)
		}
	}
	return func(sch *SnakeCharmer) error {
		sch.configFileBaseName = name
		return nil
	}
}

// WithDecoderConfigOption adds a viper.DecoderConfigOption that will be passed
// to viper.Unmarshal for configuring mapstructure.DecoderConfig options
// See https://pkg.go.dev/github.com/spf13/viper@v1.17.0#DecoderConfigOption
func WithDecoderConfigOption(opt viper.DecoderConfigOption) CharmingOption {
	return func(sch *SnakeCharmer) error {
		if sch.decoderConfigOptions == nil {
			sch.decoderConfigOptions = []viper.DecoderConfigOption{}
		}
		sch.decoderConfigOptions = append(sch.decoderConfigOptions, opt)
		return nil
	}
}

// WithIgnoreUntaggedFields allows to ignore all struct fields without
// explicit fieldTagName, comparable to `mapstructure:"-"` as default behaviour.
// See WithFieldTagName
func WithIgnoreUntaggedFields(on bool) CharmingOption {
	return func(sch *SnakeCharmer) error {
		sch.ignoreUntaggedFields = on
		if sch.decoderConfigOptions == nil {
			sch.decoderConfigOptions = []viper.DecoderConfigOption{}
		}
		sch.decoderConfigOptions = append(sch.decoderConfigOptions,
			func(dc *mapstructure.DecoderConfig) { dc.IgnoreUntaggedFields = on },
		)
		return nil
	}
}

// WithViper sets the pointer to the viper.Viper instance
// This defaults to viper.New()
func WithViper(viper *viper.Viper) CharmingOption {
	return func(sch *SnakeCharmer) error {
		sch.viper = viper
		return nil
	}
}

// WithCobraCommand sets the pointer to the cobra.Command instance
// REQUIRED
func WithCobraCommand(cmd *cobra.Command) CharmingOption {
	return func(sch *SnakeCharmer) error {
		sch.cmd = cmd
		return nil
	}
}
