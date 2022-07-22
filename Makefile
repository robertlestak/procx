VERSION=v0.0.1

qjob: bin/qjob_darwin bin/qjob_windows bin/qjob_linux

bin/qjob_darwin:
	mkdir -p bin
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/qjob_darwin cmd/qjob/*.go
	openssl sha512 bin/qjob_darwin > bin/qjob_darwin.sha512

bin/qjob_linux:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/qjob_linux cmd/qjob/*.go
	openssl sha512 bin/qjob_linux > bin/qjob_linux.sha512

bin/qjob_windows:
	mkdir -p bin
	GOOS=windows GOARCH=amd64 go build -ldflags="-X 'main.Version=$(VERSION)'" -o bin/qjob_windows cmd/qjob/*.go
	openssl sha512 bin/qjob_windows > bin/qjob_windows.sha512
