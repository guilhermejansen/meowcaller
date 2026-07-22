# WhatsApp Web diagnostics

This is a minimal call logger for comparing WhatsApp Web behavior with
meowcaller. It has no user interface and only records call-related signaling,
VoIP stack calls, internal VoIP logs, call-model changes, and media acquisition.

Start the local collector from the repository root:

```sh
go run ./diagnostics/extension-server
```

Then open `chrome://extensions`, enable developer mode, choose **Load unpacked**,
and select `diagnostics/extension`. Reload WhatsApp Web before placing a call.
The collector writes compact JSONL files under `diagnostics/captures/`.

The extension sends data only to `http://127.0.0.1:3219/events`. Remove or
disable it after collecting a diagnostic call.
