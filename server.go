package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"

	"embed"
	"seahorse/compose_template"
	"seahorse/containers"
)

//go:embed public/*
var content embed.FS

//go:embed templates/*
var templateFiles embed.FS

var indexTemplate = template.Must(template.ParseFS(templateFiles, "templates/index.html"))

func indexHandler(containerClient *containers.Containers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		buf := &bytes.Buffer{}

		err := indexTemplate.Execute(buf, containerClient.GetContainerMap())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatal(err)
			return
		}

		buf.WriteTo(w)
	}
}

func containerStartHandler(client *containers.Containers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := url.Parse(r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatal(err)
			return
		}

		params := u.Query()
		containerName := params.Get("container")

		client.Start(containerName)

		if entry, ok := (*client.GetContainerMap())[containerName]; ok {
			entry.State = client.GetContainerStatus(containerName)
			(*client.GetContainerMap())[containerName] = entry
		}

		buf := &bytes.Buffer{}
		err = indexTemplate.Execute(buf, client.GetContainerMap())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatal(err)
			return
		}

		buf.WriteTo(w)
	}
}

func containerStopHandler(client *containers.Containers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := url.Parse(r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatal(err)
			return
		}

		params := u.Query()
		containerName := params.Get("container")

		client.Stop(containerName)

		if entry, ok := (*client.GetContainerMap())[containerName]; ok {
			entry.State = client.GetContainerStatus(containerName)
			(*client.GetContainerMap())[containerName] = entry
		}

		buf := &bytes.Buffer{}
		err = indexTemplate.Execute(buf, client.GetContainerMap())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatal(err)
			return
		}

		buf.WriteTo(w)
	}
}

func containerRestartHandler(client *containers.Containers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := url.Parse(r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatal(err)
			return
		}

		params := u.Query()
		containerName := params.Get("container")

		client.Stop(containerName)
		client.Start(containerName)

		if entry, ok := (*client.GetContainerMap())[containerName]; ok {
			entry.State = client.GetContainerStatus(containerName)
			(*client.GetContainerMap())[containerName] = entry
		}

		buf := &bytes.Buffer{}
		err = indexTemplate.Execute(buf, client.GetContainerMap())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatal(err)
			return
		}

		buf.WriteTo(w)
	}
}

func composeInstallHandler(containerClient *containers.Containers, environmentFile string, dockerHost string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := url.Parse(r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatal(err)
			return
		}

		params := u.Query()
		containerName := params.Get("container")

		tmpDir, err := os.MkdirTemp("", "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatal(err)
			return
		}
		defer os.RemoveAll(tmpDir)

		tmpDir = path.Join(tmpDir, containerName)
		err = os.Mkdir(tmpDir, os.ModePerm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatal(err)
			return
		}
		defer os.RemoveAll(tmpDir)

		err = compose_template.ProcessDir((*containerClient.GetContainerMap())[containerName].TemplateDir, tmpDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatal(fmt.Sprintf("Cannot process template for %s: %s", containerName, err))
			return
		}

		cmd := exec.Command("bash", "-c", `
			set -e
			docker compose --env-file $ENV_FILE pull
			docker compose --env-file $ENV_FILE up -d --remove-orphans
		`)
		cmd.Dir = tmpDir
		cmd.Env = append(os.Environ(), fmt.Sprintf("DOCKER_HOST=%s", dockerHost), fmt.Sprintf("ENV_FILE=%s", environmentFile))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatal(err)
			return
		}

		buf := &bytes.Buffer{}
		err = indexTemplate.Execute(buf, containerClient.GetContainerMap())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatal(err)
			return
		}

		buf.WriteTo(w)
	}
}

func startServer(config *Config, containerClient *containers.Containers) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler(containerClient))
	mux.Handle("/public/", http.FileServer(http.FS(content)))

	mux.HandleFunc("/start", containerStartHandler(containerClient))
	mux.HandleFunc("/stop", containerStopHandler(containerClient))
	mux.HandleFunc("/restart", containerRestartHandler(containerClient))
	mux.HandleFunc("/install", composeInstallHandler(containerClient, config.EnvironmentFile, config.DockerHost))
	mux.HandleFunc("/update", composeInstallHandler(containerClient, config.EnvironmentFile, config.DockerHost)) // Run same handler for install and update

	fmt.Printf("Running HTTP server on port %d\n", config.Port)
	http.ListenAndServe(":"+strconv.Itoa(config.Port), mux)
}
