SHELL := $(shell which sh)
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOPATH := $(shell go env GOPATH)
PROTOC_DIR := bin/protoc-$(GOOS)-$(GOARCH)
PROTOC := bin/protoc
GOLANG_CI_LINT := bin/golangci-lint

AGENTRUNTIME_BIN := bin/agentruntime
AGENTRUNTIME_BIN_FILES := bin/agentruntime-linux-amd64 bin/agentruntime-linux-arm64 bin/agentruntime-darwin-amd64 bin/agentruntime-darwin-arm64 bin/agentruntime-windows-amd64.exe
AGENTNETWORK_BIN := bin/agentnetwork
AGENTNETWORK_BIN_FILES := bin/agentnetwork-linux-amd64 bin/agentnetwork-linux-arm64 bin/agentnetwork-darwin-amd64 bin/agentnetwork-darwin-arm64

.PHONY: build
build: $(AGENTRUNTIME_BIN) $(AGENTNETWORK_BIN)

.PHONY: all
all: $(AGENTRUNTIME_BIN_FILES) $(AGENTNETWORK_BIN_FILES)

bin/protoc-linux-amd64.zip:
	wget -O $@ "https://github.com/protocolbuffers/protobuf/releases/download/v27.1/protoc-27.1-linux-x86_64.zip"

bin/protoc-linux-arm64.zip:
	wget -O $@ "https://github.com/protocolbuffers/protobuf/releases/download/v27.1/protoc-27.1-linux-aarch_64.zip"

bin/protoc-darwin-amd64.zip:
	wget -O $@ "https://github.com/protocolbuffers/protobuf/releases/download/v27.1/protoc-27.1-osx-x86_64.zip"

bin/protoc-darwin-arm64.zip:
	wget -O $@ "https://github.com/protocolbuffers/protobuf/releases/download/v27.1/protoc-27.1-osx-aarch_64.zip"

$(PROTOC_DIR): bin/protoc-$(GOOS)-$(GOARCH).zip
	@unzip -o $< -d $(PROTOC_DIR)
	@echo "done $@"

$(PROTOC): $(PROTOC_DIR)
	chmod 755 $(PROTOC_DIR)/bin/protoc
	ln -sf protoc-$(GOOS)-$(GOARCH)/bin/protoc $(PROTOC)
	touch $(PROTOC)

PROTOC_GEN_GO := $(GOPATH)/bin/protoc-gen-go
$(PROTOC_GEN_GO):
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34

PROTOC_GEN_GO_GRPC := $(GOPATH)/bin/protoc-gen-go-grpc
$(PROTOC_GEN_GO_GRPC):
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4

%.pb.go: %.proto $(PROTOC) $(PROTOC_GEN_GO)
	@export PATH="$(shell go env GOPATH)/bin:$(PATH)"
	$(PROTOC) --go_out=. --go_opt=paths=source_relative -I. $<

%_grpc.pb.go: %.proto $(PROTOC) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC)
	@export PATH="$(shell go env GOPATH)/bin:$(PATH)"
	$(PROTOC) --go-grpc_out=. --go-grpc_opt=paths=source_relative -I. $<

PB_FILES := runtime/runtime.pb.go runtime/runtime_grpc.pb.go thread/thread.pb.go thread/thread_grpc.pb.go network/network.pb.go network/network_grpc.pb.go
.PHONY: pb
pb: $(PB_FILES)

$(GOLANG_CI_LINT):
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.64.7
	@chmod +x $(GOLANG_CI_LINT)
	@echo "golangci-lint installed"

.PHONY: lint
lint: $(GOLANG_CI_LINT) pb
	$(GOLANG_CI_LINT) run

.PHONY: test
test: pb
	ENV_TEST_FILE=$(shell pwd)/.env.test CI=true go test -timeout 15m -p 1 ./...

.PHONY: clean
clean:
	rm -rf bin/*
	rm -rf $(PB_FILES)
	rm -rf $(PROTOC_DIR)
	rm -f $(PROTOC)
	rm -f $(GOLANG_CI_LINT)
	@echo "cleared"

.PHONY: bin/agentruntime-windows-%.exe
bin/agentruntime-windows-%.exe: pb
	GOOS=windows GOARCH=$* CGO_ENABLED=0 go build -o $@ ./cmd/agentruntime

.PHONY: bin/agentruntime-%
bin/agentruntime-%: pb
	$(eval OS_NAME := $(word 1,$(subst -, ,$*)))
	$(eval ARCH_NAME := $(word 2,$(subst -, ,$*)))
	GOOS=$(OS_NAME) GOARCH=$(ARCH_NAME) CGO_ENABLED=0 go build -o $@ ./cmd/agentruntime

.PHONY: bin/agentnetwork-%
bin/agentnetwork-%: pb
	$(eval OS_NAME := $(word 1,$(subst -, ,$*)))
	$(eval ARCH_NAME := $(word 2,$(subst -, ,$*)))
	GOOS=$(OS_NAME) GOARCH=$(ARCH_NAME) CGO_ENABLED=0 go build -o $@ ./cmd/agentnetwork

$(AGENTRUNTIME_BIN): bin/agentruntime-$(GOOS)-$(GOARCH)
	ln -sf agentruntime-$(GOOS)-$(GOARCH) $(AGENTRUNTIME_BIN)

$(AGENTNETWORK_BIN): bin/agentnetwork-$(GOOS)-$(GOARCH)
	ln -sf agentnetwork-$(GOOS)-$(GOARCH) $(AGENTNETWORK_BIN)

.PHONY: install
install:
	CGO_ENABLED=0 go install ./cmd/agentruntime

.PHONY: release
release: $(AGENTRUNTIME_BIN_FILES) $(AGENTNETWORK_BIN_FILES)
	$(eval NEXT_VERSION := $(shell convco version --bump))
	git tag -a v$(NEXT_VERSION) -m "chore(release): v$(NEXT_VERSION)"
	git push origin v$(NEXT_VERSION)
	convco changelog > CHANGELOG.md
	gh release create v$(NEXT_VERSION) $(AGENTRUNTIME_BIN_FILES) --title "v$(NEXT_VERSION)" --notes-file CHANGELOG.md