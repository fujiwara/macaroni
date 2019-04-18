export GO111MODULE := on
LATEST_TAG := $(shell git describe --abbrev=0 --tags)

.PHONY: setup setup_ci test lint dist clean release

test: setup
	go test -v ./...

cmd/macaroni/macaroni: *.go cmd/macaroni/main.go
	cd cmd/macaroni && \
	go build -ldflags "-w -s"

install: cmd/macaroni/macaroni
	install cmd/macaroni/macaroni $(GOPATH)/bin

setup_ci:
	GO111MODULE=off go get \
		github.com/Songmu/goxz \
		github.com/tcnksm/ghr \
		golang.org/x/lint/golint
	go get -d -t ./...

lint: setup
	go vet ./...
	golint -set_exit_status ./...

dist: setup
	goxz -pv=$(LATEST_TAG) -os=darwin,linux -build-ldflags="-w -s" -arch=amd64 -d=dist -z ./cmd/macaroni

clean:
	rm -fr dist/* cmd/macaroni/macaroni

release: dist
	ghr -u fujiwara -r macaroni $(LATEST_TAG) dist/

