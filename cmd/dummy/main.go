package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	// Handle all routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Read the `webhook-reply-to` header
		replyToURL := r.Header.Get("webhook-reply-to")
		if replyToURL == "" {
			http.Error(w, "Missing 'webhook-reply-to' header", http.StatusBadRequest)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// Forward the request body to the `webhook-reply-to` URL with a POST request
		resp, err := http.Post(replyToURL, r.Header.Get("Content-Type"), bytes.NewReader(body))
		if err != nil {
			http.Error(w, "Failed to forward request", http.StatusInternalServerError)
			log.Printf("Error forwarding request to %s: %v", replyToURL, err)
			return
		}
		defer resp.Body.Close()

		// Send back the response from the forwarded request
		w.WriteHeader(resp.StatusCode)
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read response from forwarded request", http.StatusInternalServerError)
			return
		}
		w.Write(respBody)
	})

	// Start the HTTP server
	fmt.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
