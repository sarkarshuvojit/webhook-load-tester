package types

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type RequestTrackerPair struct {
	StartTime time.Time // start
	EndTime   time.Time
}

type DefaultWebhookTester struct {
	targetUrl string
	selfUrl   string

	iters      int
	iterGap    time.Duration
	requestsWg sync.WaitGroup

	reqTracker map[string]RequestTrackerPair
}

func NewDefaultWebhookTester() *DefaultWebhookTester {
	return &DefaultWebhookTester{
		reqTracker: map[string]RequestTrackerPair{},
	}
}

// LoadConfig implements WebhookTester.
func (wt *DefaultWebhookTester) LoadConfig() error {
	wt.targetUrl = "http://localhost:8080/"
	wt.selfUrl = "http://localhost:8081/"

	slog.Info("Loaded config...")
	return nil
}

// InitTestSetup implements WebhookTester.
func (wt *DefaultWebhookTester) InitTestSetup() error {
	wt.iterGap = 500 * time.Millisecond
	wt.iters = 10
	slog.Info("Initialised test setup...")
	return nil
}

func (wt *DefaultWebhookTester) receiverHandler(w http.ResponseWriter, r *http.Request) {
	bytedata, _ := io.ReadAll(r.Body)
	reqBodyStr := string(bytedata)
	slog.Debug("Received New Message", "body", reqBodyStr)
	// pick correlationId
	// save in common concurrent hashmap
	var resMap map[string]string
	err := json.Unmarshal(bytedata, &resMap)
	if err != nil {
		panic(err)
	}

	correlationId := resMap["correlationId"]
	_tracker := wt.reqTracker[correlationId]
	_tracker.EndTime = time.Now()

	wt.reqTracker[correlationId] = _tracker
	wt.requestsWg.Done()
}

// InitReceiver implements WebhookTester.
func (wt *DefaultWebhookTester) InitReceiver() (context.CancelFunc, error) {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(wt.receiverHandler))

	server := &http.Server{Addr: ":8081", Handler: mux}

	cancelContext, stopReceiver := context.WithCancel(context.Background())
	go func() {
		<-cancelContext.Done()

		slog.Info("Shutting down receiver...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Error shutting down server", "err", err)
		}
	}()

	go func() {
		if err := server.ListenAndServe(); err != nil {
			slog.Error("Error shutting down server", "err", err)
		}
	}()

	slog.Info("Initialised receiver...")
	return stopReceiver, nil
}

// InitRequests implements WebhookTester.
func (wt *DefaultWebhookTester) InitRequests() error {
	wt.requestsWg.Add(wt.iters)
	for i := 0; i < wt.iters; i++ {
		correlationId := uuid.New().String()

		wt.reqTracker[correlationId] = RequestTrackerPair{
			StartTime: time.Now(),
		}

		go func() error {
			reqBodyBytes := []byte(`{"correlationId": "` + correlationId + `"}`)
			req, err := http.NewRequest(http.MethodPost, wt.targetUrl, bytes.NewReader(reqBodyBytes))
			if err != nil {
				return err
			}

			req.Header.Add("webhook-reply-to", wt.selfUrl)

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}

			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}
			slog.Debug("Request Sent", "resBody", resBody)
			return nil
		}()

		slog.Debug("Going to sleep")
		time.Sleep(wt.iterGap)
		slog.Debug("Woke up")
	}
	slog.Info("Requests fired...")
	return nil
}

// PostProcess implements WebhookTester.
func (wt *DefaultWebhookTester) PostProcess() error {
	slog.Debug("PostProcess", "data", wt.reqTracker)
	var durations []time.Duration
	s := 0
	max_time := -1
	min_time := -1
	for k := range wt.reqTracker {
		slog.Debug("Processsing", "correlationId", k)
		timeDiff := wt.reqTracker[k].EndTime.Sub(wt.reqTracker[k].StartTime)
		durations = append(durations, timeDiff)
		s = s + int(timeDiff.Seconds())
		if int(timeDiff.Seconds()) > max_time {
			max_time = int(timeDiff.Seconds())
		}
		if int(timeDiff.Seconds()) < min_time || min_time == -1 {
			min_time = int(timeDiff.Seconds())
		}
	}
	fmt.Printf("Max: %ds\n", max_time)
	fmt.Printf("Min: %ds\n", min_time)
	fmt.Printf("Avg: %ds\n", int(s/wt.iters))
	return nil
}

// WaitForResults implements WebhookTester.
func (wt *DefaultWebhookTester) WaitForResults() error {
	slog.Info("Started waiting for it to finish")
	wt.requestsWg.Wait()
	slog.Info("Waiting comes to an end")
	return nil
}

var _ WebhookTester = (*DefaultWebhookTester)(nil)
