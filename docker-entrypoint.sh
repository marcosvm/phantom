#!/bin/sh

mkdir -p /pang/logs/phantom

exec /phantom \
        -header ${ORIGIN_HEADER:-X-Forwarded-For} \
        -listen ${LISTEN_ADDRESS:-:7777} \
        -log.level ${LOG_LEVEL:-info} >> /pang/logs/phantom/phantom.log 2>&1
