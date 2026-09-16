APP_NAME=gh2

build: fmt test
	go build \
		-trimpath \
		-ldflags='-s -w' \
		-o $(APP_NAME) \
		./cmd/

fmt:
	go fmt

test:
	go test ./core/

.PHONY: clean

clean:
	rm -f $(APP_NAME)

