package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/AlexG28/keyvalue/store"
)

func Set(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("r.URL.Path: %v\n", r.URL.Path)
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	fmt.Printf("parts: %v\n", parts)
	if len(parts) >= 3 || parts[0] != "Set" {
		key := parts[1]
		value := parts[2]
		fmt.Fprintf(w, "Key: %s, Value: %s\n", key, value)

		err := store.Add(key, value)

		if err != nil {
			http.Error(w, "Failed to add to store", http.StatusInternalServerError)
		}

	} else {
		http.Error(w, "Invalid URL format. Expected Set/{key}/{value}", http.StatusBadRequest)
	}

	fmt.Fprint(w, http.StatusOK)
}
func Get(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 2 || parts[0] != "Get" {
		key := parts[1]
		fmt.Fprintf(w, "Key: %s\n", key)

		val, err := store.Get(key)

		if err != nil {
			http.Error(w, "Failed to get from store", http.StatusInternalServerError)
		}

		fmt.Fprintf(w, "Value: %s\n", val)

	} else {
		http.Error(w, "Invalid URL format. Expected Get/{key}", http.StatusBadRequest)
	}

	fmt.Fprint(w, http.StatusOK)
}
func Delete(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 2 || parts[0] != "Delete" {
		key := parts[1]
		fmt.Fprintf(w, "Key: %s\n", key)

		err := store.Delete(key)

		if err != nil {
			http.Error(w, "Failed to delete", http.StatusInternalServerError)
		}

	} else {
		http.Error(w, "Invalid URL format. Expected Delete/{key}", http.StatusBadRequest)
	}

	fmt.Fprint(w, http.StatusOK)
}

func main() {
	http.HandleFunc("/Set/", Set)
	http.HandleFunc("/Get/", Get)
	http.HandleFunc("/Delete/", Delete)
	log.Print("Starting on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
