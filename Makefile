SHELL := /bin/bash

.PHONY: confv
confv:
	go install -v ./cmd/confv

.PHONY: build
build: confv

.PHONY: image.kubectl
image.kubectl:
	cd docker/kubectl && docker build -t dunjut/kubectl:v1.10.0 . && cd -

.PHONY: image.confv-install
image.confv-install:
	GOOS=linux GOARCH=amd64 go build -o ./docker/confv-install/confv ./cmd/confv
	cd ./docker/confv-install && docker build -t dunjut/confv-install:latest . && cd -
	rm ./docker/confv-install/confv

.PHONY: image
image: image.kubectl image.confv-install
