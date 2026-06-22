package rtp

// RTCP: WhatsApp compact reports (PT 208/209) and a Sender Report (PT 200). The
// SR's NTP timestamp is taken as a nowMs argument so this stays pure/no-clock.

const (
	RtcpPtSr         uint8 = 200
	RtcpPtWaCompact  uint8 = 208
	RtcpPtWaCompact2 uint8 = 209
	RtcpHeaderLen    int   = 8
	SrtcpTrailerLen  int   = 14

	ntpUnixOffsetSecs uint64 = 2208988800
)

// IsRtcpPacket reports whether data is an RTCP packet (vs a WhatsApp RTP packet).
func IsRtcpPacket(data []byte) bool {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtcp.rs#L16-L28
	// TODO
	// agent suggestion: require len>=RtcpHeaderLen+SrtcpTrailerLen and version==2; exclude WhatsApp RTP
	// (X=1 and data[1]&0x7f==Opus); then data[1] >= 64.
	// human input:
	return false
}

// RtcpPayloadType returns the RTCP payload type; ok=false if not an RTCP packet.
func RtcpPayloadType(data []byte) (uint8, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtcp.rs#L30-L32
	// TODO
	// agent suggestion: if IsRtcpPacket return data[1], true else 0, false.
	// human input:
	return 0, false
}

// ParseRtcpSenderSsrc returns the sender SSRC (bytes 4-7); ok=false if malformed.
func ParseRtcpSenderSsrc(data []byte) (uint32, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtcp.rs#L34-L39
	// TODO
	// agent suggestion: require len>=8 and version==2; return BE(data[4:8]), true.
	// human input:
	return 0, false
}

// RtcpSenderStats are the Sender Report counters.
type RtcpSenderStats struct {
	PacketsSent  uint32
	OctetsSent   uint32
	RtpTimestamp uint32
}

// BuildCompactRtcp208 builds the 12-byte compact RTCP (PT 208, RC=1).
func BuildCompactRtcp208(localSsrc, remoteSsrc uint32) [12]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtcp.rs#L49-L58
	// TODO
	// agent suggestion: buf[0]=0x81, buf[1]=208, buf[3]=2; BE localSsrc@4:8, BE remoteSsrc@8:12.
	// human input:
	return [12]byte{}
}

// BuildCompactRtcp209 builds the 8-byte compact RTCP (PT 209, RC=1).
func BuildCompactRtcp209(localSsrc uint32) [8]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtcp.rs#L61-L69
	// TODO
	// agent suggestion: buf[0]=0x81, buf[1]=209, buf[3]=1; BE localSsrc@4:8.
	// human input:
	return [8]byte{}
}

// BuildSenderReport builds the 28-byte Sender Report (PT 200, RC=0); nowMs is wall-clock ms.
func BuildSenderReport(localSsrc uint32, stats *RtcpSenderStats, nowMs uint64) [28]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/rtcp.rs#L72-L89
	// TODO
	// agent suggestion: buf[0]=0x80, buf[1]=200, buf[3]=6; BE localSsrc; NTP sec = (nowMs/1000 +
	// ntpUnixOffsetSecs) truncated to u32; NTP frac = uint32(float64(nowMs%1000)/1000*2^32); then BE
	// rtpTimestamp/packetsSent/octetsSent.
	// human input:
	return [28]byte{}
}
