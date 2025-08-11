package main

import (
	"flag"
	"net/http"
	"os"
	"strings"
)

type config struct {
	id             string
	httpPort       string
	raftPort       string
	gossipPort     string
	joinHost       string
	existingGossip string
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

	podName := os.Getenv("POD_NAME")
	joinHost := os.Getenv("JOIN_HOST")

	if podName == "" {
		podName = "node1"
	}
	if joinHost == "" {
		joinHost = "localhost"
	}

	flag.StringVar(&cfg.id, "node-id", podName, "Unique identifier for the node")
	flag.StringVar(&cfg.httpPort, "http-port", "2222", "Port for HTTP communication")
	flag.StringVar(&cfg.raftPort, "raft-port", "8222", "Port for Raft communication")
	flag.StringVar(&cfg.gossipPort, "gossip-port", "7469", "Port for Gossip communication")
	flag.StringVar(&cfg.joinHost, "join-host", joinHost, "Hostname to join cluster")
	flag.StringVar(&cfg.existingGossip, "existing-gossip", "", "Port for joining gossip cluster")
	flag.Parse()

	return cfg
}
