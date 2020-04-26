CC=go
BUILD=$(shell git describe)
VERSION=$(shell cat VERSION)
OS=$(shell uname -s)
ARCH=$(shell uname -m)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
APP=utg-bot
SOURCES=discord.go handlers.go main.go missions.go mods.go server.go watchdog.go

build: $(APP)

$(APP): $(SOURCES)
	go build -o $@ -v $^

install: build
	@cp $(APP) /usr/bin/$(APP)

clean:
	-rm -f $(APP)

mrproper: clean

test:
	go test .
