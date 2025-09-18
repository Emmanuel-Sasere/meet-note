BINARY_NAME=meetnote
INSTALL_DIR=$(HOME)/bin
VERSION=2.0.0

build: 
	go build -o $(BINARY_NAME)

install: build
	mv $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✅ Installed $(BINARY_NAME) ($(VERSION)) to $(INSTALL_DIR)"


version: 
	@echo "📌 Current project version: $(VERSION)"