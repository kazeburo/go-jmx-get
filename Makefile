VERSION=0.0.2
LDFLAGS=-ldflags "-X main.Version=${VERSION}"
GO111MODULE=on

all: jmx-get

.PHONY: jmx-get

jmx-get: jmx-get.go
	go build $(LDFLAGS) -o jmx-get

linux: jmx-get.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o jmx-get

clean:
	rm -rf jmx-get

tag:
	git tag v${VERSION}
	git push origin v${VERSION}
	git push origin master
	goreleaser --rm-dist
