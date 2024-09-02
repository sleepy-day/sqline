MAIN_PATH := .
BINARY := sqline
WINDOWS_BINARY := sqline.exe
BUILD_FLAGS := -ldflags='-w -s' -trimpath
WIN_FLAGS := CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64
LINUX_FLAGS := CGO_ENABLED=1 GOOS=linux GOARCH=amd64 

.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

.PHONY: release/linux
release/linux: tidy
	${LINUX_FLAGS} go build ${BUILD_FLAGS} -o=release/linux/${BINARY} ${MAIN_PATH}
	
.PHONY: release/windows
release/windows: tidy
	${WIN_FLAGS} go build ${BUILD_FLAGS} -o=release/windows/${WINDOWS_BINARY} ${MAIN_PATH}
