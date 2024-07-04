package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"seahorse/compose_template"
	"seahorse/config"
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

func composeInstallHandler(containerClient *containers.Containers, config *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := url.Parse(r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatal(err)
			return
		}

		params := u.Query()
		containerName := params.Get("container")

		err = compose_template.InstallCompose(containerName, containerClient, config)
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

func startServer(config *config.Config, containerClient *containers.Containers) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler(containerClient))
	mux.Handle("/public/", http.FileServer(http.FS(content)))

	mux.HandleFunc("/start", containerStartHandler(containerClient))
	mux.HandleFunc("/stop", containerStopHandler(containerClient))
	mux.HandleFunc("/restart", containerRestartHandler(containerClient))
	mux.HandleFunc("/install", composeInstallHandler(containerClient, config))
	mux.HandleFunc("/update", composeInstallHandler(containerClient, config)) // Run same handler for install and update

	fmt.Printf("Running HTTP server on port %d\n", config.Port)
	http.ListenAndServe(":"+strconv.Itoa(config.Port), mux)
}
