package opts

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const (
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
	Namespace  string      `yaml:"namespace"`
	GitRemotes []GitRemote `yaml:"remotes"`
}

func Init() (*ControllerConfig, error) {

	configFile, err := ioutil.ReadFile("config.yaml")

	if err != nil {
		return nil, err
	}
	config := &ControllerConfig{}

	if err = yaml.Unmarshal(configFile, config); err != nil {
		return nil, err
	}

	if err = validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil

}

func validateConfig(c *ControllerConfig) error {
	return nil
}
