BINARY = gcp-gpu-metrics

LAST_COMMIT = `git rev-parse HEAD`

VERSION_PKG = main

MAIN_LOCATION = ./

VERSION := $(shell git describe --exact-match --abbrev=0 --tags $(git rev-list --tags --max-count=1) 2> /dev/null)
ifndef VERSION
	VERSION = $(shell git describe --abbrev=3 --tags $(git rev-list --tags --max-count=1))-dev
endif

.PHONY: all
all: ${BINARY}

${BINARY}:
		@echo " > Build ${BINARY} binary < "
		GO111MODULE=on go build -mod=vendor -ldflags "-X ${VERSION_PKG}.Commit=${LAST_COMMIT} -X ${VERSION_PKG}.Version=${VERSION}" \
			-o ${BINARY} ${MAIN_LOCATION}

.PHONY: lint
lint:
		@echo " > Lint Go code < "
		golint -set_exit_status ./*.go

.PHONY: clean
clean:
		@echo " > Delete ${BINARY} binary < "
		rm -f ${BINARY}

.PHONY: re
re: clean all

define GORELEASER_COMMAND
VERSION_PKG=${VERSION_PKG} VERSION=${VERSION} \
	LAST_COMMIT=${LAST_COMMIT} goreleaser
endef

.PHONY: goreleaser
goreleaser:
ifneq (,$(findstring -dev,$(VERSION)))
	@echo Run Goreleaser in snapshot mod
	${GORELEASER_COMMAND} --snapshot --rm-dist
else
	@echo Run Goreleaser in release mod
	${GORELEASER_COMMAND} release --rm-dist
endif


