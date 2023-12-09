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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewSnakeCharmer creates a new snakecharmer instance.
// charmer, err = NewSnakeCharmer(
//
//	WithResultStruct(result),
//	WithFieldTagName("snakecharmer"),
//	WithViper(vpr),
//	WithCobraCommand(cmd),
//	WithConfigFilePath(defaultConfigFile),
//	WithConfigFileType("yaml"),
//	WithIgnoreUntaggedFields(true),
//	WithDecoderConfigOption(
//		func(dc *mapstructure.DecoderConfig) { dc.WeaklyTypedInput = true },
//	),
//
// )
func NewSnakeCharmer(opts ...CharmingOption) (*SnakeCharmer, error) {
	sch := SnakeCharmer{
		fieldTagName:       "mapstructure",
		envTagName:         "env",
		flagHelpTagName:    "usage",
		configFileType:     "yaml",
		configFilePath:     "",
		configFileBaseName: "config",
	}

	for _, opt := range opts {
		if err := opt(&sch); err != nil {
			return &sch, err
		}
	}

	if sch.resultStruct == nil {
		return &sch, fmt.Errorf("result struct <interface{}> is not set")
	}
	if sch.viper == nil {
		sch.viper = viper.New()
	}
	if sch.cmd == nil {
		return &sch, fmt.Errorf("cmd <*cobra.Command> is not set")
	}

	return &sch, nil
}

// SnakeCharmer helps to get Cobra and Viper work together.
// It uses a user defined Struct for reading field tags and default values.
// Please note, that the struct must be initialized (with the default values)
// before passing to the SnakeCharmer.
// It automatically creates flags and adds them to cobra PersistentFlags flagset.
// It also creates viper's config params, and sets their default values,
// binds viper's config param with a corresponding flag from the cobra flagset,
// binds viper's config param with a corresponding ENV var
// SnakeCharmer sets the following priority of values:
// 1. flags (if passed)
// 2. ENV variables (if env tag set)
// 3. config file (if used)
// 4. defaults (from user defined Struct)
type SnakeCharmer struct {
	// resultStruct is a pointer to the struct that will contain
	// the decoded values.
	resultStruct interface{}

	// The tag name that snakecharmer reads for field names.
	// This defaults to "mapstructure"
	fieldTagName string

	// The tag name that snakecharmer reads for setting ENV var name.
	// This defaults to "env"
	envTagName string

	// The tag name that snakecharmer reads for flag usage help.
	// This defaults to "usage"
	flagHelpTagName string

	// The type that will be passed to viper.SetConfigType().
	// REQUIRED in case if the config file does not have the extension or
	// if the config file extension is not in the list of supported extensions.
	// See viper.SupportedExts for full list of supported extensions.
	// This defaults to "yaml"
	configFileType string

	// The config file path that will be passed to
	// viper.AddConfigPath() if path is a directory,
	// viper.SetConfigFile() if path is a file.
	// This defaults to "", which means config file won't be used.
	configFilePath string

	// The base name of the config file (without extension)
	// that will be passed to viper.SetConfigName().
	// REQUIRED in case of the configFilePath is a directory, otherwise ignored.
	// This defaults to "config"
	configFileBaseName string

	// The pointer to the viper.Viper instance
	viper *viper.Viper

	// The pointer to the cobra.Command instance
	cmd *cobra.Command

	// A slice of viper.DecoderConfigOption that will be passed to viper.Unmarshal
	// to configure mapstructure.DecoderConfig options
	// See https://pkg.go.dev/github.com/spf13/viper@v1.17.0#DecoderConfigOption
	decoderConfigOptions []viper.DecoderConfigOption

	// ignoreUntaggedFields ignores all struct fields without explicit
	// fieldTagName, comparable to `mapstructure:"-"` as default behaviour.
	ignoreUntaggedFields bool
}

// Set sets the snakecharmer options
func (sch *SnakeCharmer) Set(opts ...CharmingOption) error {
	for _, opt := range opts {
		if err := opt(sch); err != nil {
			return err
		}
	}
	return nil
}

// ResultStruct returns the pointer to the struct that contains the decoded values.
func (sch *SnakeCharmer) ResultStruct() interface{} { return sch.resultStruct }

// FieldTagName returns the tag name that snakecharmer reads for field names.
func (sch *SnakeCharmer) FieldTagName() string { return sch.fieldTagName }

// EnvTagName returns the tag name that snakecharmer reads for ENV var names.
func (sch *SnakeCharmer) EnvTagName() string { return sch.envTagName }

// FlagHelpTagName returns the tag name that snakecharmer reads for flag usage help.
func (sch *SnakeCharmer) FlagHelpTagName() string { return sch.flagHelpTagName }

// ConfigFileType returns the type that will be passed to viper.SetConfigType().
func (sch *SnakeCharmer) ConfigFileType() string { return sch.configFileType }

// ConfigFilePath returns the config file path that will be passed to
// viper.AddConfigPath() if path is a directory,
// viper.SetConfigFile() if path is a file.
func (sch *SnakeCharmer) ConfigFilePath() string { return sch.configFilePath }

// ConfigFileBaseName returns the base name of the config file (without extension)
// that will be passed to viper.SetConfigName().
func (sch *SnakeCharmer) ConfigFileBaseName() string { return sch.configFileBaseName }

// DecoderConfigOptions returns the slice of viper.DecoderConfigOption
// that will be passed to viper.Unmarshal()
func (sch *SnakeCharmer) DecoderConfigOptions() []viper.DecoderConfigOption {
	return sch.decoderConfigOptions
}

// IgnoreUntaggedFields returns the SnakeCharmer.ignoreUntaggedFields value
// that will be passed as viper.DecoderConfigOption
func (sch *SnakeCharmer) IgnoreUntaggedFields() bool { return sch.ignoreUntaggedFields }

// AddFlags creates flags from tags of a given Result Struct.
// Adds flags to cobra PersistentFlags flagset,
// creates viper's config param and sets default value (viper.SetDefault()),
// binds viper's config param with a corresponding flag from the cobra flagset,
// binds viper's config param with a corresponding ENV var
func (sch *SnakeCharmer) AddFlags() { sch.addFlags(sch.resultStruct, "") }

func (sch *SnakeCharmer) addFlags(input interface{}, prefix string) {
	var key, env, help string
	var err error

	v := reflect.ValueOf(input)
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			panic("BUG: got nil input")
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		panic(fmt.Sprintf("BUG: invalid input type: %q", v.Kind().String()))
	}
	for i := 0; i < v.NumField(); i++ {
		structField := v.Type().Field(i)
		fieldValue := v.Field(i)
		if fieldValue.Kind() == reflect.Ptr || fieldValue.Kind() == reflect.Interface {
			if fieldValue.IsNil() {
				panic(fmt.Sprintf("BUG: got nil for field: %s", structField.Name))
			}
			fieldValue = fieldValue.Elem()
		}

		fieldTag := structField.Tag.Get(sch.fieldTagName)
		// TODO: handle if field doesn't have fieldTag tag
		if len(fieldTag) == 0 {
			if sch.ignoreUntaggedFields {
				continue
			} else {
				panic(fmt.Sprintf("BUG: got untagged field: %s", structField.Name))
			}
		}
		key = strings.Split(fieldTag, ",")[0]

		if len(prefix) > 0 {
			key = prefix + "." + key
		}

		if fieldValue.Kind() == reflect.Struct {
			// Run addFlags recursively with prefix
			sch.addFlags(fieldValue.Interface(), key)
			continue
		}

		help = structField.Tag.Get(sch.flagHelpTagName)
		if len(help) == 0 {
			panic(fmt.Sprintf("BUG: %s tag is not specified for field: %q", sch.flagHelpTagName, structField.Name))
		}

		// Add Flag to cobra flagset and Set default viper config param
		if err = sch.applySetting(fieldValue, key, help); err != nil {
			panic(err.Error())
		}

		// Bind flag to viper.
		// This overrides viper default setting
		// with values from cobra flags.
		err = sch.viper.BindPFlag(key, sch.cmd.PersistentFlags().Lookup(key))
		if err != nil {
			panic(err.Error())
		}
		env = structField.Tag.Get(sch.envTagName)
		if len(env) > 0 {
			// Bind env var to viper.
			// This overrides viper default setting
			// with values from ENV vars.
			// Note: viper treats ENV variables as case sensitive.
			err = sch.viper.BindEnv(key, env)
			if err != nil {
				panic(err.Error())
			}
		}
	}
}

// UnmarshalExact unmarshals the config into a Struct,
// erroring if a field is nonexistent in the destination struct.
func (sch *SnakeCharmer) UnmarshalExact() (err error) {
	if len(sch.configFilePath) > 0 {
		if err = sch.mergeInConfigFile(); err != nil {
			return err
		}
	}
	if sch.fieldTagName != "mapstructure" {
		sch.decoderConfigOptions = append(sch.decoderConfigOptions,
			func(dc *mapstructure.DecoderConfig) { dc.TagName = sch.fieldTagName },
		)
	}
	err = sch.viper.UnmarshalExact(sch.resultStruct, sch.decoderConfigOptions...)
	if err != nil {
		return fmt.Errorf("while unmarshalling config, flags, and env vars: %s", err.Error())
	}
	return nil
}

func (sch *SnakeCharmer) mergeInConfigFile() (err error) {
	if len(sch.configFilePath) == 0 {
		return fmt.Errorf("config file path is an empty string")
	}

	found, err := sch.findConfigFile()
	if err != nil {
		return fmt.Errorf("while finding config %q: %s", sch.configFilePath, err.Error())
	}

	if !found {
		return nil
	}

	if err = sch.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("while reading config %q: %s", sch.configFilePath, err.Error())
	}
	return nil
}

func (sch *SnakeCharmer) findConfigFile() (bool, error) {
	if len(sch.configFilePath) == 0 {
		return false, fmt.Errorf("config file path is an empty string")
	}
	fileInfo, err := os.Stat(sch.configFilePath)
	if err == nil {
		// path exists
		if fileInfo.IsDir() {
			// path is a directory
			sch.viper.AddConfigPath(sch.configFilePath)     // path to look for the config file in
			sch.viper.SetConfigName(sch.configFileBaseName) // name of config file (without extension)
		} else {
			// path is a file
			fext := strings.TrimPrefix(filepath.Ext(sch.configFilePath), ".")
			if len(fext) == 0 {
				// REQUIRED since the config file does not have the extension in the name
				sch.viper.SetConfigType(sch.configFileType)
				sch.viper.SetConfigFile(sch.configFilePath)
			} else if fileExtSupported(fext) {
				// See viper.SupportedExts for full list of supported extensions
				sch.viper.SetConfigFile(sch.configFilePath)
			} else {
				// REQUIRED since the config file extension is not in the list of supported extensions
				sch.viper.SetConfigType(sch.configFileType)
				sch.viper.SetConfigFile(sch.configFilePath)
			}
		}
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		// path does *not* exist
		return false, fmt.Errorf("no such file or directory: %q", sch.configFilePath)
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		return false, fmt.Errorf("schrodinger: %q may or may not exist: %s", sch.configFilePath, err.Error())
	}
}

// This adds Flag to cobra flagset and sets default viper config param
func (sch *SnakeCharmer) applySetting(rv reflect.Value, name, help string) error {
	switch rv.Kind() {
	case reflect.Bool:
		value := rv.Bool()
		sch.cmd.PersistentFlags().Bool(name, value, help)
		sch.viper.SetDefault(name, value)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value := rv.Uint()
		sch.cmd.PersistentFlags().Uint64(name, value, help)
		sch.viper.SetDefault(name, value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value := rv.Int()
		sch.cmd.PersistentFlags().Int64(name, value, help)
		sch.viper.SetDefault(name, value)

	case reflect.Float32, reflect.Float64:
		value := rv.Float()
		sch.cmd.PersistentFlags().Float64(name, value, help)
		sch.viper.SetDefault(name, value)

	case reflect.String:
		value := rv.String()
		sch.cmd.PersistentFlags().String(name, value, help)
		sch.viper.SetDefault(name, value)

	case reflect.Slice:
		intf := rv.Interface()
		value, ok := intf.([]string)
		if !ok {
			return fmt.Errorf("BUG: invalid type: %T for flag %q", intf, name)
		}
		if len(value) == 0 {
			return fmt.Errorf("BUG: value of flag %q (%T) is nil or empty", name, intf)
		}
		sch.cmd.PersistentFlags().StringSlice(name, value, help)
		sch.viper.SetDefault(name, value)

	case reflect.Map:
		intf := rv.Interface()
		value, ok := intf.(map[string]string)
		if !ok {
			return fmt.Errorf("BUG: invalid type: %T for flag %q", intf, name)
		}
		if value == nil {
			return fmt.Errorf("BUG: value of flag %q (%T) is nil or empty", name, intf)
		}
		sch.cmd.PersistentFlags().StringToString(name, value, help)
		sch.viper.SetDefault(name, value)

	default:
		return fmt.Errorf("BUG: unsupported type: %q", rv.Kind().String())
	}
	return nil
}
