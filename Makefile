.PHONY: build-linux run-linux test-linux verify-linux build-wasm build-windows

BIN_DIR := bin
APP_NAME := IFC-Raw-Viewer_mk2

build-linux:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/stepview

run-linux: build-linux
	./$(BIN_DIR)/$(APP_NAME)

test-linux:
	go vet ./...
	go test -race -bench=. -benchmem ./...

build-wasm:
	mkdir -p $(BIN_DIR)
	GOOS=js GOARCH=wasm go build -o $(BIN_DIR)/$(APP_NAME).wasm ./cmd/stepview

build-windows:
	"$(shell go env GOPATH)/bin/fyne-cross" windows -engine podman -arch amd64 -name $(APP_NAME).exe -app-id de.wellermann.ifcrawviewer.mk2 -dir "$(CURDIR)" ./cmd/stepview
	mkdir -p $(BIN_DIR)/windows-amd64
	unzip -o fyne-cross/dist/windows-amd64/$(APP_NAME).exe.zip -d $(BIN_DIR)/windows-amd64
	cp fyne-cross/dist/windows-amd64/$(APP_NAME).exe.zip $(BIN_DIR)/$(APP_NAME)-windows-amd64.zip

verify-linux: test-linux build-linux build-wasm
