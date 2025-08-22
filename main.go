package main

import (
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/AlexG28/keyvalue/store"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, http.StatusOK)
}

func main() {
	cfg := getConfig()

	localStore := store.InitStore()
	gossipManager, err := NewGossipManager(&cfg)
	if err != nil {
		log.Fatalf("failed to create gossip cluster: %s", err)
	}

	if cfg.existingGossip != "" {
		err = gossipManager.JoinCluster([]string{cfg.joinHost + ":" + cfg.existingGossip})
		if err != nil {
			log.Fatalf("failed to join gossip cluster: %s", err)
		}
	}

	kf := &kvFsm{store: localStore}

	dataDir := "data"
	r, err := setupRaft(path.Join(dataDir, "raft"+cfg.id), cfg.id, cfg.raftPort, kf)
	if err != nil {
		log.Fatalf("setting up Raft failed: %s", err)
	}

	gossipManager.SetRaftNode(r)

	hs := httpServer{r, localStore}

	http.HandleFunc("/Set/", hs.Set)
	http.HandleFunc("/Get/", hs.Get)
	http.HandleFunc("/Delete/", hs.Delete)
	http.HandleFunc("/Health", HealthCheck)
	http.HandleFunc("/isLeader", hs.IsLeader)
	log.Println("Starting on " + cfg.joinHost + ":" + cfg.httpPort)
	log.Fatal(http.ListenAndServe(":"+cfg.httpPort, nil))
}
