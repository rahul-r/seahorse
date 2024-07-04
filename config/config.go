package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port            int    `yaml:"port"`
	TemplatesDir    string `yaml:"template_dir"`
	OutputDir       string `yaml:"output_dir"`
	EnvironmentFile string `yaml:"env_file"`
	DockerHost      string `yaml:"docker_host"`
	UseRemoteDocker bool
}

func LoadConfig(filename string) (*Config, error) {
	defaultConfig := &Config{
		Port:            9843,
		TemplatesDir:    "/compose-templates",
		OutputDir:       "/tmp/compose-output",
		EnvironmentFile: "/environment",
		DockerHost:      "",
		UseRemoteDocker: false,
	}

	// Check if the file exists
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		// File doesn't exist, return default config
		return defaultConfig, nil
	} else if err != nil {
		// Other error occurred
		return defaultConfig, err
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return defaultConfig, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return defaultConfig, err
	}

	// If any configuration value is missing, use the default value
	if config.Port == 0 {
		log.Printf("port config not supplied,, using default value `%d`", defaultConfig.Port)
		config.Port = defaultConfig.Port
	}
	if config.TemplatesDir == "" {
		log.Printf("template_dir config not supplied,, using default value `%s`", defaultConfig.TemplatesDir)
		config.TemplatesDir = defaultConfig.TemplatesDir
	}
	if config.OutputDir == "" {
		log.Printf("output_dir config not supplied, using default value `%s`", defaultConfig.OutputDir)
		config.OutputDir = defaultConfig.OutputDir
	}
	if config.EnvironmentFile == "" {
		log.Printf("env_file config not supplied, using default value `%s`", defaultConfig.EnvironmentFile)
		config.EnvironmentFile = defaultConfig.EnvironmentFile
	}
	if config.DockerHost == "" {
		log.Println("docker_host config not supplied,, using local docker")
		config.DockerHost = defaultConfig.DockerHost
		config.UseRemoteDocker = false
	} else {
		config.UseRemoteDocker = true
	}

	return &config, nil
}
