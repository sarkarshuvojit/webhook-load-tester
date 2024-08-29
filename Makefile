default:
	@go build -o bin/default cmd/default/main.go 
run:
	@go run cmd/default/main.go
run_dummy:
	@go run cmd/dummy/main.go
