.PHONY: all fix lint test test-all test-no-default fmt fmt-check publish-dry run debug build build-snapshot release release-snapshot check clean

all: check

fmt:
	cargo fmt --all

fmt-check:
	cargo fmt --all -- --check

clippy:
	cargo clippy --all-features -- -D warnings

lint: fmt-check clippy

fix:
	cargo fix --all-features --allow-dirty

test:
	cargo test --all --all-features

test-all:
	cargo test --all

test-no-default:
	cargo test --all --no-default-features

check: fix lint test test-all test-no-default publish-dry

publish-dry:
	cargo publish --manifest-path crates/acta-build/Cargo.toml --dry-run --allow-dirty
	cargo publish --dry-run --allow-dirty

run:
	cargo run -p acta-debug --all-features

check-debug:
	cargo check -p acta-debug --all-features

build:
	goreleaser build --snapshot --clean --skip=before --config .goreleaser.yml

build-snapshot:
	goreleaser release --snapshot --clean --skip=before --config .goreleaser.yml

release:
	goreleaser release --clean --skip=before --config .goreleaser.yml

release-snapshot:
	goreleaser release --snapshot --clean --skip=before --config .goreleaser.yml

clean:
	cargo clean
	rm -rf dist/
