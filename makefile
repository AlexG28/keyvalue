BINARY_NAME=kvstore

SRC_DIR=.

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(SRC_DIR)
	@echo "$(BINARY_NAME) built successfully!"

clean:
	@echo "Cleaning up..."
	@if [ -f "$(BINARY_NAME)" ]; then \
		rm $(BINARY_NAME); \
		echo "$(BINARY_NAME) removed."; \
	else \
		echo "$(BINARY_NAME) not found, nothing to clean."; \
	fi

.PHONY: all build clean