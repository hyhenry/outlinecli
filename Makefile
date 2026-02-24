.PHONY: build install clean

BINARY := outline
BUILD_DIR := bin
CMD := ./cmd/outline

build:
	go build -o $(BUILD_DIR)/$(BINARY) $(CMD)

install: build
	cp $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "Installed to /usr/local/bin/$(BINARY)"

clean:
	rm -rf $(BUILD_DIR)
