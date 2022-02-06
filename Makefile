
VERSION=$(shell git describe --abbrev=0 --tags)

BUILD=$(shell git rev-parse --short HEAD)
DATE=$(shell date +%FT%T%z)

# Binaries to be build
PLATFORMS = linux/glauth-ui windows/glauth-ui.exe darwin/glauth-ui-app
BINS = $(wildcard build/*/*)

# functions
temp = $(subst /, ,$@)
os = $(word 1, $(temp))

# Setup the -ldflags option for go building, interpolate the variable values
LDFLAGS=-trimpath -ldflags "-w -s -X 'glauth-ui-light/handlers.Version=${VERSION}, git: ${BUILD}, build: ${DATE}'"

# Build binaries
#  first build : linux/glauth-ui
$(PLATFORMS):
	@mkdir -p build/${os}
	CGO_ENABLED=0 GOOS=${os} go build ${LDFLAGS} -o build/$@
	@echo " => bin builded: build/$@"

build: $(PLATFORMS)

# List binaries
$(BINS):
	@sha256sum $@ 

sha: $(BINS)

# Cleans our project: deletes binaries
clean:
	rm -rf build/
	@echo "Build cleaned"

tr:
	@echo "shell is $$0"
	rgrep -hoP '{{ tr "(.*?)" }}' routes/web/templates/ | sed  "s/{{ tr \"//" | sed "s/\" }}/: \"\"/" | sort | uniq  > tr.tmp
	rgrep -hoP 'Tr\(lang, "(.*?)"' handlers/* | sed "s/Tr(lang, \"//" | sort | uniq | awk -F'"' '{print $$1": \"\""}' >> tr.tmp
	cat tr.tmp | sort | uniq > locales/tr.yml
	rm tr.tmp



all: build

.PHONY: clean build sha tr $(BINS) $(PLATFORMS)

