package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AlexG28/keyvalue/store"
	"github.com/hashicorp/raft"
)

type RaftNode interface {
	Apply([]byte, time.Duration) raft.ApplyFuture
	AddVoter(raft.ServerID, raft.ServerAddress, uint64, time.Duration) raft.IndexFuture
	State() raft.RaftState
}

type httpServer struct {
	r RaftNode
	s store.Store
}

func (hs httpServer) Set(w http.ResponseWriter, r *http.Request) {
	cmd, args := parsePath(r)
	if cmd != "Set" || len(args) < 2 {
		http.Error(w, "Invalid URL format. Expected Set/{key}/{value}", http.StatusBadRequest)
		return
	}
	key, value := args[0], args[1]
	if key == "" {
		http.Error(w, "Invalid URL format. Expected Set/{key}/{value}", http.StatusBadRequest)
		return
	}
	if value == "" {
		http.Error(w, "Invalid URL format. Expected Set/{key}/{value}", http.StatusBadRequest)
		return
	}

	payload := setPayload{Key: key, Value: value}
	data, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to marshal payload", http.StatusInternalServerError)
		return
	}

	future := hs.r.Apply(data, 3*time.Second)
	if err := future.Error(); err != nil {
		http.Error(w, fmt.Sprintf("Could not write key-value: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Set key '%s' to value '%s'\n", key, value)
}

func (hs httpServer) Get(w http.ResponseWriter, r *http.Request) {
	cmd, args := parsePath(r)
	if cmd != "Get" || len(args) < 1 {
		http.Error(w, "Invalid URL format. Expected Get/{key}", http.StatusBadRequest)
		return
	}
	key := args[0]
	val, err := hs.s.Get(key)
	if err != nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "%s", val)
}

func (hs httpServer) Join(w http.ResponseWriter, r *http.Request) {
	followerId := r.URL.Query().Get("followerId")
	followerAddr := r.URL.Query().Get("followerAddr")

	if followerId == "" || followerAddr == "" {
		http.Error(w, "Missing followerId or followerAddr", http.StatusBadRequest)
		return
	}

	if hs.r.State() != raft.Leader {
		http.Error(w, "Error not the leader", http.StatusBadRequest)
		return
	}

	err := hs.r.AddVoter(raft.ServerID(followerId), raft.ServerAddress(followerAddr), 0, 0).Error()

	if err != nil {
		log.Printf("Failed to add follower: %s", err)
		http.Error(w, "Failed to add follower", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully added follower %s at %s\n", followerId, followerAddr)
}

func (hs httpServer) Delete(w http.ResponseWriter, r *http.Request) {
	cmd, args := parsePath(r)
	if cmd != "Delete" || len(args) < 1 {
		http.Error(w, "Invalid URL format. Expected Delete/{key}", http.StatusBadRequest)
		return
	}
	key := args[0]
	if key == "" {
		http.Error(w, "Invalid URL format. Expected Delete/{key}", http.StatusBadRequest)
		return
	}

	payload := deletePayload{Key: key}
	data, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to marshal payload", http.StatusInternalServerError)
		return
	}

	future := hs.r.Apply(data, 3*time.Second)
	if err := future.Error(); err != nil {
		http.Error(w, fmt.Sprintf("Could not delete key: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Deleted key '%s'\n", key)
}
