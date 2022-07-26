VERSION=v0.0.65

.PHONY: procx
procx: clean bin/procx_darwin bin/procx_windows bin/procx_linux

bin/procx_darwin:
	mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/procx_darwin cmd/procx/*.go
	openssl sha512 bin/procx_darwin > bin/procx_darwin.sha512

bin/procx_linux:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/procx_linux cmd/procx/*.go
	openssl sha512 bin/procx_linux > bin/procx_linux.sha512

bin/procx_hostarch:
	mkdir -p bin
	go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/procx_hostarch cmd/procx/*.go
	openssl sha512 bin/procx_hostarch > bin/procx_hostarch.sha512

bin/procx_windows:
	mkdir -p bin
	GOOS=windows GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/procx_windows cmd/procx/*.go
	openssl sha512 bin/procx_windows > bin/procx_windows.sha512

.PHONY: envvars
envvars:
	egrep -oh --exclude Makefile \
		--exclude-dir bin \
		--exclude-dir scripts \
		-R 'os.Getenv\(.*?\)' . | \
		tr -d ' ' | \
		sort | \
		uniq | \
		sed -e 's,os.Getenv(,,g' -e 's,),,g' \
		-e 's,",,g' \
		-e 's,prefix+,PROCX_,g'

.PHONY: envvarsyaml
envvarsyaml:
	bash scripts/envvarsyaml.sh

.PHONY: clean
clean:
	rm -rf bin

.PHONY: slim
slim:
	bash scripts/build_drivers.sh build $(drivers)

.PHONY: listdrivers
listdrivers:
	bash scripts/build_drivers.sh list
