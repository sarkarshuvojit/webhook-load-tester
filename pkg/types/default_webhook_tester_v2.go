package types

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type internalConfig struct {
	targetUrl string
	selfUrl   string

	iterGap    time.Duration
	requestWg  sync.WaitGroup
	reqTracker map[string]RequestTrackerPair
}

type DefaultWebhookTesterv2 struct {
	config         *InputConfig
	internalConfig *internalConfig
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
	_tracker := wt.internalConfig.reqTracker[correlationId]
	_tracker.EndTime = time.Now()

	slog.Debug("Updating tracker", "key", correlationId, "value", _tracker)
	wt.internalConfig.reqTracker[correlationId] = _tracker
	wt.internalConfig.requestWg.Done()
}

// FireRequests implements WebhookTesterv2.
func (wt *DefaultWebhookTesterv2) FireRequests() error {
	wt.internalConfig.requestWg.Add(wt.config.Run.Iterations)

	for i := 0; i < wt.config.Run.Iterations; i++ {
		correlationId := uuid.New().String()

		wt.internalConfig.reqTracker[correlationId] = RequestTrackerPair{
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
				tmp[injectors.ReplyPathInjector.GetKey()] = wt.internalConfig.selfUrl
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

			if injectors.CorrelationIDInjector.GetRootType() == RootHeader {
				slog.Debug("Setting correlation to header")
				req.Header.Add(injectors.CorrelationIDInjector.GetKey(), correlationId)
			}

			slog.Debug("ReplyPathInjector", "path", injectors.ReplyPathInjector.Path)
			if injectors.ReplyPathInjector.GetRootType() == RootHeader {
				slog.Debug("Setting replyPath to header", "key", injectors.ReplyPathInjector.GetKey())
				req.Header.Add(injectors.ReplyPathInjector.GetKey(), wt.internalConfig.selfUrl)
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
		time.Sleep(wt.internalConfig.iterGap)
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

// StartReceiver implements WebhookTesterv2.
func (wt *DefaultWebhookTesterv2) StartReceiver() (context.CancelFunc, error) {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(wt.receiverHandler))

	server := &http.Server{Addr: ":8081", Handler: mux}
	wt.internalConfig.selfUrl = "http://localhost:8081/"

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

// WaitForResults implements WebhookTesterv2.
func (wt *DefaultWebhookTesterv2) WaitForResults() error {
	slog.Info("Waiting for results...")
	wt.internalConfig.requestWg.Wait()
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
	wt2.internalConfig = &internalConfig{
		selfUrl:    "http://localhost:8081/",
		iterGap:    iterGap,
		requestWg:  sync.WaitGroup{},
		reqTracker: map[string]RequestTrackerPair{},
	}
}

var _ WebhookTesterv2 = (*DefaultWebhookTesterv2)(nil)
