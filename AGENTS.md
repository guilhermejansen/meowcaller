# AGENTS.md — how meowcaller gets built

This library is built **module by module, under human audit, in real time**. The
human reviewer directs the engineering; agents prepare and explain. This file is
the working protocol. It is binding.

Unfamiliar with a term or acronym (TOC, KAT, SID, VAD, LTP, NLSF, PVQ, SRTP,
SFrame, WARP, …)? See [`GLOSSARY.md`](GLOSSARY.md).

> If you are an agent reading this: you are not autonomous here. You scaffold and
> explain; the human decides how logic flows and what trade-offs are made. When in
> doubt, stop and ask. A wrong guess committed quietly is the worst outcome.

## The prime directives

1. **Do not decide logic. Scaffold it.** When you reach a function whose behavior
   involves a real engineering choice (an algorithm, a data layout, an
   error-handling strategy, a buffering decision), write the **signature, the doc
   comment, and the datasheet reference**, leave the body as a `// TODO` with the
   open question stated, and **stop for the human**. Do not fill it in to "keep
   moving."
2. **One module at a time. Take the break.** Finish a single, reviewable unit —
   often just a scaffold, or one function the human approved — then pause for
   review and approval before continuing. No multi-module sprints. The human is
   watching this get built and will say when to proceed.
3. **Explain the why in the conversation, never in the code.** As you scaffold
   each thing, say in the chat — to the human, in plain language — *why* it exists
   and what it does, so the human understands what is being built **without
   reading any code comment**. The backing detail (Rust source, constants,
   validation) lives in the datasheet you read; you speak the relevant part aloud.
   The human should never have to read a comment to follow the reasoning. Code
   carries logic; the conversation carries the why.
4. **Verify against vectors, never vibes.** A behavior is correct only when its
   KAT passes. Reverse-engineered names and analysis notes are frequently wrong;
   the vector is the proof. If a module has no vector, that is a decision point
   for the human, not a license to guess.

## The build loop (per module)

```
1. READ      datasheets/<module>.md — three parts: the reference source VERBATIM
             (the authoritative ground truth), the Go envelope (signatures), and
             implementation suggestions. Implement from the verbatim source; the
             suggestions are guidance, not proof.
2. SCAFFOLD  create the package file: types, exported signatures, doc comments,
             the KAT test wired to the (copied) vector — but function bodies are
             `// TODO` stubs that state the open question.  → COMMIT, PAUSE.
3. DIRECT    the human reviews the scaffold and decides how each body should work
             (or approves translating the embedded reference 1:1 into Go). One
             function at a time.
4. IMPLEMENT the approved function(s) only, as clean Go that never names the
             reference. Keep the KAT test running.  → COMMIT per fn, PAUSE.
5. VERIFY    when the module's body is complete, its KAT must pass.
             Update the datasheet status and CHANGELOG.  → COMMIT, PAUSE.
```

If the datasheet for the module does not exist yet, write it first (embed the
reference source verbatim + the Go target, per `datasheets/_TEMPLATE.md`), get it
reviewed, and only then scaffold. Datasheets are written one module at a time.

Each arrow `→ PAUSE` is a real stop: hand control back to the human. Do not chain
steps without approval.

## Scaffolding standard

A scaffold is a complete, compiling skeleton with no logic. The doc comment is a
normal, brief Go doc comment; there is **no** provenance/spec reference in the
code (that was already said in the conversation):

```go
// SealFrame encrypts one media frame with the participant SFrame key.
func (s *Session) SealFrame(plaintext []byte) ([]byte, error) {
	// TODO(human): nonce + whether the header is authenticated.
	return nil, errNotImplemented
}
```

When you write this, you say in the chat — not in the file — something like: "I'm
scaffolding `SealFrame` for the send path's per-frame encryption; it comes from
`sframe.rs`, it's AES-GCM, and the open decision is the 16-byte LE-counter nonce
and whether the varint header is AAD." That is the explanation; the code stays
clean.

It must `go build` and `go vet` cleanly. The KAT test exists and **fails** (or
skips with a clear reason) until the body lands — never a fake pass.

## Comment policy

Comments earn their place or they do not exist:

- **`// TODO(human): ...`** — an open decision or unfinished body.
- **`// ASSUMPTION: ...`** — a choice made without full confirmation, stating what
  would invalidate it.
- **A short context note** — only when a future reader would otherwise lose
  non-obvious context (a magic constant's origin, a byte-order quirk, a deviation
  from the reference and why).
- **Doc comments** on exported identifiers, per Go convention — a brief statement
  of what it is. **No** spec/ref/KAT provenance lines in code; that belongs in the
  conversation and the datasheet, not in comments the human must read.

Do **not** narrate what the code plainly does, and do **not** use comments to
explain *why* — say the why out loud. Clean code is the default; comments are the
exception.

## Commits and changelog

- One module change per commit. Subject: `(<module>: <imperative change>)`.
  Examples: `(mlow/toc: scaffold smpl TOC parser)`,
  `(srtp/e2e: implement RFC3711 AES-CM PRF)`,
  `(mlow/pitch: KAT-verify against pitch_vectors.json)`.
- Every commit updates [`CHANGELOG.md`](CHANGELOG.md) under the module with the
  new state (`scaffolded` / `implemented` / `KAT-verified`).
- Commit messages state **what was validated** when relevant (which vector,
  pass/fail). No attribution lines. No pushing unless the human asks.

## Decision artifacts (ADRs in wacrg)

When a conversation over a specific implementation detail **ends with a decision**,
that decision is recorded — but only on the human's direction, never on your own.

- **When:** the conversation has concluded and the human says to record it. Not
  mid-discussion, not pre-emptively, not for every change — only for decisions
  worth a timestamp.
- **Where:** wacrg (`docs/decisions/`), as a timestamped Markdown record stating
  the decision, the context, the options weighed, and why this one. wacrg is the
  spec home; the decision becomes part of the agreed record.
- **How it connects to code:** the Go file may carry a single plain-URL comment to
  the decision artifact where it genuinely helps a future reader find the rationale
  — a pointer, not an explanation. Nothing else about the reference or the why goes
  in the code.

You must know how to write one (a clear, dated ADR) and must **never** write one
unprompted. Generating decision artifacts autonomously is a violation.

## Explaining as you go (in conversation)

Narrate the why **proactively**, in the chat, as you work — so the human follows
what is being built and why in real time, without reading the code to find out.
Before scaffolding a module, say what it is and why it is next; as you place each
envelope, say what it does and where it comes from; when you make or defer a
decision, say so.

If the human asks "what is this / is this right / what did we validate?", answer
concretely from the datasheet and the vector: the Rust source location, the
constants, the KAT file and what it covers, and any open assumptions. If you
cannot answer from the datasheet, the datasheet is incomplete — say so and fix it
before proceeding. None of this lives in code comments.

## What never happens here

- No autonomous multi-module runs.
- No filling a function body with a guess to avoid stopping.
- No "it compiles, ship it" — green KAT or it is not done.
- No silently copying logic whose meaning you cannot explain.
- No naming or alluding to the reference library anywhere in the Go code.
- No writing a decision artifact (ADR) without the human's direction.
- No reuse of the old dublin/meowmeow calling code as a source.
