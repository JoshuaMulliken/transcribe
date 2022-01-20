bindir=/usr/local/bin
builddir=build
tmpdir=/tmp

.PHONY: build build_all build_linux build_darwin build_windows install uninstall clean archive
build:
	go build -o $(builddir)/transcribe

build_all: build_linux build_darwin build_windows

archive:
	tar -czf $(tmpdir)/transcribe$(VERSION_STR).tar.gz .

build_linux:
	GOOS=linux GOARCH=amd64 go build -o $(builddir)/transcribe$(VERSION_STR)-linux-amd64
	GOOS=linux GOARCH=386 go build -o $(builddir)/transcribe$(VERSION_STR)-linux-386
	GOOS=linux GOARCH=arm64 go build -o $(builddir)/transcribe$(VERSION_STR)-linux-arm64
	GOOS=linux GOARCH=arm go build -o $(builddir)/transcribe$(VERSION_STR)-linux-arm

build_darwin:
	GOOS=darwin GOARCH=amd64 go build -o $(builddir)/transcribe$(VERSION_STR)-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o $(builddir)/transcribe$(VERSION_STR)-darwin-arm64

build_windows:
	GOOS=windows GOARCH=amd64 go build -o $(builddir)/transcribe$(VERSION_STR)-windows-amd64.exe
	GOOS=windows GOARCH=386 go build -o $(builddir)/transcribe$(VERSION_STR)-windows-386.exe
	GOOS=windows GOARCH=arm64 go build -o $(builddir)/transcribe$(VERSION_STR)-windows-arm64.exe

install:
	go build -o $(DESTDIR)$(bindir)/transcribe

uninstall:
	rm -f $(DESTDIR)$(bindir)/transcribe

clean:
	rm -rf $(builddir)