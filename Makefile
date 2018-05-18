SERVICE=emailserv
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS = -X 'main.version=$(VERSION)'
PACKAGE=github.com/mikolajb/${SERVICE}

build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags "${LDFLAGS}" -a -o bin/${SERVICE} ${PACKAGE}/cmd/${SERVICE}
