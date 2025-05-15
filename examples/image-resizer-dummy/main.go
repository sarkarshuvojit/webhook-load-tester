package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"time"
)

type ResizeRequest struct {
	ImageURL string `json:"image_url"`
	Size     string `json:"size"`
}

type ResizeResponse struct {
	CorrelationID string `json:"correlation_id"`
	Message       string `json:"message"`
	ResizedURL    string `json:"resized_url"`
}

var cdnDomains = []string{
	"https://cdn.imgzr.net/",
	"https://fastpix.cdn.io/",
	"https://assets.rszr.cloud/",
	"https://images.storix.ai/",
}

func sleepRandomly(minSleep, maxSleep int) {
	randSleep := rand.Intn(maxSleep)
	time.Sleep(time.Duration(minSleep+randSleep) * time.Second)
}

func generateFakeURL(size string) string {
	prefix := cdnDomains[rand.Intn(len(cdnDomains))]
	return fmt.Sprintf("%sprocessed/img_%d_%s.jpg", prefix, time.Now().Unix(), size)
}

func handler(w http.ResponseWriter, r *http.Request) {
	replyToURL := r.Header.Get("webhook-reply-to")
	if replyToURL == "" {
		http.Error(w, "Missing 'webhook-reply-to' header", http.StatusBadRequest)
		return
	}

	correlationID := r.Header.Get("correlation-id")
	if correlationID == "" {
		http.Error(w, "Missing 'correlation-id' header", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	slog.Info("Received image resize request", "correlation-id", correlationID)

	go func() {
		sleepRandomly(5, 10)

		var resizeReq ResizeRequest
		if err := json.Unmarshal(body, &resizeReq); err != nil {
			log.Printf("[correlation-id: %s] Invalid JSON payload", correlationID)
			return
		}

		slog.Info("Processing image resize", "correlation-id", correlationID, "image_url", resizeReq.ImageURL, "size", resizeReq.Size)

		// Fake processing
		sleepRandomly(1, 2)

		resizedURL := generateFakeURL(resizeReq.Size)

		res := ResizeResponse{
			CorrelationID: correlationID,
			Message:       "Image resized successfully",
			ResizedURL:    resizedURL,
		}

		resBody, err := json.Marshal(res)
		if err != nil {
			log.Printf("[correlation-id: %s] Failed to marshal response: %v", correlationID, err)
			return
		}

		slog.Info("Sending response", "replyTo", replyToURL, "resized_url", resizedURL)

		resp, err := http.Post(replyToURL, r.Header.Get("Content-Type"), bytes.NewReader(resBody))
		if err != nil {
			log.Printf("[correlation-id: %s] Failed to POST response: %v", correlationID, err)
			return
		}
		defer resp.Body.Close()
	}()

	w.WriteHeader(200)
	w.Write([]byte(`{"message": "request accepted"}`))
}

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/", handler)

	fmt.Println("Fake Image Resizer listening on :9000...")
	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
