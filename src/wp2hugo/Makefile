BINARY_NAME := wp2hugo

build_debug:
	go build -v -o bin/${BINARY_NAME} ./cmd/wp2hugo

build_prod:
	# Shrink binary by removing symbol and DWARF table
	# Ref: https://lukeeckley.com/post/useful-go-build-flags/
	go build -v -ldflags="-s -w" -o bin/${BINARY_NAME} ./cmd/wp2hugo

go_lint:
	go mod tidy
	go vet ./...
	# Deprecated
	# golint -set_exit_status ./...
	go tool fix ./..
	golangci-lint run

go_format:
	go fmt ./...
	go mod tidy
	go tool fix .
	golangci-lint run --fix

format: go_format

lint: go_lint

clean:
	go clean --modcache
	rm -rf bin/*

test:
	go test ./... -v

update_go_deps:
	go get -t -u ./...
