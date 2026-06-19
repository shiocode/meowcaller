# meowcaller

A clean-room, **pure-Go** WhatsApp 1:1 calling library — signaling, keying,
transport, and the MLow audio codec — built as a byte-exact, KAT-verified port of
the validated Rust reference (`whatsapp-rust`).

No WASM bridge. No cgo for the protocol. No code inherited from earlier attempts.

**Start here:** [`PLAN.md`](PLAN.md) — the full execution plan: the reference and
its known-answer vectors, the KAT-driven port methodology, the dependency-ordered
module list, milestones, and a worked example you can copy for every module.

Status: planning. Nothing implemented yet — execute `PLAN.md` module by module,
each green against its vector before moving on.
