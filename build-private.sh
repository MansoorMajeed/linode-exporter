#!/bin/bash

docker build -t mansoor1/linode-exporter:$(git rev-parse HEAD) -t mansoor1/linode-exporter:latest .
docker push mansoor1/linode-exporter:$(git rev-parse HEAD)
docker push mansoor1/linode-exporter:latest
