# Webhook Relay Server

This is a simple Go HTTP server that receives POST requests with specific headers, optionally logs their content, and forwards the request body to another URL specified in the request headers. It introduces a randomized delay before forwarding to simulate asynchronous processing.

## Features

- Validates required headers:
  - `webhook-reply-to`: Target URL to forward the request to
  - `client-id` and `client-secret`: Required for authentication/authorization validation (not enforced but required)
- Reads and forwards the request body to the specified webhook URL
- Introduces a random delay (5–10 seconds) before forwarding
- Sends an immediate 200 OK response to the sender while asynchronously forwarding the data

## Usage

### Run the Server

```bash
go run main.go

