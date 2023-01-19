# Copyright 2022 The OWASP Coraza contributors
# SPDX-License-Identifier: Apache-2.0

FROM rust:1-alpine

RUN apk add --no-cache binaryen && rustup target add wasm32-wasi

ADD . /aho-corasick
WORKDIR /aho-corasick
ENV RUSTFLAGS "-C target-feature=-crt-static -C link-args=--export=malloc -C link-args=--export=free"
RUN cargo build --release --target wasm32-wasi

RUN wasm-opt -o target/wasm32-wasi/release/aho_corasick.wasm --flatten --rereloop --converge -O3 target/wasm32-wasi/release/aho_corasick.wasm

CMD ["cp", "target/wasm32-wasi/release/libaho_corasick.a", "target/wasm32-wasi/release/aho_corasick.wasm", "/out/"]
