package main

import (
	"fmt"
	"log"
	"net/http"
	"path"

	"github.com/AlexG28/keyvalue/store"
)

// var localStore = store.InitStore()

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, http.StatusOK)
}

func main() {

	cfg := getConfig()

	localStore := store.InitStore()

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
