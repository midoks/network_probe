.PHONY: help build test release clean install uninstall release-local release-check build-linux build-macos build-windows

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
	@echo "  make build-linux    - Build for Linux (x86_64)"
	@echo "  make build-macos    - Build for macOS (x86_64 and arm64)"
	@echo "  make build-windows  - Build for Windows (x86_64)"
	@echo ""
	@echo "Packaging:"
	@echo "  make package-deb    - Build Debian package"
	@echo "  make package-rpm    - Build RPM package"
	@echo "  make package-all    - Build all packages"
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

build-linux:
	@echo "Building for Linux x86_64..."
	cargo build --release --target x86_64-unknown-linux-gnu
	@echo "Linux binary built at target/x86_64-unknown-linux-gnu/release/network-probe"

build-macos:
	@echo "Building for macOS x86_64..."
	cargo build --release --target x86_64-apple-darwin
	@echo "Building for macOS arm64..."
	cargo build --release --target aarch64-apple-darwin
	@echo "macOS binaries built at target/x86_64-apple-darwin/release/ and target/aarch64-apple-darwin/release/"

build-windows:
	@echo "Building for Windows x86_64..."
	cargo build --release --target x86_64-pc-windows-gnu
	@echo "Windows binary built at target/x86_64-pc-windows-gnu/release/network-probe.exe"

package-deb:
	@echo "Building Debian package..."
	cargo deb
	@echo "Debian package built at target/debian/"

package-rpm:
	@echo "Building RPM package..."
	cargo rpm build
	@echo "RPM package built at target/release/rpmbuild/RPMS/"

package-all: package-deb package-rpm
	@echo "All packages built successfully!"

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
