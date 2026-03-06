.PHONY: build test clean release mock embed-cfe

embed-cfe:
	mkdir -p cmd/mcp-1c/extension
	cp extension/MCP_HTTPService.cfe cmd/mcp-1c/extension/

build: embed-cfe
	go build -o bin/mcp-1c ./cmd/mcp-1c

test: embed-cfe
	go test ./... -v -race

clean:
	rm -rf bin/ dist/ cmd/mcp-1c/extension

release: embed-cfe
	GOOS=windows GOARCH=amd64 go build -o dist/mcp-1c-windows-amd64.exe ./cmd/mcp-1c
	GOOS=linux GOARCH=amd64 go build -o dist/mcp-1c-linux-amd64 ./cmd/mcp-1c
	GOOS=darwin GOARCH=arm64 go build -o dist/mcp-1c-darwin-arm64 ./cmd/mcp-1c
	GOOS=darwin GOARCH=amd64 go build -o dist/mcp-1c-darwin-amd64 ./cmd/mcp-1c

mock:
	go run ./cmd/mock-1c
