package compose_template

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"gopkg.in/yaml.v3"
)

type ValuesMap map[string]interface{}

func ProcessFile(templateFilePath, outputFilePath string) error {
	templateFile, err := os.ReadFile(templateFilePath)
	if err != nil {
		return fmt.Errorf("error reading template file %s: %v", templateFilePath, err)
	}

	templatesDir := path.Dir(path.Dir(templateFilePath))

	// Remove ".Values" from the template
	templateStr := strings.ReplaceAll(string(templateFile), ".Values", "")

	valuesFileContent, err := os.ReadFile(path.Join(templatesDir, "values.yml"))
	if err != nil {
		return fmt.Errorf("error reading values file: %v", err)
	}

	var values ValuesMap
	err = yaml.Unmarshal(valuesFileContent, &values)
	if err != nil {
		return fmt.Errorf("error parsing values YAML: %v", err)
	}

	// Create a custom function map with the "include" function
	funcMap := sprig.TxtFuncMap()
	funcMap["include"] = func(name string, data interface{}) (string, error) {
		includeFile, err := os.ReadFile(path.Join(templatesDir, "includes.template"))
		if err != nil {
			return "", err
		}

		tmpl, err := template.New(name).Parse(string(includeFile))
		if err != nil {
			return "", err
		}

		var output bytes.Buffer
		err = tmpl.Execute(&output, data)
		if err != nil {
			return "", err
		}

		return output.String(), nil
	}

	tmpl, err := template.New("template").Funcs(funcMap).Parse(templateStr)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, values)
	if err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	err = os.MkdirAll(path.Dir(outputFilePath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("Error creating directory: %f", err)
	}

	err = os.WriteFile(outputFilePath, output.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("error writing output file: %v", err)
	}

	return nil
}

func ProcessDir(templatesDir, outputDir string) error {
	fileInfo, err := os.Stat(templatesDir)
	if err != nil {
		return err
	}
	if !fileInfo.IsDir() {
		panic(fmt.Sprintf("%s is not a directory", templatesDir))
	}

	return filepath.Walk(templatesDir, func(templatePath string, templateFileInfo os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}

		if templateFileInfo.IsDir() {
			return nil
		}

		templateFileRelativePath, err := filepath.Rel(templatesDir, templatePath)
		if err != nil {
			panic(err)
		}

		outputFilePath := filepath.Join(outputDir, templateFileRelativePath)

		if strings.HasSuffix(templateFileInfo.Name(), ".yml") && templateFileInfo.Name() != "values.yml" {
			err = ProcessFile(templatePath, outputFilePath)
			if err != nil {
				panic(err)
			}
		} else {
			// Copy non-YML files without applying the template
			input, err := os.ReadFile(templatePath)
			if err != nil {
				panic(err)
			}

			err = os.MkdirAll(filepath.Dir(outputFilePath), 0755)
			if err != nil {
				panic(err)
			}

			err = os.WriteFile(outputFilePath, input, 0644)
			if err != nil {
				panic(err)
			}
		}

		return nil
	})
}

type ContainerInfo struct {
	Name        string
	TemplateDir string
	ID          string `json:"Id"`
	Image       string
	ImageID     string
	Command     string
	Created     int64
	SizeRw      int64 `json:",omitempty"`
	SizeRootFs  int64 `json:",omitempty"`
	Labels      map[string]string
	State       string
	Status      string
	HostConfig  struct {
		NetworkMode string            `json:",omitempty"`
		Annotations map[string]string `json:",omitempty"`
	}
	// Names      []string
	// Ports      []Port
	// NetworkSettings *SummaryNetworkSettings
	// Mounts          []MountPoint
}

func ScanDir(templatesDir string) (*map[string]ContainerInfo, error) {
	containers := make(map[string]ContainerInfo, 0)

	filepath.Walk(templatesDir, func(templatePath string, templateFileInfo os.FileInfo, err error) error {
		if err != nil {
			log.Fatal(err)
			return err
		}

		if templateFileInfo.IsDir() {
			return nil
		}

		if path.Dir(templatePath) == templatesDir {
			return nil
		}

		if templateFileInfo.Name() == "docker-compose.yml" || templateFileInfo.Name() == "compose.yml" {
			name := path.Base(path.Dir(templatePath))
			absPath, err := filepath.Abs(path.Dir(templatePath))
			if path.Dir(templatePath) == templatesDir {
				log.Fatalf("Cannot create absolute path to %s: %s", templatePath, err)
				return nil
			}
			containers[name] = ContainerInfo{Name: name, TemplateDir: absPath}
		}

		return nil
	})

	return &containers, nil
}
