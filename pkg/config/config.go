/*
 * MIT License
 *
 * Copyright (c) 2019 Felix Wiedmann
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package config

import (
	"io/ioutil"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type MetricsEndpoint struct {
	Port int    `yaml:"port"`
	Path string `yaml:"path"`
}

// GitRemote struct describes remote config
type GitRemote struct {
	Provider       string `yaml:"provider" validate:"required"`
	Repositoryname string `yaml:"reponame" validate:"required"`
	Username       string `yaml:"username" validate:"required"`
	CustomURL      string `yaml:"customURL,omitempty"`
}

// ControllerConfig holds required controller configuration
type ControllerConfig struct {
	Namespace                            string          `yaml:"namespace"`
	UnparsedRegistryRequestSleepDuration string          `yaml:"registryRequestSleepDuration,omitempty"`
	GitRemotes                           []GitRemote     `yaml:"remotes,omitempty" validate:"dive,required"`
	Metrics                              MetricsEndpoint `yaml:"metrics"  validate:"required,dive,required"`
	LogLevel                             string          `yaml:"loglevel,omitempty"`
	ParsedRegistryRequestSleepDuration   time.Duration   `yaml:"-"`
	configPath                           string          `yaml:"-"`
	Version                              string          `yaml:"-"`
}

type Config struct {
	config     *ControllerConfig
	configLock sync.RWMutex
	configFile string
}

func NewConfig(configPath, differVersion string) (*Config, error) {
	config, err := initConfig(configPath)
	if err != nil {
		return nil, err
	}
	config.Version = differVersion
	return &Config{
		config:     config,
		configLock: sync.RWMutex{},
		configFile: configPath,
	}, nil
}

func (o *Config) GetConfig() *ControllerConfig {
	o.configLock.Lock()
	defer o.configLock.Unlock()

	return o.config
}

func (o *Config) ReloadConfig() error {
	o.configLock.Lock()
	defer o.configLock.Unlock()
	log.Infof("Reloading config")
	newConf, err := initConfig(o.configFile)
	if err != nil {
		return err
	}
	o.config = newConf

	return nil
}

// Init initialize controller configuration
func initConfig(configPath string) (*ControllerConfig, error) {

	config := &ControllerConfig{configPath: configPath,
		Metrics: MetricsEndpoint{
			Port: 9100,
			Path: "/metrics",
		}}

	configFile, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(configFile, config); err != nil {
		return nil, err
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, err
	}

	if config.UnparsedRegistryRequestSleepDuration == "" {
		config.UnparsedRegistryRequestSleepDuration = "5s"
	}

	dur, err := time.ParseDuration(config.UnparsedRegistryRequestSleepDuration)
	if err != nil {
		return nil, err
	}
	config.ParsedRegistryRequestSleepDuration = dur

	if err = setLoglevel(config.LogLevel); err != nil {
		return nil, err
	}
	log.Debugf("Parsed config: %+v", config)
	return config, nil

}

func setLoglevel(level string) error {
	if level == "" {
		level = "info"
	}
	parsedLevel, err := log.ParseLevel(level)
	if err != nil {
		return err
	}

	log.SetLevel(parsedLevel)
	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetFormatter(&nested.Formatter{
		HideKeys: true,
	})
	return nil
}
