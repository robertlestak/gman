VERSION=v0.0.1

.PHONY: build clean docker docker-latest docker-push install
.PHONY: web web-install clean-web
build: clean-web bin/gman_darwin_amd64 bin/gman_darwin_arm64 bin/gman_linux_amd64 bin/gman_linux_arm64 bin/gman_windows_amd64.exe

bin/gman_darwin_amd64: clean-web
	GOOS=darwin GOARCH=amd64 go build  -ldflags="-X 'main.Version=$(VERSION)'" -o bin/gman_darwin_amd64 cmd/gman/*.go
	openssl sha512 bin/gman_darwin_amd64 > bin/gman_darwin_amd64.sha512
	
bin/gman_darwin_arm64: clean-web
	GOOS=darwin GOARCH=arm64 go build  -ldflags="-X 'main.Version=$(VERSION)'" -o bin/gman_darwin_arm64 cmd/gman/*.go
	openssl sha512 bin/gman_darwin_arm64 > bin/gman_darwin_arm64.sha512

bin/gman_linux_amd64: clean-web
	GOOS=linux GOARCH=amd64 go build  -ldflags="-X 'main.Version=$(VERSION)'" -o bin/gman_linux_amd64 cmd/gman/*.go
	openssl sha512 bin/gman_linux_amd64 > bin/gman_linux_amd64.sha512

bin/gman_linux_arm64: clean-web
	GOOS=linux GOARCH=arm64 go build  -ldflags="-X 'main.Version=$(VERSION)'" -o bin/gman_linux_arm64 cmd/gman/*.go
	openssl sha512 bin/gman_linux_arm64 > bin/gman_linux_arm64.sha512

bin/gman_windows_amd64.exe: clean-web
	GOOS=windows GOARCH=amd64 go build  -ldflags="-X 'main.Version=$(VERSION)'" -o bin/gman_windows_amd64.exe cmd/gman/*.go
	openssl sha512 bin/gman_windows_amd64.exe > bin/gman_windows_amd64.exe.sha512

bin/gman: clean-web
	go build  -ldflags="-X 'main.Version=$(VERSION)'" -o bin/gman cmd/gman/*.go
	openssl sha512 bin/gman > bin/gman.sha512

install: bin/gman
	sudo cp bin/gman /usr/local/bin/gman

clean-web:
	rm -rf internal/web/web/.docusaurus
	rm -rf internal/web/web/build
	rm -rf internal/web/web/node_modules

web-install:
	cd internal/web/web && npm install

web: web-install
	cd internal/web/web && npm run build

docker:
	docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 -t robertlestak/gman:$(VERSION) .

docker-push:
	docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 -t robertlestak/gman:$(VERSION) --push .

docker-latest: docker-push
	docker buildx imagetools create robertlestak/gman:$(VERSION) --tag robertlestak/gman:latest

clean: clean-web
	rm -rf bin/*