package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"time"
)

func sleepRandomly(minSleep, maxSleep int) {
	randSleep := rand.Intn(maxSleep)
	slog.Info("Sleeping", "ttl", minSleep+randSleep)
	time.Sleep(time.Duration(minSleep+randSleep) * time.Second)
	slog.Info("Slept", "ttl", minSleep+randSleep)
}

func dumpHttpDetails(r *http.Request) {
	// Log request headers
	fmt.Println("Request Headers:")
	for name, values := range r.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", name, value)
		}
	}

	// Log request body
	fmt.Println("Request Body:")
	// Read and log the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading body:", err)
		return
	}
	fmt.Println(string(body))
	r.Body = io.NopCloser(bytes.NewReader(body))
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Read the `webhook-reply-to` header
	dumpHttpDetails(r)
	replyToURL := r.Header.Get("webhook-reply-to")
	if replyToURL == "" {
		http.Error(w, "Missing 'webhook-reply-to' header", http.StatusBadRequest)
		return
	}

	_clientId := r.Header.Get("client-id")
	if _clientId == "" {
		http.Error(w, "Missing 'client-id' header", http.StatusBadRequest)
		return
	}

	_secret := r.Header.Get("client-secret")
	if _secret == "" {
		http.Error(w, "Missing 'client-secret' header", http.StatusBadRequest)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	slog.Debug("Request Body:", "body", body)

	go func() {
		sleepRandomly(5, 10)
		// Forward the request body to the `webhook-reply-to` URL with a POST request
		slog.Info("Sending response", "replyUrl", replyToURL, "body", body)
		resp, err := http.Post(replyToURL, r.Header.Get("Content-Type"), bytes.NewReader(body))
		if err != nil {
			http.Error(w, "Failed to forward request", http.StatusInternalServerError)
			log.Printf("Error forwarding request to %s: %v", replyToURL, err)
			return
		}
		defer resp.Body.Close()
	}()

	// Send back the response from the forwarded request
	w.WriteHeader(200)
	w.Write([]byte(`{"message": "request accepted"}`))
}

func main() {
	// Handle all routes
	http.HandleFunc("/", handler)

	// Start the HTTP server
	fmt.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
