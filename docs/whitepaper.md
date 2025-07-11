# Webhook Load Tester: Technical Whitepaper

## Abstract

The Webhook Load Tester is a specialized testing framework designed to address the unique challenges of testing and load testing asynchronous webhook-based APIs. Unlike traditional HTTP load testing tools that measure only synchronous request-response cycles, this tool provides end-to-end timing measurements from the initial API call through to the eventual webhook callback, enabling accurate performance analysis of asynchronous systems.

## 1. Introduction

### 1.1 Problem Statement

Modern API architectures increasingly rely on asynchronous communication patterns where the initial API request triggers a background process, and the result is delivered later via a webhook callback. Traditional load testing tools fail to capture the complete transaction flow in such systems, leading to incomplete performance metrics and inadequate testing coverage.

### 1.2 Key Challenges Addressed

- **Temporal Decoupling**: API calls and webhook responses occur at different times, making it difficult to correlate requests with their corresponding callbacks
- **End-to-End Timing**: Measuring the complete flow from initial request to webhook delivery
- **Load Testing Complexity**: Simulating realistic load patterns while capturing webhook responses
- **Development Environment Constraints**: Testing webhooks locally without exposing internal services

## 2. Architecture Overview

### 2.1 Core Components

The Webhook Load Tester consists of five primary components that work together to provide comprehensive testing capabilities:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Layer     │    │  Configuration  │    │  Mock Server    │
│  (cmd/root.go)  │    │   (types/)      │    │ (webhook_tester)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌─────────────────┐    ┌─────────────────┐
         │   Tracker       │    │   Reporter      │
         │  (tracker/)     │    │  (reporter/)    │
         └─────────────────┘    └─────────────────┘
```

### 2.2 Component Responsibilities

#### 2.2.1 CLI Layer (`cmd/`)
- **Command Processing**: Handles user commands (`create`, `run`) via Cobra CLI framework
- **Configuration Loading**: Parses YAML/JSON configuration files
- **Orchestration**: Coordinates the testing workflow

#### 2.2.2 Configuration System (`pkg/types/`)
- **Input Validation**: Validates test configurations and injector/picker paths
- **Type Definitions**: Defines structured configuration formats
- **Path Resolution**: Handles complex JSON path operations for data injection and extraction

#### 2.2.3 Webhook Tester (`pkg/webhook_tester/`)
- **Mock Server Management**: Operates HTTP server to receive webhook callbacks
- **Request Generation**: Fires API requests with injected correlation IDs and webhook URLs
- **Ngrok Integration**: Exposes local mock server via Ngrok tunnels for external testing

#### 2.2.4 Request Tracking (`pkg/tracker/`)
- **Correlation Management**: Maps API requests to webhook responses using unique identifiers
- **Timing Capture**: Records precise timestamps for start and end of transactions
- **Concurrent Access**: Thread-safe operations for handling concurrent requests

#### 2.2.5 Metrics and Reporting (`pkg/reporter/`)
- **Statistical Analysis**: Calculates comprehensive performance metrics
- **Output Generation**: Produces reports in multiple formats (text, stdout)

## 3. Technical Implementation

### 3.1 Request Correlation Mechanism

The core innovation of the Webhook Load Tester lies in its request correlation system:

```go
type RequestTrackerPair struct {
    StartTime time.Time // Initial API call timestamp
    EndTime   time.Time // Webhook response timestamp
}
```

#### 3.1.1 Correlation ID Injection

The system uses configurable injectors to embed correlation IDs into outgoing requests:

```yaml
injectors:
  correlationIdInjector:
    path: "body.uniqueId"  # JSON path for ID placement
  replyPathInjector:
    path: "headers.webhook-reply-to"  # Webhook URL injection
```

#### 3.1.2 Response Correlation

When webhook responses arrive, the system extracts correlation IDs using configurable pickers:

```yaml
pickers:
  correlationPicker:
    path: "body.uniqueId"  # JSON path for ID extraction
```

### 3.2 Timing and Load Distribution

#### 3.2.1 Request Scheduling

The system calculates precise inter-request intervals to achieve desired load patterns:

```go
iterGap := time.Duration((config.Run.DurationSeconds*1000)/config.Run.Iterations) * time.Millisecond
```

This ensures even distribution of requests over the specified duration, enabling accurate requests-per-second calculations.

#### 3.2.2 Concurrent Request Handling

The implementation uses Go's `sync.WaitGroup` to manage concurrent request tracking:

```go
wt.internal.requestWg.Add(wt.config.Run.Iterations)
// ... fire requests concurrently
wt.internal.requestWg.Wait() // Wait for all webhook responses
```

### 3.3 Network Configuration Options

#### 3.3.1 Local HTTP Server Mode
- Operates on `localhost:8081`
- Suitable for testing local services
- Direct HTTP communication without external dependencies

#### 3.3.2 Ngrok Integration Mode
- Exposes local server via Ngrok tunnel
- Enables testing with external services
- Requires `NGROK_AUTHTOKEN` environment variable
- Automatic tunnel URL injection into test requests

### 3.4 Configuration Schema

The system uses a structured YAML configuration format:

```yaml
version: v1
server: ngrok  # Optional: "ngrok" for external exposure

test:
  name: test-api-1
  url: http://localhost:8080/
  body: "{\"message\": \"ok\"}"
  timeout: 60
  headers:
    client-id: example-id
    client-secret: example-secret
  
  injectors:
    replyPathInjector:
      path: "headers.webhook-reply-to"
    correlationIdInjector:
      path: "body.uniqueId"
  
  pickers:
    correlationPicker:
      path: "body.uniqueId"

run:
  iterations: 1000
  durationSeconds: 10

outputs:
  - type: stdout
  - type: text
    path: results.txt
```

### 3.5 Metrics Calculation

The reporter component calculates comprehensive performance metrics:

```go
type Metrics struct {
    TotalRequests       int
    TotalDuration       time.Duration
    AverageResponseTime time.Duration
    MinResponseTime     time.Duration
    MaxResponseTime     time.Duration
    MedianResponseTime  time.Duration
    Percentile95Time    time.Duration
    RequestsPerSecond   float64
}
```

#### 3.5.1 Statistical Analysis
- **Sorting**: Response times are sorted for percentile calculations
- **Median Calculation**: Uses proper median calculation for even/odd data sets
- **95th Percentile**: Provides insight into tail latency performance
- **Throughput Calculation**: Measures actual requests per second based on execution timeframe

## 4. Security and Best Practices

### 4.1 Environment Variable Usage
- Sensitive data (Ngrok auth tokens) are read from environment variables
- Configuration files should not contain secrets
- Headers support for authentication tokens and API keys

### 4.2 Resource Management
- Proper goroutine lifecycle management
- HTTP server graceful shutdown implementation
- Context-based cancellation for clean termination

### 4.3 Error Handling
- Comprehensive error types defined in `pkg/types/errors.go`
- Timeout mechanisms prevent indefinite waiting
- Graceful degradation when partial results are available

## 5. Use Cases and Applications

### 5.1 Payment Processing APIs
Testing payment webhooks that notify merchants of transaction status changes with realistic load patterns.

### 5.2 Notification Systems
Load testing push notification services that use webhooks to confirm delivery status.

### 5.3 Data Processing Pipelines
Testing ETL systems where webhook callbacks signal completion of data processing jobs.

### 5.4 Integration Testing
Validating webhook reliability and timing in CI/CD pipelines.

## 6. Performance Characteristics

### 6.1 Scalability Limits
- Memory usage scales linearly with concurrent requests
- CPU usage primarily bounded by JSON parsing and HTTP operations
- Network bandwidth limits based on request payload sizes

### 6.2 Timing Accuracy
- Microsecond-precision timing using Go's `time.Now()`
- Minimal overhead from correlation tracking
- Separate goroutines prevent blocking during request generation

## 7. Extensibility and Future Enhancements

### 7.1 Plugin Architecture Potential
The interface-based design (`WebhookTester` interface) enables:
- Custom webhook tester implementations
- Alternative reporter formats
- Additional server backends beyond HTTP and Ngrok

### 7.2 Configuration Enhancements
- Support for dynamic payload generation
- Multiple webhook endpoint testing
- Advanced load patterns (ramp-up, burst testing)

### 7.3 Monitoring Integration
- OpenTelemetry instrumentation potential
- Real-time metrics streaming
- Integration with monitoring systems

## 8. Technical Dependencies

### 8.1 Core Dependencies
- **Go 1.21+**: Core runtime and standard library
- **Cobra**: CLI framework for command structure
- **Ngrok SDK**: Tunnel creation and management
- **UUID**: Correlation ID generation
- **YAML**: Configuration file parsing

### 8.2 Development Dependencies
- Standard Go testing framework
- JSON/YAML processing libraries
- HTTP client and server implementations

## 9. Installation and Deployment

### 9.1 Installation Options
```bash
# Direct installation
go install github.com/sarkarshuvojit/webhook-load-tester@latest

# Local development
git clone <repository>
go build -o webhook-load-tester
```

### 9.2 Configuration Setup
```bash
# Create configuration template
webhook-load-tester create -c test-config.yaml

# Execute test
webhook-load-tester run -c test-config.yaml
```

## 10. Conclusion

The Webhook Load Tester addresses a significant gap in the API testing ecosystem by providing purpose-built tooling for asynchronous webhook-based systems. Its correlation-based approach to timing measurement, combined with flexible configuration options and robust error handling, makes it suitable for both development and production performance testing scenarios.

The tool's architecture demonstrates several key design principles:
- **Separation of Concerns**: Clear boundaries between configuration, execution, and reporting
- **Extensibility**: Interface-based design enabling future enhancements
- **Reliability**: Comprehensive error handling and timeout mechanisms
- **Usability**: Simple YAML configuration with sensible defaults

As API architectures continue to evolve toward asynchronous patterns, tools like the Webhook Load Tester become essential components of the testing toolkit, enabling developers to accurately measure and optimize the performance of their webhook-based systems.

---

*This whitepaper provides a comprehensive technical overview of the Webhook Load Tester as of the current implementation. For the latest updates and enhancements, please refer to the project repository.*