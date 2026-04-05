BINARY := spacetrader
BINARY_ARM := spacetrader-arm

.PHONY: build build-pi test clean run

build:
	go build -o $(BINARY) .

build-pi:
	GOOS=linux GOARCH=arm GOARM=6 go build -o $(BINARY_ARM) -ldflags="-s -w" .

run: build
	./$(BINARY)

test:
	go test ./internal/...

test-cover:
	go test -cover ./internal/...

clean:
	rm -f $(BINARY) $(BINARY_ARM)
