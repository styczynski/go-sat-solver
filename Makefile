COMMIT_ID=$(shell git rev-parse --short HEAD)
VERSION=$(shell cat VERSION)

NAME=sat-solver

all: clean build

clean:
	@echo ">> cleaning..."
	@rm -f $(NAME)

build: clean
	@echo ">> building..."
	@echo "Commit: $(COMMIT_ID)"
	@echo "Version: $(VERSION)"
	@go build -o $(NAME) -ldflags "-X main.Version=$(VERSION) -X main.CommitId=$(COMMIT_ID)" ./cmd/go-sat-solver/...

install:
	@go install -ldflags "-X main.Version=$(VERSION) -X main.CommitId=$(COMMIT_ID)" ./cmd/go-sat-solver/...

wasm:
	GOOS=js GOARCH=wasm go build -o ./web/go-sat-solver-web/public/sat.wasm cmd/wasm/wasm.go

start-web:
	cd web/go-sat-solver-web && yarn start

build-web:
	cd web/go-sat-solver-web && yarn build && sleep 2 && mv ./web/go-sat-solver-web/build ./docs

.PHONY: all clean build install
