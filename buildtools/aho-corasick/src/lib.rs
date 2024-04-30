// Copyright The OWASP Coraza contributors
// SPDX-License-Identifier: Apache-2.0

extern crate aho_corasick;

use std::slice;
use std::str;
use aho_corasick::{AhoCorasick, AhoCorasickBuilder, AhoCorasickKind, FindIter, FindOverlappingIter, MatchKind};

#[no_mangle]
pub extern "C" fn new_matcher(patterns_ptr: usize, patterns_len: *const usize, num_patterns: usize, ascii_case_insensitive: bool, dfa: bool, match_kind: MatchKind) -> Box<AhoCorasick> {
    let mut patterns = Vec::new();

    let mut off = 0usize;
    for i in 0..num_patterns {
        unsafe {
            let len = *patterns_len.offset(i as isize);
            let pattern = ptr_to_string(patterns_ptr+off, len);
            patterns.push(pattern);
            off += len;
        }
    }

    let mut ac = AhoCorasickBuilder::new();
    ac
        .ascii_case_insensitive(ascii_case_insensitive)
        .match_kind(match_kind);

    if dfa {
        ac.kind(Some(AhoCorasickKind::DFA));
    }

    return Box::new(ac.build(patterns).unwrap());
}

#[no_mangle]
pub extern "C" fn delete_matcher(_matcher: Box<AhoCorasick>) {
    // Box takes ownership and will release
}

#[no_mangle]
pub extern "C" fn find_iter(ac: &AhoCorasick, value_ptr: usize, value_len: usize) -> Box<FindIter> {
    let value = ptr_to_string(value_ptr, value_len);
    return Box::new(ac.find_iter(value));
}

#[no_mangle]
pub extern "C" fn find_iter_next(iter: &mut FindIter, pattern: &mut usize, start: &mut usize, end: &mut usize) -> bool {
    iter.next().map(|m| {
        *pattern = m.pattern().as_usize();
        *start = m.start();
        *end = m.end();
        true
    }).unwrap_or(false)
}

#[no_mangle]
pub extern "C" fn find_iter_delete(_iter: Box<FindIter>) {
    // Box takes ownership and will release
}

#[no_mangle]
pub extern "C" fn overlapping_iter(ac: &AhoCorasick, value_ptr: usize, value_len: usize) -> Box<FindOverlappingIter> {
    let value = ptr_to_string(value_ptr, value_len);
    return Box::new(ac.find_overlapping_iter(value));
}

#[no_mangle]
pub extern "C" fn overlapping_iter_next(iter: &mut FindOverlappingIter, pattern: &mut usize, start: &mut usize, end: &mut usize) -> bool {
    iter.next().map(|m| {
        *pattern = m.pattern().as_usize();
        *start = m.start();
        *end = m.end();
        true
    }).unwrap_or(false)
}

#[no_mangle]
pub extern "C" fn overlapping_iter_delete(_iter: Box<FindOverlappingIter>) {
    // Box takes ownership and will release
}

#[no_mangle]
pub extern "C" fn matches(ac: &mut AhoCorasick, value_ptr: usize, value_len: usize, limit: usize, num: &mut usize) -> *const usize {
    let mut matches = Vec::new();
    let value = ptr_to_string(value_ptr, value_len);

    let mut count = 0;
    for value in ac.find_iter(value.as_bytes()) {
        if count == limit {
            break;
        }

        matches.push(value.pattern().as_usize());
        matches.push(value.start());
        matches.push(value.end());

        count += 1;
    }

    let b = matches.into_boxed_slice();
    let ptr = b.as_ptr();
    let len = b.len(); // Same as count since into_boxed_slice() truncates
    std::mem::forget(b);
    *num = len;
    return ptr;
}

#[no_mangle]
pub extern "C" fn matches_delete(ptr: *const usize, len: usize) {
    unsafe {
        let _ = Vec::from_raw_parts(ptr as *mut usize, len, len);
    }
}

extern "C" {
    fn __wasm_call_ctors();
}

// Rust flag for reactor mode requires nightly, simpler to just implement it ourselves.
// https://github.com/rust-lang/rust/pull/79997
#[no_mangle]
#[cfg(target_arch = "wasm32-wasi")]
pub unsafe extern "C" fn __initialize() {
    __wasm_call_ctors()
}

/// Returns a string from WebAssembly compatible numeric types representing
/// its pointer and length.
fn ptr_to_string(ptr: usize, len: usize) -> &'static str {
    unsafe {
        let slice = slice::from_raw_parts(ptr as *mut u8, len as usize);
        return str::from_utf8_unchecked(slice);
    }
}
