# webhook-load-tester

![Overview](docs/overview.png "Overview of design")


## Setting up locally

### Start Dummy Webhook API 

Start a simple webhook api

```bash
$ make start_dummy
```

### Start tests  

Start the actual tester

#### Using Make

```bash
$ make run
```

#### Using Go

```bash
$ go run cmd/default/main.go
$ go run -version cmd/default/main.go # for verbose output
```

## TODO

- Timeout for watiting to prevent deadlock when downstream fails to reply to few requests
