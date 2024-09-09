package types

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

type internalConfig struct {
	targetUrl   string
	selfUrl     string
	selfUrlChan chan string

	iterGap    time.Duration
	requestWg  sync.WaitGroup
	reqTracker map[string]RequestTrackerPair
}

type DefaultWebhookTesterv2 struct {
	config   *InputConfig
	internal *internalConfig
}

func (wt *DefaultWebhookTesterv2) receiverHandler(w http.ResponseWriter, r *http.Request) {
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
	var correlationId string
	if wt.config.Test.Pickers.CorrelationPicker.GetRootType() == RootBody {
		correlationId = resMap[wt.config.Test.Pickers.CorrelationPicker.GetKey()]
	}
	_tracker := wt.internal.reqTracker[correlationId]
	_tracker.EndTime = time.Now()

	slog.Debug("Updating tracker", "key", correlationId, "value", _tracker)
	wt.internal.reqTracker[correlationId] = _tracker
	wt.internal.requestWg.Done()
}

// FireRequests implements WebhookTesterv2.
func (wt *DefaultWebhookTesterv2) FireRequests() error {
	wt.internal.requestWg.Add(wt.config.Run.Iterations)
	slog.Info("Waiting for server to be ready")
	serverURL := <-wt.internal.selfUrlChan
	slog.Debug("Server ready", "addr", serverURL)

	for i := 0; i < wt.config.Run.Iterations; i++ {
		correlationId := uuid.New().String()

		wt.internal.reqTracker[correlationId] = RequestTrackerPair{
			StartTime: time.Now(),
		}

		go func() error {
			var tmp map[string]interface{}
			reqBodyBytes := []byte(wt.config.Test.Body)
			err := json.Unmarshal(reqBodyBytes, &tmp)
			if err != nil {
				slog.Error("Failed to call api", "err", err)
				return err
			}

			injectors := wt.config.Test.Injectors

			if injectors.CorrelationIDInjector.GetRootType() == RootBody {
				slog.Debug("Setting correlationId to body", "key", injectors.CorrelationIDInjector.GetKey())
				tmp[injectors.CorrelationIDInjector.GetKey()] = correlationId
			}

			if injectors.ReplyPathInjector.GetRootType() == RootBody {
				slog.Debug("Setting replyPath to body", "key", injectors.ReplyPathInjector.GetKey())
				tmp[injectors.ReplyPathInjector.GetKey()] = wt.internal.selfUrl
			}

			reqBodyBytes, err = json.Marshal(tmp)
			if err != nil {
				slog.Error("Failed to create json from interface", "err", err)
				return err
			}

			req, err := http.NewRequest(
				http.MethodPost,
				wt.config.Test.URL,
				bytes.NewReader(reqBodyBytes),
			)

			if err != nil {
				return err
			}

			// Add Test related custom headers
			customheaders := wt.config.Test.Headers
			if len(wt.config.Test.Headers) != 0 {
				for k, v := range customheaders {
					req.Header.Add(k, v)
				}
			}

			if injectors.CorrelationIDInjector.GetRootType() == RootHeader {
				slog.Debug("Setting correlation to header")
				req.Header.Add(injectors.CorrelationIDInjector.GetKey(), correlationId)
			}

			slog.Debug("ReplyPathInjector", "path", injectors.ReplyPathInjector.Path)
			if injectors.ReplyPathInjector.GetRootType() == RootHeader {
				slog.Debug("Setting replyPath to header", "key", injectors.ReplyPathInjector.GetKey())
				req.Header.Add(injectors.ReplyPathInjector.GetKey(), wt.internal.selfUrl)
			}

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
		time.Sleep(wt.internal.iterGap)
		slog.Debug("Woke up")
	}
	slog.Info("Requests fired...")
	return nil
}

// LoadConfig implements WebhookTesterv2.
// validate and throw results
func (wt *DefaultWebhookTesterv2) LoadConfig() error {
	slog.Info("Loading and validating config...")
	injectors := wt.config.Test.Injectors
	if corrRootType := injectors.CorrelationIDInjector.GetRootType(); corrRootType == RootUnknown {
		return errors.New("Unknown root type: " + injectors.CorrelationIDInjector.GetRootTypeString())
	}
	if replyPathRootType := injectors.ReplyPathInjector.GetRootType(); replyPathRootType == RootUnknown {
		return errors.New("Unknown root type: " + injectors.ReplyPathInjector.GetRootTypeString())
	}

	pickers := wt.config.Test.Pickers
	if corrPickerRt := pickers.CorrelationPicker.GetRootType(); corrPickerRt == RootUnknown {
		return errors.New("Unknown root type: " + pickers.CorrelationPicker.GetRootTypeString())
	}

	return nil
}

// PostProcess implements WebhookTesterv2.
func (*DefaultWebhookTesterv2) PostProcess() error {
	panic("unimplemented")
}

func (wt *DefaultWebhookTesterv2) startHttpServer() (context.CancelFunc, error) {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(wt.receiverHandler))

	server := &http.Server{Addr: ":8081", Handler: mux}
	wt.internal.selfUrl = "http://localhost:8081/"

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

func (wt *DefaultWebhookTesterv2) startNgrokServer() (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		// Handle system interrupts and termination signals
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigCh)

		// Initialize the ngrok listener
		listener, err := ngrok.Listen(ctx,
			config.HTTPEndpoint(),
			ngrok.WithAuthtokenFromEnv(),
		)
		if err != nil {
			slog.Error("Error setting up ngrok:", err)
			os.Exit(1)
			return
		}

		slog.Debug("Ingress established at:", "addr", listener.URL())
		wt.internal.selfUrl = listener.URL()
		wt.internal.selfUrlChan <- listener.URL()
		// Start HTTP server
		err = http.Serve(listener, http.HandlerFunc(wt.receiverHandler))
		if err != nil {
			log.Println("HTTP server error:", err)
		}

		// Listen for system interrupts
		select {
		case <-sigCh:
			log.Println("Received shutdown signal")
		case <-ctx.Done():
			log.Println("Context cancelled")
		}
	}()

	slog.Info("Initialised receiver...")
	return cancel, nil
}

// StartReceiver implements WebhookTesterv2.
func (wt *DefaultWebhookTesterv2) StartReceiver() (cancelFunc context.CancelFunc, err error) {
	if wt.config.Server == "ngrok" {
		cancelFunc, err = wt.startNgrokServer()
	} else {
		cancelFunc, err = wt.startHttpServer()
	}
	return
}

// WaitForResults implements WebhookTesterv2.
func (wt *DefaultWebhookTesterv2) WaitForResults() error {
	slog.Info("Waiting for results...")
	timeout := time.Duration(1) * time.Minute
	waitingFinished := make(chan bool)
	go func() {
		wt.internal.requestWg.Wait()
		waitingFinished <- true
	}()
	select {
	case <-waitingFinished:
		slog.Info("Finished waiting within timeout")
	case <-time.After(timeout):
		slog.Info("Timed out while waiting for 2mins")
	}
	return nil
}

func NewDefaultWebhookTesterv2(config *InputConfig) *DefaultWebhookTesterv2 {
	wt := &DefaultWebhookTesterv2{
		config: config,
	}
	wt.setup()

	return wt
}

func (wt2 *DefaultWebhookTesterv2) setup() {
	iterGap := time.Duration((wt2.config.Run.DurationSeconds*1000)/wt2.config.Run.Iterations) * time.Millisecond
	wt2.internal = &internalConfig{
		iterGap:     iterGap,
		requestWg:   sync.WaitGroup{},
		reqTracker:  map[string]RequestTrackerPair{},
		selfUrlChan: make(chan string, 1),
	}
}

var _ WebhookTesterv2 = (*DefaultWebhookTesterv2)(nil)
