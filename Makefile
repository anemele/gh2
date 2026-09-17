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
	go test ./pkg/rest/

clean:
	rm -f $(APP_NAME)

