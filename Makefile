NAME        = context
HOSTNAME   ?= registry.terraform.io
NAMESPACE  ?= local
VERSION    ?= 9999.99.99
GOOS       ?= linux
GOARCH     ?= amd64

BINARY       = terraform-provider-${NAME}
PLUGIN_PATH  = ~/.terraform.d/plugins
INSTALL_PATH = $(PLUGIN_PATH)/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(GOOS)_$(GOARCH)

install: build
	mkdir -p $(INSTALL_PATH)
	mv $(BINARY) $(INSTALL_PATH)

build:
	go mod tidy
	go generate
	GOOS=${GOOS} GOARCH=${GOARCH} go build

snapshot:
	goreleaser release --snapshot --rm-dist

test:
	go test ./... -v $(TESTARGS) -timeout 120m
