build:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.Version=`git describe --tags`" -o bin/rait cmd/rait/rait.go
clean:
	rm -r bin/
.PHONY: build
