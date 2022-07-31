VERSION=v0.0.17

procx: bin/procx_darwin bin/procx_windows bin/procx_linux

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
