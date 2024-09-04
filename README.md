# webhook-load-tester

If you have ever created a webhook based async API, you might have run into issues with integration testing it. 

The main problem in this case is that the code which calls the API never recieves the response in the same block.

The easiest solution is to create a mock server which can display all requests coming to it; using a tool like Postman, Beeceptor, etc. 

But what if you need to load test your apis and well, and get and idea of how long the whole flow takes; i.e. from the point when the client aka another server calling your api till the time your server responds at their webhook url

![Overview](docs/overview.png "Overview of design")

[![asciicast](https://asciinema.org/a/iPDFUjZSNDOpd2o9sgtI9tcpj.svg)](https://asciinema.org/a/iPDFUjZSNDOpd2o9sgtI9tcpj)

## Configuration 

It uses configuration files to run tests, where tests & load can be defined using a declarative yaml format.

Following is an example test config which can be used against the dummy webhook api.

```yaml
version: v1

test:
  name: test-api-1
  
  # details about api which needs to be tested
  url: http://localhost:8080/
  body: "{\"message\": \"ok\"}"
  headers:
    client-id: gg
    client-secret: wp
  
  # injectors are used to update user defined requests with test-related variables
  injectors:

    # injects the reply-path to a specific path in the requests
    # when running locally it will use Ngrok url
    # when running on a public server it will use localhost or userDefinedHost
    replyPathInjector:
      path: "headers.webhook-reply-to"
    
    # injects the correlationId/traceId to the request
    correlationIdInjector:
      path: "body.uniqueId"

  # pickers are used to figure out where to pick specific info from the response
  pickers:
    # defines where to expect the correlationId/traceId when the downstream gives a callback
    correlationPicker:
      path: "body.uniqueId"

# actual run configuration 
# defines how many requests need to be fired over the span of how many seconds
# in the following example we can expect a request to be fired every 200ms (10s/50r)
run:
  iterations: 50
  durationSeconds: 10

# defines where and in which format the analysis output should go to
# analysis comprises of max, min, avg, etc
outputs:
  - type: text
    path: out.txt
```

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
