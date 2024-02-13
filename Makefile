ifeq ($(OS),Windows_NT)
    SHELL=CMD.EXE
    SET=set
    DEL=del
    NUL=nul
    WHICH=where.exe
else
    SET=export
    DEL=rm
    NUL=/dev/null
    WHICH=which
endif

ifndef GO
    SUPPORTGO=go1.20.14
    GO:=$(shell $(WHICH) $(SUPPORTGO) 2>$(NUL) || echo go)
endif

NAME:=$(notdir $(CURDIR))
VERSION:=$(shell git describe --tags 2>$(NUL) || echo v0.0.0)
GOOPT:=-ldflags "-s -w -X main.version=$(VERSION)"
EXE:=$(shell go env GOEXE)

all:
	$(GO) fmt ./...
	$(SET) "CGO_ENABLED=0" && $(GO) build $(GOOPT)

test:
	$(GO) test -v
	.\$(NAME) test.lua

_dist:
	$(SET) "CGO_ENABLED=0" && $(GO) build $(GOOPT)
	zip -9 $(NAME)-$(VERSION)-$(GOOS)-$(GOARCH).zip $(NAME)$(EXE)

dist:
	$(SET) "GOOS=windows" && $(SET) "GOARCH=386"   && $(MAKE) _dist
	$(SET) "GOOS=windows" && $(SET) "GOARCH=amd64" && $(MAKE) _dist

clean:
	$(DEL) *.zip $(NAME)$(EXE)

manifest:
	make-scoop-manifest *-windows-*.zip > expect-lua.json

release:
	gh release create -d --notes "" -t $(VERSION) $(VERSION) $(wildcard $(NAME)-$(VERSION)-*.zip)

$(SUPPORTGO):
	go install golang.org/dl/$(SUPPORTGO)@latest
	$(SUPPORTGO) download

.PHONY: all test dist _dist clean manifest release
