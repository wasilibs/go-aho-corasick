// Copyright The OWASP Coraza contributors
// SPDX-License-Identifier: Apache-2.0

extern crate aho_corasick;

use std::ffi::CStr;
use std::mem::MaybeUninit;
use std::os::raw::c_char;
use std::slice;
use std::str;
use aho_corasick::{AhoCorasick, AhoCorasickBuilder, FindIter, FindOverlappingIter, MatchKind};

#[no_mangle]
pub extern "C" fn new_matcher(patterns_ptr: *const c_char, patterns_len: usize, ascii_case_insensitive: bool, dfa: bool, match_kind: MatchKind) -> Box<AhoCorasick> {
    let mut patterns = Vec::new();

    let mut off = 0;
    while off < patterns_len {
        let pattern_slice = unsafe { CStr::from_ptr(patterns_ptr.offset(off as isize)) };
        let pattern = pattern_slice.to_str().unwrap();
        patterns.push(pattern);
        off += pattern.len() + 1;
    }

    let ac = AhoCorasickBuilder::new()
        .ascii_case_insensitive(ascii_case_insensitive)
        .dfa(dfa)
        .match_kind(match_kind)
        .build(patterns);

    return Box::new(ac);
}

#[no_mangle]
pub extern "C" fn delete_matcher(_matcher: Box<AhoCorasick>) {
    // Box takes ownership and will release
}

#[no_mangle]
pub extern "C" fn find_iter(ac: &AhoCorasick, value_ptr: usize, value_len: usize) -> Box<FindIter<usize>> {
    let value = ptr_to_string(value_ptr, value_len);
    return Box::new(ac.find_iter(value));
}

#[no_mangle]
pub extern "C" fn find_iter_next(iter: &mut FindIter<usize>, pattern: &mut usize, start: &mut usize, end: &mut usize) -> bool {
    iter.next().map(|m| {
        *pattern = m.pattern();
        *start = m.start();
        *end = m.end();
        true
    }).unwrap_or(false)
}

#[no_mangle]
pub extern "C" fn find_iter_delete(_iter: Box<FindIter<usize>>) {
    // Box takes ownership and will release
}

#[no_mangle]
pub extern "C" fn overlapping_iter(ac: &AhoCorasick, value_ptr: usize, value_len: usize) -> Box<FindOverlappingIter<usize>> {
    let value = ptr_to_string(value_ptr, value_len);
    return Box::new(ac.find_overlapping_iter(value));
}

#[no_mangle]
pub extern "C" fn overlapping_iter_next(iter: &mut FindOverlappingIter<usize>, pattern: &mut usize, start: &mut usize, end: &mut usize) -> bool {
    iter.next().map(|m| {
        *pattern = m.pattern();
        *start = m.start();
        *end = m.end();
        true
    }).unwrap_or(false)
}

#[no_mangle]
pub extern "C" fn overlapping_iter_delete(_iter: Box<FindOverlappingIter<usize>>) {
    // Box takes ownership and will release
}

#[no_mangle]
pub extern "C" fn matches(ac: &mut AhoCorasick, value_ptr: usize, value_len: usize, n: usize, matches: *mut usize) -> usize {
    let value = ptr_to_string(value_ptr, value_len);
    std::mem::forget(&value);

    let mut num = 0;
    for value in ac.find_iter(value.as_bytes()) {
        if num == n {
            break;
        }
        unsafe {
            *matches.offset(2*num as isize) = value.start();
            *matches.offset((2*num+1) as isize) = value.end();
        }
        num += 1;
    }

    return num
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
