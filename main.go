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

	// Create a single store that will be shared between Raft FSM and HTTP handlers
	tempStore := store.InitStore()

	// Create FSM that uses the same store
	kf := &kvFsm{store: tempStore}

	dataDir := "data"
	r, err := setupRaft(path.Join(dataDir, "raft"+cfg.id), cfg.id, "localhost:"+cfg.raftPort, kf)
	if err != nil {
		log.Fatalf("something went wrong in main: %s", err)
	}

	hs := httpServer{r, tempStore}

	http.HandleFunc("/Set/", hs.Set)
	http.HandleFunc("/Get/", hs.Get)
	http.HandleFunc("/Delete/", hs.Delete)
	http.HandleFunc("/Join", hs.Join)
	http.HandleFunc("/Health", HealthCheck)
	log.Println("Starting on localhost:" + cfg.httpPort)
	log.Fatal(http.ListenAndServe(":"+cfg.httpPort, nil))

	// db := &sync.Map{}
	// kf := &kvFsm{db}

	// dataDir := "data"
	// r, err := setupRaft(path.Join(dataDir, "raft"+cfg.id), cfg.id, "localhost:"+cfg.raftPort, kf)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// hs := httpServer{r, db}

	// http.HandleFunc("/set", hs.setHandler)
	// http.HandleFunc("/get", hs.getHandler)
	// http.HandleFunc("/join", hs.joinHandler)
	// http.ListenAndServe(":"+cfg.httpPort, nil)

	// http.HandleFunc("/Set/", Set)
	// http.HandleFunc("/Get/", Get)
	// http.HandleFunc("/Delete/", Delete)
	// http.HandleFunc("/Health", HealthCheck)
	// log.Print("Starting on localhost:8080")
	// log.Fatal(http.ListenAndServe(":8080", nil))
}
