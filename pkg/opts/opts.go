package opts

import (
	"fmt"
	"io/ioutil"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

const (
	// DifferAnnotation represents the required kubrentes manifest annotation to get scraped from differ
	DifferAnnotation = "differ/active"
)

// GitRemote struct descripes remote config
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
func Init(configPath, logLevel string) (*ControllerConfig, error) {

	if err := setLoglevel(logLevel); err != nil {
		return &ControllerConfig{}, err
	}
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
		return fmt.Errorf("Configuration file \"%s\" is invalid. Please resolve errors", c.configPath)
	}
	return nil
}

func setLoglevel(level string) error {
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

// ControllerSleep sleep for configured duration
func (c *ControllerConfig) ControllerSleep() {
	time.Sleep(c.ParsedSleep)
}
