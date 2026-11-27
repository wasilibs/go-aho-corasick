# Copyright 2022 The OWASP Coraza contributors
# SPDX-License-Identifier: Apache-2.0

FROM rust:1-alpine

# wasm32-wasi was renamed to wasm32-wasip1 and removed as a target name in Rust 1.84.
RUN apk add --no-cache binaryen && rustup target add wasm32-wasip1

ADD buildtools/aho-corasick /aho-corasick
WORKDIR /aho-corasick
ENV RUSTFLAGS="-C target-feature=-crt-static -Ctarget-feature=+simd128 -C link-args=--export=malloc -C link-args=--export=free"
RUN cargo build --release --target wasm32-wasip1

RUN wasm-opt -o target/wasm32-wasip1/release/aho_corasick.wasm --flatten --rereloop --converge -O3 target/wasm32-wasip1/release/aho_corasick.wasm

CMD ["cp", "target/wasm32-wasip1/release/libaho_corasick.a", "target/wasm32-wasip1/release/aho_corasick.wasm", "/out/"]
