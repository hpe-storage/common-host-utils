# (c) Copyright 2019 Hewlett Packard Enterprise Development LP

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GO_VERSION = 1.10
VERSION = $(shell git tag|tail -n1)
ifeq ($(VERSION),)
VERSION = v0.0.0
endif

ifeq ($(DOCKER_TAG),)
DOCKER_TAG = edge
endif

ifndef BUILD_NUMBER
BUILD_NUMBER = 0
endif

# Where our code lives
PKG_PATH = ./
CMD_PATH = ./cmd/
VEN_PATH = vendor
# This is the commit id of the branch we're building from
COMMIT = $(shell git log -n 1 --pretty=format:"%H")
BRANCH = $(shell git rev-parse --abbrev-ref HEAD)

# The version of make for OSX doesn't allow us to export, so
# we add these variables to the env in each invocation.
GOENV = PATH=$$PATH:$(GOPATH)/bin GLIDE_HOME=$(GOPATH)/.glide BRANCH=$(BRANCH)

# Our target binary is for Linux.  To build an exec for your local (non-linux)
# machine, use go build directly.
ifndef GOOS
GOOS = linux
endif
TEST_ENV  = GOOS=$(GOOS) GOARCH=amd64
BUILD_ENV = GOOS=linux GOARCH=amd64 CGO_ENABLED=0 VERSION=$(VERSION) BUILD_NUMBER=$(BUILD_NUMBER) COMMIT=$(COMMIT)

# gometalinter allows us to have a single target that runs multiple linters in
# the same fashion.  This variable controls which linters are used.
LINTER_FLAGS = --vendor --disable-all --enable=vet --enable=vetshadow --enable=golint --enable=ineffassign --enable=goconst --enable=deadcode --enable=dupl --enable=varcheck --enable=gocyclo --enable=misspell -i "simplivity"

# list of packages
PACKAGE_LIST =   $(shell export $(GOENV) && go list ./$(PKG_PATH)...| grep -v vendor)

ifeq ($(GITORG),)
GITORG = $(USER)
endif

# prefixes to make things pretty
A1 = $(shell printf "»")
A2 = $(shell printf "»»")
A3 = $(shell printf "»»»")
S0 = 😁
S1 = 😔

.PHONY: help
help:
	@echo "Targets:"
	@echo "    tools                        - Download and install go tooling required to build."
	@echo "    vendor                       - Download dependencies (glide install)"
	@echo "    vendup                       - Download dependencies. (glide up)"
	@echo "    lint                         - Static analysis of source code.  Note that this must pass in order to build."
	@echo "    test                         - Run unit tests."
	@echo "    int                          - Run integration tests.  (Not implemented yet)."
	@echo "    clean                        - Remove build artifacts."
	@echo "    debug                        - Display make's view of the world."
	@echo "    all                          - Build all cmds."
	@echo "    all_local                    - Build all cmds for local OS (make sure GOOS is set)."
	@echo "    utils                        - Build everything in ./cmd/"
	@echo "    nimbletunepkg                - Package nimbletune in a tar"

.PHONY: debug
debug:
	@echo "Debug:"
	@echo "  Go:           `go version`"
	@echo "  GOPATH:       $(GOPATH)"
	@echo "  GOOS:         $(GOOS)"
	@echo "  Packages:     $(PACKAGE_LIST)"
	@echo "  VERSION:      $(VERSION)"
	@echo "  BRANCH:       $(BRANCH)"
	@echo "  COMMIT:       $(COMMIT)"
	@echo "  BUILD_NUMBER: $(BUILD_NUMBER)"
	@echo "  DOCKER_TAG:   $(DOCKER_TAG)"
	@echo "  BUILD_ENV:    $(BUILD_ENV)"
	@echo "  COMMOM_GIT:   $(COMMOM_GIT)"
	@echo "  GOENV:        $(GOENV)"
	@echo "$(S0)"

.PHONY : all
all : clean lint docker_run

.PHONY : all_local
all_local : clean debug lint utils nimbletunepkg test

# this is the target called from within the container
.PHONY: container_all
container_all: debug utils nimbletunepkg test

.PHONY: tools
tools: ; $(info $(A1) gettools)
ifeq ($(NESTED_MAKE),)
	@echo "$(A2) get gometalinter"
	go get -u github.com/alecthomas/gometalinter
	@echo "$(A2) install gometalinter"
	export $(GOENV) && gometalinter --install
	@echo "$(A2) get glide"
	go get -u github.com/Masterminds/glide
	go install github.com/Masterminds/glide
	go get -u github.com/josephspurrier/goversioninfo
	go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo
	go install ./cmd/updatewinversioninfo/
endif
	@echo "$(S0)"

vendup: tools; $(info $(A1) vendup)
	@echo "$(A2) glide up"
	export $(GOENV) && glide up
	@echo "$(S0)"

vendor: tools; $(info $(A1) vendor)
ifeq ($(NESTED_MAKE),)
	# ignore glide install if vendor is already present. As this will wipe out common if already present under vendor/ and causes clone everytime
	if [ ! -d vendor ]; then echo "$(A2) glide install"; export $(GOENV) && glide cc && glide install; fi
endif
	@echo "$(S0)"

build: ; $(info $(A1) mkdir build)
	@mkdir build
	@echo "$(S0)"

.PHONY: lint
lint: vendor; $(info $(A1) lint)
	@echo "$(A2) lint $(PKG_PATH)"
	export $(GOENV) $(BUILD_ENV) && gometalinter $(LINTER_FLAGS) $(PKG_PATH)... --exclude $(VEN_PATH)...
	@echo "$(S0)"

.PHONY: clean
clean: ; $(info $(A1) clean)
	@echo "$(A2) remove build"
	@rm -rf build
	@echo "$(A2) remove glide.lock to fetch latest deps(common etc)"
	@rm -rf glide.lock > /dev/null 2>&1
	@echo "$(A2) remove src"
	@rm -rf src
	@echo "$(A2) remove bin"
	@rm -rf bin
	@echo "$(A2) remove pkg"
	@rm -rf pkg
	@echo "$(A2) remove resource.syso"
	@find cmd -name "resource.syso" -type f -delete
	@echo "$(A2) remove vendor"
	@rm -rf vendor
	@echo "$(S0)"

.PHONY: test
test: utils; $(info $(A1) test)
	@echo "$(A2) unit tests"
ifeq ("$(GOOS)","linux")
	export $(GOENV) $(TEST_ENV) && ./vendor/github.com/hpe-storage/common-host-libs/package_tester.sh $(PACKAGE_LIST)
else
	@echo "Skipping tests... only linux is supported!"
endif
	@echo "$(S0)"

.PHONY: int
int: ; $(info $(A1) int)
	@echo "$(A2) There are no integration tests yet."
	@echo "$(S1)"

.PHONY: utils
utils: build vendor; $(info $(A1) utils)
	@echo "$(A2) build utilities"
	export $(GOENV) $(BUILD_ENV) && ./vendor/github.com/hpe-storage/common-host-libs/package_builder.sh $(PACKAGE_LIST)
	@echo "$(S0)"

.PHONY: nimbletunepkg
 nimbletunepkg: utils; $(info $(A1) nimbletune)
	# copying files temporarily
	cp ./vendor/github.com/hpe-storage/common-host-libs/tunelinux/config/* .
	cp build/nimbletune .
	# create package
	tar -cvzf build/nimbletune.tar.gz config.json 99-nimble-tune.rules multipath.* nimbletune
	# remove temporary files
	rm -f nimbletune config.json cloud_vm_config.json 99-nimble-tune.rules multipath.*

.PHONY: docker_run
docker_run: ; $(info $(A1) docker_run)
	@echo "$(A2) using docker image for build"
	docker run --rm -t --env BUILD_NUMBER=$(BUILD_NUMBER) --user `id -u $(USER)` -v $(GOPATH):/go -w /go golang:$(GO_VERSION) sh -c "cd src/github.com/hpe-storage/common-host-utils && export NESTED_MAKE=true && export XDG_CACHE_HOME=/tmp/.cache && make container_all"
	@echo "$(A2) leaving container happy $(S0)"
