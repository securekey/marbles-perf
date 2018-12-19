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
package fabricclient

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/core/logging/api"
	"github.com/op/go-logging"
)

// sdkLoggerProvider a logger provider that implements api.LoggerProvider interface in fabric-sdk
type sdkLoggerProvider struct{ log *logging.Logger }

// sdkLogger is a logger that implements api.Logger interface in fabric-sdk
type sdkLogger struct{ logging.Logger }

// GetLogger is an implementation of api.LoggerProvider GetLogger
func (p *sdkLoggerProvider) GetLogger(module string) api.Logger {
	return &sdkLogger{*p.log}
}

// Fatalln is an implementation of api.Logger Fataln
func (l *sdkLogger) Fatalln(v ...interface{}) {
	v = append(v, "\n")
	l.Fatal(v)
}

// Panicln is an implementation of api.Logger Panicln
func (l *sdkLogger) Panicln(v ...interface{}) {
	v = append(v, "\n")
	l.Panic(v)
}

// Print is an implementation of api.Logger Print
func (l *sdkLogger) Print(v ...interface{}) {
	l.Info(v)
}

// Println is an implementation of api.Logger Println
func (l *sdkLogger) Println(v ...interface{}) {
	v = append(v, "\n")
	l.Print(v)
}

// Printf is an implementation of api.Logger Printf
func (l *sdkLogger) Printf(format string, v ...interface{}) {
	l.Infof(format, v)
}

// Debugln is an implementation of api.Logger Debugln
func (l *sdkLogger) Debugln(v ...interface{}) {
	v = append(v, "\n")
	l.Debug(v)
}

// Infoln is an implementation of api.Logger Infoln
func (l *sdkLogger) Infoln(v ...interface{}) {
	v = append(v, "\n")
	l.Info(v)
}

// Warn is an implementation of api.Logger Warn
func (l *sdkLogger) Warn(v ...interface{}) {
	l.Warning(v)
}

// Warnln is an implementation of api.Logger Warnln
func (l *sdkLogger) Warnln(v ...interface{}) {
	v = append(v, "\n")
	l.Warn(v)
}

// Warnf is an implementation of api.Logger Warnf
func (l *sdkLogger) Warnf(format string, v ...interface{}) {
	l.Warningf(format, v)
}

// Errorln is an implementation of api.Logger Errorln
func (l *sdkLogger) Errorln(v ...interface{}) {
	v = append(v, "\n")
	l.Error(v)
}
