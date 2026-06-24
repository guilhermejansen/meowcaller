package main

import (
	"testing"

	"go.mau.fi/whatsmeow/types"
)

// TestParseCallTarget pins the input → JID mapping: a bare number (with or without a
// leading +) becomes a phone JID, a phone JID passes through, and a LID is recognized
// as a LID. Regression for the "+961… → unknown user server" bug, where ParseJID
// silently stuffed a bare number into the server field instead of erroring.
func TestParseCallTarget(t *testing.T) {
	cases := []struct {
		in         string
		wantUser   string
		wantServer string
	}{
		{"+96170000007", "96170000007", types.DefaultUserServer},
		{"96170000007", "96170000007", types.DefaultUserServer},
		{"  +96170000007  ", "96170000007", types.DefaultUserServer},
		{"96170000007@s.whatsapp.net", "96170000007", types.DefaultUserServer},
		{"123456789012345@lid", "123456789012345", types.HiddenUserServer},
	}
	for _, c := range cases {
		jid, err := parseCallTarget(c.in)
		if err != nil {
			t.Errorf("parseCallTarget(%q) error: %v", c.in, err)
			continue
		}
		if jid.User != c.wantUser || jid.Server != c.wantServer {
			t.Errorf("parseCallTarget(%q) = %s@%s; want %s@%s",
				c.in, jid.User, jid.Server, c.wantUser, c.wantServer)
		}
	}

	// A bare number must never land in the server field (the original bug).
	jid, _ := parseCallTarget("+96170000007")
	if jid.Server != types.DefaultUserServer {
		t.Errorf("bare number leaked into server: %q", jid.Server)
	}
	if _, err := parseCallTarget("   "); err == nil {
		t.Error("empty target should error")
	}
}
