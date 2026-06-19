<!-- Datasheet = three things only: the reference source VERBATIM, the Go envelope
     (signatures, no bodies), and implementation suggestions. No behavioral summary,
     no implementation. The verbatim source is the only authoritative content. -->

# Datasheet: `<package>/<module>`

One line: what this module is and where it sits in the call stack (signaling /
keying / transport / media).

**Validation vector:** `<name>.json` — the test vector this module must match.
Copy it verbatim from the reference's test data into `<package>/testdata/`.

## Reference source (verbatim — authoritative)

The reference implementation of this behavior, pasted **exactly as it is**. Do not
paraphrase, summarize, reformat, or "clean up" a single line. This block is the
ground truth; everything else on this page is secondary to it.

```rust
// pasted verbatim, in full
```

## Go envelope (signatures only)

The corresponding Go declarations — exported types and function **signatures with
no bodies**. This is the surface to implement; it is not the implementation.

```go
// types + func signatures only — no logic
```

## Implementation suggestions (guidance, not authoritative)

Non-binding notes for whoever writes the bodies. Suggestions, not proofs — the
verbatim source above governs. Useful kinds of note: Go type/width mapping
(e.g. `i32` → `int`), slice vs array, endianness, error-return vs panic, allocation
or buffering choices, and any spot that needs a human decision (mark it
`TODO(human)`).

- ...
