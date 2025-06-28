package main

import (
	"flag"
	"log"
	"net/http"
	"strings"
)

type config struct {
	id       string
	httpPort string
	raftPort string
}

type setPayload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type deletePayload struct {
	Key string `json:"key"`
}

func parsePath(r *http.Request) (cmd string, args []string) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}
func getConfig() config {
	cfg := config{}

	flag.StringVar(&cfg.id, "node-id", "node1", "Unique identifier for the node")
	flag.StringVar(&cfg.httpPort, "http-port", "2222", "Port for HTTP communication")
	flag.StringVar(&cfg.raftPort, "raft-port", "8222", "Port for Raft communication")

	flag.Parse()

	if cfg.id == "" {
		log.Fatal("Missing required parameter: --node-id")
	}

	if cfg.raftPort == "" {
		log.Fatal("Missing required parameter: --raft-port")
	}

	if cfg.httpPort == "" {
		log.Fatal("Missing required parameter: --http-port")
	}

	return cfg
}
