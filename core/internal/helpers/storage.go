/* Copyright 2017 LinkedIn Corp. Licensed under the Apache License, Version
 * 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 */

package helpers

import (
	"fmt"
	"os"
	"strings"
	"time"
	"strconv"

	"github.com/linkedin/Burrow/core/protocol"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func ConfigureLogger() (*zap.Logger, *zap.AtomicLevel) {
	var level zap.AtomicLevel
	var syncOutput zapcore.WriteSyncer

	// Set config defaults for logging
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.maxsize", 100)
	viper.SetDefault("logging.maxbackups", 10)
	viper.SetDefault("logging.maxage", 30)

	// Create an AtomicLevel that we can use elsewhere to dynamically change the logging level
	logLevel := viper.GetString("logging.level")
	switch strings.ToLower(logLevel) {
	case "", "info":
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "debug":
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "warn":
		level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "panic":
		level = zap.NewAtomicLevelAt(zap.PanicLevel)
	case "fatal":
		level = zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		fmt.Printf("Invalid log level supplied. Defaulting to info: %s", logLevel)
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// If a filename has been set, set up a rotating logger. Otherwise, use Stdout
	logFilename := viper.GetString("logging.filename")
	if logFilename != "" {
		syncOutput = zapcore.AddSync(&lumberjack.Logger{
			Filename:   logFilename,
			MaxSize:    viper.GetInt("logging.maxsize"),
			MaxBackups: viper.GetInt("logging.maxbackups"),
			MaxAge:     viper.GetInt("logging.maxage"),
			LocalTime:  viper.GetBool("logging.use-localtime"),
			Compress:   viper.GetBool("logging.use-compression"),
		})
	} else {
		syncOutput = zapcore.Lock(os.Stdout)
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		syncOutput,
		level,
	)
	logger := zap.New(core)
	zap.ReplaceGlobals(logger)
	return logger, &level
}

// TimeoutSendStorageRequest is a helper func for sending a protocol.StorageRequest to a channel with a timeout,
// specified in seconds. If the request is sent, return true. Otherwise, if the timeout is hit, return false.
func TimeoutSendStorageRequest(storageChannel chan *protocol.StorageRequest, request *protocol.StorageRequest, maxTime int) bool {
	logger, _ := ConfigureLogger()
	logger.Info("In TimeoutSendStorageRequest(...)")
	timeout := time.After(time.Duration(maxTime) * time.Second)
	logger.Info("Duration = " + strconv.Itoa(maxTime))
	logger.Info("After the timeout in TimeoutSendStorageRequest(...)")
	select {
	case storageChannel <- request:
		return true
	case <-timeout:
		return false
	}
}
