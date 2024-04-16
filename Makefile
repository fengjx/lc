export PATH := $(PATH):`go env GOPATH`/bin
export GO111MODULE=on
LDFLAGS := -s -w

.PHONY: all
all: build

.PHONY: build
build:
	go build -gcflags='all=-N -l'

.PHONY: tidy
install:
	go tidy

.PHONY: format
format:
	gofumpt -l -w .

.PHONY: init-git
init-git:
	@cp .github/hooks/* .git/hooks/
	@chmod +x .git/hooks/*
	@echo 'git init done'
