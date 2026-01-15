package webhook_tester

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
	"github.com/sarkarshuvojit/webhook-load-tester/pkg/reporter"
	"github.com/sarkarshuvojit/webhook-load-tester/pkg/tracker"
	"github.com/sarkarshuvojit/webhook-load-tester/pkg/types"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
)

var DEFAULT_WAITING_TIMEOUT = time.Duration(30) * time.Second

type internalConfig struct {
	targetUrl     string
	selfUrl       string
	selfUrlChan   chan string
	requestsFired chan bool

	iterGap    time.Duration
	requestWg  sync.WaitGroup
	reqTracker *tracker.Tracker
}

type DefaultWebhookTester struct {
	config   *types.InputConfig
	internal *internalConfig
}

func (wt *DefaultWebhookTester) receiverHandler(w http.ResponseWriter, r *http.Request) {
	bytedata, _ := io.ReadAll(r.Body)
	reqBodyStr := string(bytedata)
	slog.Debug("Received New Message", "body", reqBodyStr)
	// pick correlationId
	// save in common concurrent hashmap
	var resMap map[string]any
	err := json.Unmarshal(bytedata, &resMap)
	if err != nil {
		panic(err)
	}
	var correlationId string
	if wt.config.Test.Pickers.CorrelationPicker.GetRootType() == types.RootBody {
		// TODO: Add error handling here, what if it's not found?
		correlationId = (*wt.config.Test.Pickers.CorrelationPicker.GetByLocator(&resMap)).(string)
	}
	_tracker := wt.internal.reqTracker.Get(correlationId)
	_tracker.EndTime = time.Now()

	slog.Debug("Updating tracker", "key", correlationId, "value", _tracker)
	wt.internal.reqTracker.Set(correlationId, _tracker)
	wt.internal.requestWg.Done()
}

// FireRequests implements WebhookTesterv2.
func (wt *DefaultWebhookTester) FireRequests() error {
	wt.internal.requestWg.Add(wt.config.Run.Iterations)
	slog.Info("Waiting for server to be ready")
	serverURL := <-wt.internal.selfUrlChan
	slog.Debug("Server ready", "addr", serverURL)

	for i := 0; i < wt.config.Run.Iterations; i++ {
		correlationId := uuid.New().String()

		wt.internal.reqTracker.Set(correlationId, tracker.RequestTrackerPair{
			StartTime: time.Now(),
		})

		go func() error {
			var tmp map[string]any
			reqBodyBytes := []byte(wt.config.Test.Body)
			err := json.Unmarshal(reqBodyBytes, &tmp)
			if err != nil {
				slog.Error("Failed to call api", "err", err)
				return err
			}

			injectors := wt.config.Test.Injectors

			if injectors.CorrelationIDInjector.GetRootType() == types.RootBody {
				slog.Debug("Setting correlationId to body", "key", injectors.CorrelationIDInjector.GetKey())
				injectors.CorrelationIDInjector.SetToLocator(
					&tmp,
					correlationId,
				)
			}

			if injectors.ReplyPathInjector.GetRootType() == types.RootBody {
				slog.Debug("Setting replyPath to body", "key", injectors.ReplyPathInjector.GetKey())
				injectors.ReplyPathInjector.SetToLocator(
					&tmp,
					wt.internal.selfUrl,
				)
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

			if injectors.CorrelationIDInjector.GetRootType() == types.RootHeader {
				slog.Debug("Setting correlation to header")
				req.Header.Add(
					injectors.CorrelationIDInjector.GetKey(),
					correlationId,
				)
			}

			if injectors.ReplyPathInjector.GetRootType() == types.RootHeader {
				slog.Debug(
					"Setting replyPath to header",
					"key", injectors.ReplyPathInjector.GetKey(),
				)
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
			slog.Debug("Request Sent", "req", req, "resBody", resBody)
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
func (wt *DefaultWebhookTester) LoadConfig() error {
	slog.Info("Loading and validating config...")
	injectors := wt.config.Test.Injectors
	if corrRootType := injectors.CorrelationIDInjector.GetRootType(); corrRootType == types.RootUnknown {
		return errors.New("Unknown root type: " + injectors.CorrelationIDInjector.GetRootTypeString())
	}
	if replyPathRootType := injectors.ReplyPathInjector.GetRootType(); replyPathRootType == types.RootUnknown {
		return errors.New("Unknown root type: " + injectors.ReplyPathInjector.GetRootTypeString())
	}

	pickers := wt.config.Test.Pickers
	if corrPickerRt := pickers.CorrelationPicker.GetRootType(); corrPickerRt == types.RootUnknown {
		return errors.New("Unknown root type: " + pickers.CorrelationPicker.GetRootTypeString())
	}

	return nil
}

// PostProcess implements WebhookTesterv2.
func (wt *DefaultWebhookTester) PostProcess() error {
	allReqs := wt.internal.reqTracker.GetAll()
	tp := []tracker.RequestTrackerPair{}
	for _, v := range allReqs {
		tp = append(tp, v)
	}

	metrics := reporter.CalculateMetrics(tp, time.Duration(wt.config.Run.DurationSeconds)*time.Second)

	for _, output := range wt.config.Outputs {
		switch output.Type {
		case "text":
			w, err := createFileWithParentDirs(output.Path)
			defer w.Close()
			if err != nil {
				return err
			}
			reporter.PrintTextMetrics(w, metrics)
		case "stdout":
			reporter.PrintTextMetrics(os.Stdout, metrics)
		default:
			return types.UnsupportedOutputErr
		}
	}

	return nil
}

func (wt *DefaultWebhookTester) startHttpServer() (context.CancelFunc, error) {
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

	wt.internal.selfUrl = "http://localhost:8081/"
	wt.internal.selfUrlChan <- "http://localhost:8081/"
	slog.Info("Initialised receiver...")
	return stopReceiver, nil
}

func (wt *DefaultWebhookTester) startNgrokServer() (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())

	if _, found := os.LookupEnv("NGROK_AUTHTOKEN"); !found {
		return cancel, types.NgrokAuthMissingErr
	}

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
			slog.Error("Error setting up ngrok:", "error", err)
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
func (wt *DefaultWebhookTester) StartReceiver() (cancelFunc context.CancelFunc, err error) {
	if wt.config.Server == "ngrok" {
		cancelFunc, err = wt.startNgrokServer()
	} else {
		cancelFunc, err = wt.startHttpServer()
	}
	return cancelFunc, err
}

// WaitForResults implements WebhookTesterv2.
func (wt *DefaultWebhookTester) WaitForResults() error {
	timeout := time.Duration(wt.config.Test.Timeout) * time.Second
	slog.Info("Waiting for results...", "timeout", timeout)
	waitingFinished := make(chan bool)
	go func() {
		wt.internal.requestWg.Wait()
		waitingFinished <- true
	}()
	select {
	case <-waitingFinished:
		slog.Info("Finished waiting within timeout")
	case <-time.After(timeout):
		slog.Error("Timed out while waiting for 2mins")
		return types.TimedOutWaitingForResultsErr
	}
	return nil
}

func NewDefaultWebhookTester(config *types.InputConfig) *DefaultWebhookTester {
	wt := &DefaultWebhookTester{
		config: config,
	}
	wt.setup()

	return wt
}

func (wt2 *DefaultWebhookTester) setup() {
	iterGap := time.Duration((wt2.config.Run.DurationSeconds*1000)/wt2.config.Run.Iterations) * time.Millisecond
	wt2.internal = &internalConfig{
		iterGap:       iterGap,
		requestWg:     sync.WaitGroup{},
		reqTracker:    tracker.NewRequestTracker(),
		selfUrlChan:   make(chan string, 1),
		requestsFired: make(chan bool, 1),
	}

	if wt2.config.Test.Timeout == 0 {
		wt2.config.Test.Timeout = int(DEFAULT_WAITING_TIMEOUT.Seconds())
	}

	configStr, _ := json.MarshalIndent(wt2.config, "", "  ")
	slog.Debug(string(configStr))
}

var _ WebhookTester = (*DefaultWebhookTester)(nil)
