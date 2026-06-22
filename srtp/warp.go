package srtp

// WARP RTP extension constants and the WARP MESSAGE-INTEGRITY tag (HMAC-SHA1
// appended to protected packets).

const WarpExtProfile uint16 = 0xdebe

const (
	WarpMITagLen             = 4
	WarpPiggybackStartPacket = 2
)

// WarpAudioPiggybackExt is the audio piggyback extension word (big-endian bytes).
var WarpAudioPiggybackExt = [4]byte{0x30, 0x01, 0x00, 0x00}

// AudioPiggybackExtensionFor returns the audio piggyback extension word for
// packetIndex, or nil for the first packets / when disabled.
func AudioPiggybackExtensionFor(packetIndex int, enabled bool, startPacket int) *uint32 {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/warp.rs#L15-L24
	// TODO
	// agent suggestion: if !enabled || packetIndex < startPacket return nil; else w :=
	// binary.BigEndian.Uint32(WarpAudioPiggybackExt[:]); return &w.
	// human input:
	return nil
}

// ComputeWarpMITag is the WARP MI tag: the first tagLen bytes of
// HMAC-SHA1(authKey, packetWithoutTag || roc_be32).
func ComputeWarpMITag(authKey, packetWithoutTag []byte, roc uint32, tagLen int) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/warp.rs#L27-L38
	// TODO
	// agent suggestion: mac := hmac.New(sha1.New, authKey); mac.Write(packetWithoutTag); write BE(roc);
	// return mac.Sum(nil)[:tagLen]. hmac.New never errors on key length.
	// human input:
	return nil
}

// AppendWarpMITag appends the WARP MI tag to a protected packet.
func AppendWarpMITag(authKey, packetWithoutTag []byte, roc uint32, tagLen int) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/warp.rs#L41-L52
	// TODO
	// agent suggestion: tag := ComputeWarpMITag(...); return append(append([]byte(nil),
	// packetWithoutTag...), tag...).
	// human input:
	return nil
}
