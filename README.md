# webhook-load-tester

If you have ever created a webhook based async API, you might have run into issues with integration testing it. 

The main problem in this case is that the code which calls the API never recieves the response in the same block.

The easiest solution is to create a mock server which can display all requests coming to it; using a tool like Postman, Beeceptor, etc. 

But what if you need to load test your apis and well, and get and idea of how long the whole flow takes; i.e. from the point when the client aka another server calling your api till the time your server responds at their webhook url

![Overview](docs/overview.png "Overview of design")

[![asciicast](https://asciinema.org/a/iPDFUjZSNDOpd2o9sgtI9tcpj.svg)](https://asciinema.org/a/iPDFUjZSNDOpd2o9sgtI9tcpj)


## Setting up locally

### Start Dummy Webhook API 

Start a simple webhook api

```bash
$ make start_dummy
```

### Start tests  

Start the actual tester

#### Using Go

```bash
$ go run cmd/runtest/main.go -v -f docs/input-example.yml
$ go run cmd/runtest/main.go -f docs/input-example.yml # for verbose output
```

## TODO

- Use Ngrok instead of go/http to start receiver
- Timeout for watiting to prevent deadlock when downstream fails to reply to few requests
