//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
//
package utils

import (
	"fmt"
	"os"
	"strings"

	logging "github.com/op/go-logging"
	"github.com/spf13/viper"
)

const (
	configLoggingFormat = "logging.format"
	configLoggingLevel  = "logging.level"
	defaultLogFormat    = "%{time:2006-01-02T15:04:05.999Z-05:00} %{shortfunc} ▶ %{level:.4s} %{id:03x} %{message}"
	defaultLogLevel     = "info"
)

// InitLogger sets the logging format and level.
func InitLogger() error {
	v := viper.GetViper()
	logPattern := v.GetString(configLoggingFormat)
	logLevel := v.GetString(configLoggingLevel)
	if len(logPattern) > 0 && len(logLevel) > 0 {
		return initLogging(logPattern, logLevel)
	} else if len(logPattern) > 0 {
		return initLogging(logPattern, defaultLogLevel)
	} else if len(logLevel) > 0 {
		return initLogging(defaultLogFormat, logLevel)
	}
	return initLogging(defaultLogFormat, defaultLogLevel)
}

func initLogging(pattern string, level string) error {
	format := logging.MustStringFormatter(pattern)
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	formatter := logging.NewBackendFormatter(backend, format)
	backendLeveled := logging.AddModuleLevel(formatter)
	var logLevel logging.Level
	switch strings.ToLower(level) {
	case "critical":
		logLevel = logging.CRITICAL
	case "error":
		logLevel = logging.ERROR
	case "warning":
		logLevel = logging.WARNING
	case "notice":
		logLevel = logging.NOTICE
	case "info":
		logLevel = logging.INFO
	case "debug":
		logLevel = logging.DEBUG
	default:
		return fmt.Errorf("unknown log level: %s, available log levels are critical, error, warning, notice, info, and debug", level)
	}
	backendLeveled.SetLevel(logLevel, "")
	logging.SetBackend(backendLeveled)
	return nil
}
