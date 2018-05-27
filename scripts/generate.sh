#!/usr/bin/env sh

mockgen -destination internal/emailclient/mock_emailclient.go \
        -package emailclient \
        -source internal/emailclient/common.go EmailClient
