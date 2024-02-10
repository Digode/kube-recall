package config

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
)

type (
	DataDog struct {
		ApiKey string `yaml:"apiKey"`
		AppKey string `yaml:"appKey"`
		Times  struct {
			Begin int `yaml:"begin"`
			End   int `yaml:"end"`
		} `yaml:"times"`
		Tags struct {
			Cluster    string   `yaml:"cluster"`
			Namespace  string   `yaml:"namespace"`
			Deployment string   `yaml:"deployment"`
			Replicas   string   `yaml:"replicas"`
			Cpu        Resource `yaml:"cpu"`
			Memory     Resource `yaml:"memory"`
		} `yaml:"tags"`
	}
	Kubernetes struct {
		Targets []Target `yaml:"targets"`
	}
	Resource struct {
		Usage   string `yaml:"usage"`
		Request string `yaml:"request"`
		Limit   string `yaml:"limit"`
	}
	Scale struct {
		Request float64 `yaml:"request"`
		Limit   float64 `yaml:"limit"`
	}
	Config struct {
		DataDog    DataDog           `yaml:"datadog"`
		Kubernetes Kubernetes        `yaml:"kubernetes"`
		Filters    map[string]string `yaml:"filters"`
		Scales     struct {
			Cpu    Scale `yaml:"cpu"`
			Memory Scale `yaml:"memory"`
		} `yaml:"scales"`
	}
	Target struct {
		id         string
		From       string `yaml:"from"`
		To         string `yaml:"to"`
		ConfigPath string `yaml:"configPath"`
	}
)

var config = Config{}

func Get() Config {
	if !cmp.Equal(config, Config{}) {
		return config
	}
	err := config.file()
	if err != nil {
		panic(err)
	}
	config.env()
	return config
}

func (t *Target) ID() string {
	if t.id == "" {
		t.id = uuid.NewString()
	}
	return t.id
}
