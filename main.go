package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/AlexG28/keyvalue/store"
)

var localStore = store.InitStore()

func parsePath(r *http.Request) (cmd string, args []string) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

func Set(w http.ResponseWriter, r *http.Request) {
	cmd, args := parsePath(r)
	if cmd != "Set" || len(args) < 2 {
		http.Error(w, "Invalid URL format. Expected Set/{key}/{value}", http.StatusBadRequest)
		return
	}
	key, value := args[0], args[1]
	if key == "" {
		http.Error(w, "Missing Key", http.StatusBadRequest)
		return
	}
	if value == "" {
		http.Error(w, "Missing Value", http.StatusBadRequest)
		return
	}
	if err := localStore.Add(key, value); err != nil {
		http.Error(w, "Failed to add to store", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Set key '%s' to value '%s'\n", key, value)
}

func Get(w http.ResponseWriter, r *http.Request) {
	cmd, args := parsePath(r)
	if cmd != "Get" || len(args) < 1 {
		http.Error(w, "Invalid URL format. Expected Get/{key}", http.StatusBadRequest)
		return
	}
	key := args[0]
	val, err := localStore.Get(key)
	if err != nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "%s", val)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	cmd, args := parsePath(r)
	if cmd != "Delete" || len(args) < 1 {
		http.Error(w, "Invalid URL format. Expected Delete/{key}", http.StatusBadRequest)
		return
	}
	key := args[0]
	if err := localStore.Delete(key); err != nil {
		http.Error(w, "Failed to delete key", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Deleted key '%s'\n", key)
}

func main() {
	http.HandleFunc("/Set/", Set)
	http.HandleFunc("/Get/", Get)
	http.HandleFunc("/Delete/", Delete)
	log.Print("Starting on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
