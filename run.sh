#! /bin/sh

./nsq_http_sub \
  -http-address="0.0.0.0:9090" \
  -lookupd-http-address="127.0.0.1:4161" \
  -max-in-flight=1 \
  -timeout=1 \
  -max-messages=5
