package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/AlexG28/keyvalue/store"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, http.StatusOK)
}

func main() {

	cfg := getConfig()

	localStore := store.InitStore()

	gossipManager, err := NewGossipManager(&cfg)
	fmt.Println("Created gossip mananger")
	if err != nil {
		log.Fatalf("failed to create gossip cluster: %s", err)
	}

	if cfg.existingGossip != "" {
		err = gossipManager.JoinCluster([]string{"localhost:" + cfg.existingGossip})

		if err != nil {
			log.Fatalf("failed to join gossip cluster: %s", err)
		}
		fmt.Println("Successfully joined cluster")
	}

	go func() {
		for {
			fmt.Println("Current members: ")
			for _, member := range gossipManager.memberlist.Members() {
				fmt.Printf(" Name: %s. Address %s\n", member.Name, member.Addr)
			}
			time.Sleep(time.Second * 5)
		}
	}()

	kf := &kvFsm{store: localStore}

	dataDir := "data"
	r, err := setupRaft(path.Join(dataDir, "raft"+cfg.id), cfg.id, "localhost:"+cfg.raftPort, kf)
	if err != nil {
		log.Fatalf("something went wrong in main: %s", err)
	}

	hs := httpServer{r, localStore}

	http.HandleFunc("/Set/", hs.Set)
	http.HandleFunc("/Get/", hs.Get)
	http.HandleFunc("/Delete/", hs.Delete)
	http.HandleFunc("/Join", hs.Join)
	http.HandleFunc("/Leader", hs.IsLeader)
	http.HandleFunc("/Health", HealthCheck)
	log.Println("Starting on localhost:" + cfg.httpPort)
	log.Fatal(http.ListenAndServe(":"+cfg.httpPort, nil))
}
