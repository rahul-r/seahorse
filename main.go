package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"seahorse/compose_template"
	"seahorse/config"
	"seahorse/containers"
)

func parseCommandLineArgs() (string, string, string) {
	dirFlag := flag.String("dir", "", "Directory containing template files")
	fileFlag := flag.String("file", "", "Single template file to process")
	installFlag := flag.String("install", "", "Install the container")
	helpFlag := flag.Bool("help", false, "Print help message")
	flag.StringVar(dirFlag, "d", "", "Directory containing template files (shorthand)")
	flag.StringVar(fileFlag, "f", "", "Single template file to process (shorthand)")
	flag.BoolVar(helpFlag, "h", false, "Print help message (shorthand)")
	flag.Parse()

	if *helpFlag {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return "", "", ""
	}

	return *dirFlag, *fileFlag, *installFlag
}

func main() {
	// Enable line numbers in log messages
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config, err := config.LoadConfig("config.yml")
	if err != nil {
		log.Println(err)
		log.Println("Using default config values")
	}

	dirPath, filePath, installContainer := parseCommandLineArgs()

	if installContainer != "" {
		fmt.Printf("Installing container %s\n", installContainer)

		var containerClient containers.Containers
		if config.UseRemoteDocker {
			containerClient = containers.NewRemoteClient(config.DockerHost)
		} else {
			containerClient = containers.NewLocalClient()
		}

		composeFiles, err := compose_template.ScanDir(config.TemplatesDir)
		if err != nil {
			log.Fatal(err)
			return
		}

		err = containerClient.CreateContainerMap(*composeFiles)
		if err != nil {
			log.Fatal(err)
			return
		}

		err = compose_template.InstallCompose(installContainer, &containerClient, config)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	if len(os.Args) == 1 {
		composeFiles, err := compose_template.ScanDir(config.TemplatesDir)
		if err != nil {
			log.Fatal(err)
			return
		}

		var containerClient containers.Containers
		if config.UseRemoteDocker {
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
