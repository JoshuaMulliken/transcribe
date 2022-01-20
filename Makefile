bindir=/usr/local/bin
builddir=build

.PHONY: build build_all build_linux build_darwin build_windows install uninstall clean
build:
	mkdir -p $(builddir)
	go build -o $(builddir)/transcribe

build_all: build_linux build_darwin build_windows

build_linux:
	mkdir -p $(builddir)
	GOOS=linux GOARCH=amd64 go build -o $(builddir)/transcribe-linux-amd64
	GOOS=linux GOARCH=386 go build -o $(builddir)/transcribe-linux-386
	GOOS=linux GOARCH=arm64 go build -o $(builddir)/transcribe-linux-arm64
	GOOS=linux GOARCH=arm go build -o $(builddir)/transcribe-linux-arm

build_darwin:
	mkdir -p $(builddir)
	GOOS=darwin GOARCH=amd64 go build -o $(builddir)/transcribe-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o $(builddir)/transcribe-darwin-arm64

build_windows:
	mkdir -p $(builddir)
	GOOS=windows GOARCH=amd64 go build -o $(builddir)/transcribe-windows-amd64.exe
	GOOS=windows GOARCH=386 go build -o $(builddir)/transcribe-windows-386.exe
	GOOS=windows GOARCH=arm64 go build -o $(builddir)/transcribe-windows-arm64.exe

install:
	go build -o $(DESTDIR)$(bindir)/transcribe

uninstall:
	rm -f $(DESTDIR)$(bindir)/transcribe

clean:
	rm -rf $(builddir)