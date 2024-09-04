default:
	@go build -o bin/runtest cmd/runtest/main.go 
run:
	@go run cmd/runtest/main.go -f docs/input-example.yml
run_dummy:
	@go run cmd/dummy/main.go
