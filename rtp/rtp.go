package rtp

import "github.com/purpshell/meowcaller/srtp"

// RTP WARP framing: WhatsApp's 16-byte speech / 20-byte DTX headers (extension
// profile 0xdebe), Opus payload classifiers, and the send-side sequencer.

const (
	RtpPayloadTypeOpus          uint8  = 120
	WhatsappRtpExtensionProfile uint16 = 0xdebe
	WhatsappRtpHeaderSize       int    = 16
	WhatsappRtpHeaderDtxSize    int    = 20
	WhatsappRtpExtensionDtxWord uint32 = 0x30010000

	rtpVersion          = 2
	srtpAuthTagLen      = 10
	srtpAuthTagLenShort = 4
)

// OpusPrimingFrame1 is the Android first priming frame (18 bytes).
var OpusPrimingFrame1 = [18]byte{
	0x12, 0x36, 0x26, 0x2b, 0x4a, 0xc8, 0x2b, 0x09, 0xc9, 0x1f, 0x34, 0xc2, 0xd6, 0x7a, 0x01, 0x73,
	0x1b, 0x2e,
}

// OpusPrimingFrame1Wasm is the WASM/Web caller priming frame (24 bytes).
var OpusPrimingFrame1Wasm = [24]byte{
	0x32, 0x36, 0x26, 0x2b, 0x4a, 0xcb, 0x1b, 0x5f, 0xba, 0x91, 0x68, 0x7e, 0xb8, 0x50, 0x93, 0x58,
	0xe6, 0xd0, 0xa3, 0xa9, 0xd7, 0x1d, 0x81, 0x8c,
}

// OpusPrimingFrame2 is the second priming frame (5 bytes).
var OpusPrimingFrame2 = [5]byte{0x90, 0xb8, 0x14, 0x14, 0xc4}

// IsWhatsappOpusRtpPayload reports whether the payload type is WhatsApp Opus.
func IsWhatsappOpusRtpPayload(payloadType uint8) bool {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtp.rs#L28-L30
	// TODO
	// agent suggestion: payloadType == RtpPayloadTypeOpus || payloadType == 121.
	// human input:
	return false
}

// IsOpusDtxPayload reports DTX / comfort-noise frames (RFC 0x10, mlow 0x90, warmup).
func IsOpusDtxPayload(payload []byte) bool {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtp.rs#L33-L46
	// TODO
	// agent suggestion: len 0 false; len 1 -> 0x10|0x88|0x90; len<=15 -> (b0&0xf8)==0x08 || b0==0x0a ||
	// ((b0&0xf0)==0x30 && len<=6); else false.
	// human input:
	return false
}

// IsOpusMlowSpeechPayload reports mlow speech frames (20ms 0x48..0x4f or 60ms 0x50..0x57).
func IsOpusMlowSpeechPayload(payload []byte) bool {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtp.rs#L49-L55
	// TODO
	// agent suggestion: len>=18 && ((b0&0xf8)==0x48 || (b0&0xf8)==0x50).
	// human input:
	return false
}

// IsOpusPrimingPayload reports whether the payload equals a priming frame.
func IsOpusPrimingPayload(payload []byte) bool {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtp.rs#L57-L59
	// TODO
	// agent suggestion: bytes.Equal(payload, OpusPrimingFrame1[:]) || bytes.Equal(payload, OpusPrimingFrame2[:]).
	// human input:
	return false
}

// EstimateSrtpRtpWireBytes estimates the on-wire SRTP size (header + opus + tag).
func EstimateSrtpRtpWireBytes(opusPayload []byte) int {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtp.rs#L62-L79
	// TODO
	// agent suggestion: dtx => 20-byte header + short tag(4); else 16-byte header, short tag if
	// priming or len<=18 else 10; return header + len(payload) + tag.
	// human input:
	return 0
}

// RtpHeader is the fixed RTP header plus an optional 0xdebe extension word.
type RtpHeader struct {
	Marker         bool
	PayloadType    uint8
	SequenceNumber uint16
	Timestamp      uint32
	Ssrc           uint32
	ExtensionWord  *uint32 // nil = no 0xdebe extension word
}

// ByteSize is the on-wire header size (16, or 20 with an extension word).
func (h *RtpHeader) ByteSize() int {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtp.rs#L93-L99
	// TODO
	// agent suggestion: if h.ExtensionWord != nil return WhatsappRtpHeaderDtxSize else WhatsappRtpHeaderSize.
	// human input:
	return 0
}

// RtpHeaderByteLength returns the full on-wire header size (12 + CSRC + ext); ok=false if malformed.
func RtpHeaderByteLength(data []byte) (int, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtp.rs#L103-L127
	// TODO
	// agent suggestion: require len>=12 and version==2; cc=data[0]&0x0f, headerLen=12+cc*4; if X bit
	// set add 4 + extWords*4 (from data[headerLen+2..+4]); bounds-check each step.
	// human input:
	return 0, false
}

// IsRtpVersion2 reports a version-2 RTP packet.
func IsRtpVersion2(data []byte) bool {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtp.rs#L129-L131
	// TODO
	// agent suggestion: len>=12 && (data[0]>>6)&0x03 == rtpVersion.
	// human input:
	return false
}

// ParseRtpHeader parses the fixed RTP header fields (the extension word is not decoded).
func ParseRtpHeader(data []byte) (RtpHeader, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtp.rs#L134-L144
	// TODO
	// agent suggestion: RtpHeaderByteLength gate; then marker=data[1]>>7, pt=data[1]&0x7f, seq BE,
	// timestamp BE, ssrc BE, ExtensionWord nil.
	// human input:
	return RtpHeader{}, false
}

// EncodeRtpHeader encodes the RTP header (16 or 20 bytes with the 0xdebe extension).
func EncodeRtpHeader(header *RtpHeader) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtp.rs#L146-L175
	// TODO
	// agent suggestion: size=ByteSize; byte0=version<<6 (|0x10 if size>12); byte1=marker<<7|pt; seq BE;
	// ts BE; ssrc BE; if size>=16 write profile + ext-words count; if size>=20 write the extension word.
	// human input:
	return nil
}

// RtpStream is the send-side RTP sequencer: seq starts at 1, timestamp advances per packet.
type RtpStream struct {
	Ssrc             uint32
	seq              uint16
	timestamp        uint32
	samplesPerPacket uint32
	speechStarted    bool
	audioPacketIndex int
	warpPiggyback    bool
}

// NewRtpStream builds a sequencer for ssrc with samplesPerPacket per packet.
func NewRtpStream(ssrc, samplesPerPacket uint32, warpPiggyback bool) *RtpStream {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtp.rs#L189-L200
	// TODO
	// agent suggestion: &RtpStream{Ssrc: ssrc, seq: 1, samplesPerPacket: samplesPerPacket, warpPiggyback: warpPiggyback}.
	// human input:
	return nil
}

// resolveWarpExtension picks the extension word: the DTX word for DTX frames, else
// the warp audio piggyback word when piggyback is enabled, else nil.
func (s *RtpStream) resolveWarpExtension(dtx bool) *uint32 {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtp.rs#L201-L212
	// TODO
	// agent suggestion: if dtx return &WhatsappRtpExtensionDtxWord copy; if !warpPiggyback return nil;
	// idx=audioPacketIndex++, return srtp.AudioPiggybackExtensionFor(idx, true, srtp.WarpPiggybackStartPacket).
	// human input:
	return nil
}

// NextPacket builds the next RTP header for payload, latching the marker on the first speech frame.
func (s *RtpStream) NextPacket(payload []byte, marker bool) RtpHeader {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtp.rs#L213-L234
	// TODO
	// agent suggestion: dtx/priming classify; speech = !dtx && !priming; useMarker = marker || (speech &&
	// !speechStarted); latch speechStarted; build header with resolveWarpExtension(dtx); advance seq/timestamp.
	// human input:
	return RtpHeader{}
}

// NextPreSpeechPacket builds a pre-speech ladder packet (advances seq/timestamp, no marker/latch).
func (s *RtpStream) NextPreSpeechPacket() RtpHeader {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtp.rs#L235-L247
	// TODO
	// agent suggestion: header with marker false, resolveWarpExtension(false); advance seq/timestamp.
	// human input:
	return RtpHeader{}
}

var _ = srtp.WarpPiggybackStartPacket
