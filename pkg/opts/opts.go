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

package opts

import (
	"fmt"
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

const (
	// DifferAnnotation represents the required kubernetes manifest annotation to get scraped from differ
	DifferAnnotation = "differ/active"
)

// GitRemote struct describes remote config
type GitRemote struct {
	Provider       string `yaml:"provider"`
	Repositoryname string `yaml:"reponame"`
	Username       string `yaml:"username"`
	CustomURL      string `yaml:"customURL,omitempty"`
}

// ControllerConfig holds required controller configuration
type ControllerConfig struct {
	Namespace   string        `yaml:"namespace"`
	GitRemotes  []GitRemote   `yaml:"remotes"`
	Sleep       string        `yaml:"controllerSleep"`
	configPath  string        `ymaml:"-"`
	ParsedSleep time.Duration `ymaml:"-"`
}

// Init initialize controller configuration
func Init(configPath string) (*ControllerConfig, error) {

	config := &ControllerConfig{configPath: configPath}

	configFile, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(configFile, config); err != nil {
		return nil, err
	}
	log.Debugf("Parsed config: %+v", config)

	if err = validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil

}

func validateConfig(c *ControllerConfig) error {
	isValid := true
	duration, err := time.ParseDuration(c.Sleep)
	if err != nil {
		log.Error(err)
		isValid = false
	}
	c.ParsedSleep = duration

	if !isValid {
		return fmt.Errorf("configuration file \"%s\" is invalid. Please resolve errors", c.configPath)
	}
	return nil
}

// ControllerSleep sleep for configured duration
func (c *ControllerConfig) ControllerSleep() {
	log.Infof("Done, start sleeping for %s", c.Sleep)
	time.Sleep(c.ParsedSleep)
}
