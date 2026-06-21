package util

// HKDFSHA256 derives length bytes of key material from ikm using HKDF-SHA256
// (RFC 5869): an HMAC-SHA256 extract keyed by salt, then expand with info. Every
// VoIP key schedule (SRTP session keys, SFrame keys, the WARP auth key) reduces
// to this one shape.
func HKDFSHA256(salt, ikm, info []byte, length int) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/mod.rs#L32-L39
	// TODO
	// agent suggestion: use the Go 1.25 stdlib crypto/hkdf (zero new deps, vs the
	// datasheet's x/crypto/hkdf which adds a dependency the PLAN forbids beyond
	// protobuf): okm, err := hkdf.Key(sha256.New, ikm, salt, string(info), length).
	// The reference panics (.expect / debug_assert) when length exceeds 255*32 =
	// 8160 bytes; hkdf.Key returns an error there, so panic on err to match (the
	// signature has no error return). Salt passes straight through, nil included.
	// human input:
	return nil
}
