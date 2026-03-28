export PATH := $(PATH):`go env GOPATH`/bin
export GO111MODULE=on
LDFLAGS := -s -w
NOWEB_TAG = $(shell [ ! -d web/frps/dist ] || [ ! -d web/frpc/dist ] && echo ',noweb')
NOWEBUI_TAG = $(shell [ ! -d webui/frpc/dist ] || [ ! -d webui/frps/dist ] && echo ',nowebui')

.PHONY: web webui frps-web frpc-web frps-webui frpc-webui frps frpc

all: env fmt web build

build: frps frpc

env:
	@go version

web: frps-web frpc-web frps-webui frpc-webui

webui: frps-webui frpc-webui

frps-web:
	$(MAKE) -C web/frps build

frpc-web:
	$(MAKE) -C web/frpc build

frpc-webui:
	$(MAKE) -C webui/frpc build

frps-webui:
	$(MAKE) -C webui/frps build

fmt:
	go fmt ./...

fmt-more:
	gofumpt -l -w .

gci:
	gci write -s standard -s default -s "prefix(github.com/fatedier/frp/)" ./

vet:
	go vet -tags "$(NOWEB_TAG)$(NOWEBUI_TAG)" ./...

frps:
	mkdir -p release
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -tags "frps$(NOWEB_TAG)$(NOWEBUI_TAG)" -o release/frps ./cmd/frps

frpc:
	mkdir -p release
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -tags "frpc$(NOWEB_TAG)$(NOWEBUI_TAG)" -o release/frpc ./cmd/frpc

test: gotest

gotest:
	go test -tags "$(NOWEB_TAG)$(NOWEBUI_TAG)" -v --cover ./assets/...
	go test -tags "$(NOWEB_TAG)$(NOWEBUI_TAG)" -v --cover ./cmd/...
	go test -tags "$(NOWEB_TAG)$(NOWEBUI_TAG)" -v --cover ./client/...
	go test -tags "$(NOWEB_TAG)$(NOWEBUI_TAG)" -v --cover ./server/...
	go test -tags "$(NOWEB_TAG)$(NOWEBUI_TAG)" -v --cover ./pkg/...

e2e:
	./hack/run-e2e.sh

e2e-trace:
	DEBUG=true LOG_LEVEL=trace ./hack/run-e2e.sh

e2e-compatibility-last-frpc:
	if [ ! -d "./lastversion" ]; then \
		TARGET_DIRNAME=lastversion ./hack/download.sh; \
	fi
	FRPC_PATH="`pwd`/lastversion/frpc" ./hack/run-e2e.sh
	rm -r ./lastversion

e2e-compatibility-last-frps:
	if [ ! -d "./lastversion" ]; then \
		TARGET_DIRNAME=lastversion ./hack/download.sh; \
	fi
	FRPS_PATH="`pwd`/lastversion/frps" ./hack/run-e2e.sh
	rm -r ./lastversion

alltest: vet gotest e2e
	
clean:
	rm -f ./release/frpc
	rm -f ./release/frpc.exe
	rm -f ./release/frps
	rm -f ./release/frps.exe
	rm -rf ./lastversion
