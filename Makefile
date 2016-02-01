VERSION=0.0.1

all: jmx-get

.PHONY: jmx-get

gom:
	go get -u github.com/mattn/gom

bundle:
	gom install

jmx-get: jmx-get.go
	gom build -o jmx-get

linux: jmx-get.go
	GOOS=linux GOARCH=amd64 gom build -o jmx-get

fmt:
	go fmt ./...

dist:
	git archive --format tgz HEAD -o jmx-get-$(VERSION).tar.gz --prefix jmx-get-$(VERSION)/

clean:
	rm -rf jmx-get jmx-get-*.tar.gz

