package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"seahorse/compose_template"
	"seahorse/containers"

	"gopkg.in/yaml.v3"
)

func parseCommandLineArgs() (string, string) {
	dirFlag := flag.String("dir", "", "Directory containing template files")
	fileFlag := flag.String("file", "", "Single template file to process")
	helpFlag := flag.Bool("help", false, "Print help message")
	flag.StringVar(dirFlag, "d", "", "Directory containing template files (shorthand)")
	flag.StringVar(fileFlag, "f", "", "Single template file to process (shorthand)")
	flag.BoolVar(helpFlag, "h", false, "Print help message (shorthand)")
	flag.Parse()

	if *helpFlag {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return "", ""
	}

	return *dirFlag, *fileFlag
}

type Config struct {
	Port            int    `yaml:"port"`
	TemplatesDir    string `yaml:"template_dir"`
	OutputDir       string `yaml:"output_dir"`
	EnvironmentFile string `yaml:"env_file"`
	DockerHost      string `yaml:"docker_host"`
	useRemoteDocker bool
}

func LoadConfig(filename string) (*Config, error) {
	defaultConfig := &Config{
		Port:            9843,
		TemplatesDir:    "/compose-templates",
		OutputDir:       "/tmp/compose-output",
		EnvironmentFile: "/environment",
		DockerHost:      "",
		useRemoteDocker: false,
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
		config.useRemoteDocker = false
	} else {
		config.useRemoteDocker = true
	}

	return &config, nil
}

func main() {
	// Enable line numbers in log messages
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config, err := LoadConfig("config.yml")
	if err != nil {
		log.Println(err)
		log.Println("Using default config values")
	}

	dirPath, filePath := parseCommandLineArgs()
	if len(os.Args) == 1 {
		composeFiles, err := compose_template.ScanDir(config.TemplatesDir)
		if err != nil {
			log.Fatal(err)
			return
		}

		var containerClient containers.Containers
		if config.useRemoteDocker {
			containerClient = containers.NewRemoteClient(config.DockerHost)
		} else {
			containerClient = containers.NewLocalClient()
		}

		err = containerClient.CreateContainerMap(*composeFiles)
		if err != nil {
			log.Fatal(err)
			return
		}

		startServer(config, &containerClient)
	} else if dirPath != "" {
		log.Printf("Processing directory `%s`\n", dirPath)
		err := compose_template.ProcessDir(dirPath, config.OutputDir)
		if err != nil {
			log.Fatal(err)
			return
		}
	} else if filePath != "" {
		log.Printf("Processing file `%s`\n", filePath)
		outputFilePath := filepath.Join(config.OutputDir, filepath.Base(filePath))
		err := compose_template.ProcessFile(filePath, outputFilePath)
		if err != nil {
			log.Fatal(err)
			return
		}
		fmt.Printf("Output written to %s\n", outputFilePath)
	} else {
		log.Fatalf("Unknown commandline arguments: %v\n", os.Args)
		return
	}
}
