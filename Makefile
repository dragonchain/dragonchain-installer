LDFLAGS := -X "github.com/dragonchain/dragonchain-installer/internal/configuration.Version=$(shell cat .version)-$(shell git rev-parse --short HEAD)"
BINARY := dc-installer
UNIXPLATFORMS := linux darwin
os = $(word 1, $@)

.PHONY: $(UNIXPLATFORMS)
$(UNIXPLATFORMS):
	mkdir -p release
	GOOS=$(os) GOARCH=amd64 go build -v -ldflags '-s -w $(LDFLAGS)' -o release/$(BINARY)-$(os)-amd64 github.com/dragonchain/dragonchain-installer/cmd/dc-installer

.PHONY: windows
windows:
	mkdir -p release
	cp windows/windows_amd64.syso cmd/dc-installer/windows_amd64.syso
	GOOS=$(os) GOARCH=amd64 go build -v -ldflags '-s -w $(LDFLAGS)' -o release/$(BINARY)-$(os)-amd64.exe github.com/dragonchain/dragonchain-installer/cmd/dc-installer
	rm cmd/dc-installer/windows_amd64.syso

.PHONY: linux-arm64
linux-arm64:
	mkdir -p release
	GOOS=linux GOARCH=arm64 go build -v -ldflags '-s -w $(LDFLAGS)' -o release/$(BINARY)-linux-arm64 github.com/dragonchain/dragonchain-installer/cmd/dc-installer

.PHONY: release
release: linux linux-arm64 darwin windows

clean:
	rm -rf release
