SERVICE=emailserv
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS = -X 'main.version=$(VERSION)'
PACKAGE=github.com/mikolajb/${SERVICE}

test:
	cd ${GOPATH}/src/${PACKAGE} && ./scripts/test.sh
	go tool cover -func=coverage.out | tail -n 1

get:
	go get github.com/golang/mock/gomock
	go install github.com/golang/mock/mockgen

gen:
	cd ${GOPATH}/src/${PACKAGE} && ./scripts/generate.sh

build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags "${LDFLAGS}" -a -o bin/${SERVICE} ${PACKAGE}/cmd/${SERVICE}
