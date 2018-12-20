/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"fmt"

	"github.com/spf13/viper"
)

// LoadConfigWithEnvSubstitutionsAndContentReplacements reads the given configuration file, make substitutions for any embedded
// environment variables and loads the expanded content into the given viper instance.  It also performs any (optional) additional replacements
// as specified by contentReplacer
//
func LoadConfigWithEnvSubstitutionsAndContentReplacements(v *viper.Viper, confType string, data []byte, contentReplacer *strings.Replacer) error {

	cfgStr := string(data)
	if contentReplacer != nil {
		cfgStr = contentReplacer.Replace(cfgStr)
	}

	envExpandedCfg := os.ExpandEnv(cfgStr)

	v.SetConfigType(confType)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	err := v.ReadConfig(bytes.NewReader([]byte(envExpandedCfg)))
	if err != nil {
		return fmt.Errorf("failed to load configuration data. %v", err)
	}

	// By doing unmarshal and marshal, all env variables will be properly expanded.
	// viper.GetStringMap was not working properly with env variables. This will fix it.

	yamlMap := make(map[string]interface{})
	err = v.Unmarshal(&yamlMap)
	if err != nil {
		return fmt.Errorf("viper.Unmarshal failed %v", err)
	}

	rawCfg, err := yaml.Marshal(&yamlMap)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration content into YAML %v", err)
	}

	err = v.ReadConfig(bytes.NewReader(rawCfg))
	if err != nil {
		return fmt.Errorf("readconfig failed: %v", err)
	}

	return nil

}

// SetupViper reads the given configuration file, makes substitutions for any embedded
// environment variables and loads the expanded content into the default viper instance
func SetupViper(cfgFile string) error {
	var err error
	var data []byte
	confType := "yaml"
	if cfgFile != "" {
		data, err = ioutil.ReadFile(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to read configuration file %s. %v", cfgFile, err)
		}
		ext := filepath.Ext(cfgFile)
		if len(ext) > 1 {
			confType = ext[1:len(ext)]
		}
	}

	v := viper.GetViper()
	return LoadConfigWithEnvSubstitutionsAndContentReplacements(v, confType, data, nil)
}

// GetViperFromCfgFile reads the given configuration file and returns a viper instance
func GetViperFromCfgFile(cfgFile string) (*viper.Viper, error) {
	if err := SetupViper(cfgFile); err != nil {
		return nil, fmt.Errorf("Failed to set up viper using config file and environmental variables, %v", err)
	}
	return viper.GetViper(), nil
}
