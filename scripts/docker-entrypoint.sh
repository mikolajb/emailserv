#!/usr/bin/env sh

set -e

: ${AMAZON_KEY_ID:=missing}
: ${AMAZIN_SECRET_KEY:=missing}
: ${SENDGRID_KEY:=missing}
: ${LOG_LEVEL:=debug}

emailserv \
    -amazon.key=${AMAZON_KEY_ID} \
    -amazon.secret=${AMAZIN_SECRET_KEY} \
    -sendgrid.key=${SENDGRID_KEY}
