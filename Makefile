.PHONY: help build test release clean install uninstall release-local release-check

help:
	@echo "Network Probe - Makefile commands:"
	@echo ""
	@echo "Development:"
	@echo "  make build          - Build the project in debug mode"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Release:"
	@echo "  make release        - Build release binary"
	@echo "  make release-local  - Run goreleaser locally (skip publishing)"
	@echo "  make release-check  - Check goreleaser configuration"
	@echo ""
	@echo "Installation:"
	@echo "  make install        - Install the binary (requires sudo)"
	@echo "  make uninstall      - Uninstall the binary (requires sudo)"

build:
	cargo build

test:
	cargo test --all

release:
	cargo build --release

clean:
	cargo clean

install:
	@echo "Installing Network Probe..."
	@cargo install --path .
	@echo "Network Probe installed successfully!"

uninstall:
	@echo "Uninstalling Network Probe..."
	@cargo uninstall network-probe
	@echo "Network Probe uninstalled successfully!"

release-local:
	@echo "Building release packages locally..."
	@goreleaser release --snapshot --clean

release-check:
	@echo "Checking goreleaser configuration..."
	@goreleaser check

release-snapshot:
	@echo "Building snapshot release..."
	@goreleaser release --snapshot --clean --skip-publish

release-publish:
	@echo "Building and publishing release..."
	@goreleaser release --clean

docker-build:
	@echo "Building Docker image..."
	@docker build -t network-probe:latest .

docker-run:
	@echo "Running Docker container..."
	@docker run -it --rm network-probe:latest

lint:
	@echo "Running linters..."
	@cargo clippy --all-targets --all-features

fmt:
	@echo "Formatting code..."
	@cargo fmt --all

check: lint test
