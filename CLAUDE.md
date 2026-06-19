# CLAUDE.md

You are working on **meowcaller**, a clean-room, pure-Go implementation of the
WhatsApp 1:1 VoIP stack (signaling, keying, transport, media). Read these before
doing anything:

1. **[`AGENTS.md`](AGENTS.md)** — the build protocol. It is binding. In short: you
   are **not autonomous**. You build **one module at a time**; you **scaffold**
   function envelopes with `// TODO` bodies and then **stop for the human** to
   direct the logic. You explain *why* in the conversation, never in code comments.
   Verify against the test vector or it is not done.
2. **[`PLAN.md`](PLAN.md)** — the model and the standard.
3. **[`MODULES.md`](MODULES.md)** — the module registry and build order. Pick the
   next `planned` module in dependency order.
4. **[`datasheets/`](datasheets/)** — one datasheet per module. Each has the
   reference source verbatim, the Go envelope (signatures), and suggestions. Build
   a module only from its datasheet.
5. **[`GLOSSARY.md`](GLOSSARY.md)** — every acronym/term used in the datasheets.

Non-negotiables: the Go code never names or imports any reference library (only a
plain URL to the spec/decision repo is allowed in a comment); commits are
`(<module>: <change>)` and update `CHANGELOG.md`; commit but do not push; no real
PII in tests. Scope is 1:1 calls.

If you are about to write a function body that involves a real engineering choice,
**stop and ask** instead. Scaffold, explain in chat, and let the human decide.
