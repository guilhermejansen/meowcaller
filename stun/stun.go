package stun

// STUN/WARP relay framing: an RFC 5389 TLV encoder with WhatsApp's
// MESSAGE-INTEGRITY (HMAC-SHA1) and FINGERPRINT (CRC-32), the allocate-request
// builders, the consent ping, and the response parsers.

const (
	stunMagic          = 0x2112a442
	stunFingerprintXor = 0x5354554e
	stunXorPort        = 0x2112

	attrMessageIntegrity      = 0x0008
	attrFingerprint           = 0x8028
	attrErrorCode             = 0x0009
	attrRelayToken            = 0x4000
	attrStreamDescriptors     = 0x4024
	attrWasmRelayEndpoint     = 0x0016
	attrSenderSubscriptionsV2 = 0x4025
)

// STUN message types.
const (
	MsgBindingRequest  uint16 = 0x0001
	MsgAllocateRequest uint16 = 0x0003
	MsgBindingSuccess  uint16 = 0x0101
	MsgAllocateSuccess uint16 = 0x0103
	MsgAllocateError   uint16 = 0x0113
	MsgWhatsappPing    uint16 = 0x0801
	MsgWhatsappPong    uint16 = 0x0802
)

// stunXorAddr is the magic-cookie prefix XORed into XOR-MAPPED addresses.
var stunXorAddr = [4]byte{0x21, 0x12, 0xa4, 0x42}

// wasmStreamDescriptorsTemplate is the WASM/Web StreamDescriptors blob (attr 0x4024).
var wasmStreamDescriptorsTemplate = []byte{
	0x0a, 0x06, 0x18, 0xca, 0xbc, 0x85, 0xae, 0x04, 0x0a, 0x08, 0x10, 0x01, 0x18, 0xa5, 0xac, 0xaf,
	0xae, 0x0a, 0x0a, 0x08, 0x10, 0x02, 0x18, 0xd6, 0xa4, 0xe6, 0xf9, 0x0f, 0x0a, 0x08, 0x08, 0x01,
	0x18, 0xf7, 0xdd, 0x9e, 0xb6, 0x0a, 0x0a, 0x0a, 0x08, 0x01, 0x10, 0x01, 0x18, 0xab, 0xcc, 0xb1,
	0xf3, 0x0d, 0x0a, 0x0a, 0x08, 0x01, 0x10, 0x02, 0x18, 0xda, 0xda, 0xef, 0x8a, 0x05, 0x0a, 0x08,
	0x08, 0x02, 0x18, 0xc5, 0xe9, 0xec, 0x8e, 0x0b, 0x0a, 0x0a, 0x08, 0x02, 0x10, 0x01, 0x18, 0xfd,
	0xc2, 0xb1, 0xb6, 0x0f, 0x0a, 0x0a, 0x08, 0x02, 0x10, 0x02, 0x18, 0xb0, 0x97, 0xf7, 0xb2, 0x09,
}

func pad4(n int) int {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L44-L46
	// TODO
	// agent suggestion: return (4 - (n % 4)) % 4.
	// human input:
	return 0
}

// stunAttr encodes one STUN attribute (type, length, value, 4-byte alignment pad).
func stunAttr(attrType uint16, value []byte) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L49-L57
	// TODO
	// agent suggestion: BE attrType, BE uint16(len(value)), value, then pad4(len) zero bytes.
	// human input:
	return nil
}

// stunFingerprint is the STUN FINGERPRINT CRC-32 (reflected IEEE poly 0xedb88320),
// which is exactly stdlib hash/crc32.ChecksumIEEE.
func stunFingerprint(buf []byte) uint32 {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L60-L69
	// TODO
	// agent suggestion: return crc32.ChecksumIEEE(buf) (same reflected IEEE poly as the verbatim
	// bitwise loop; proven by the crc32_abc KAT). Resolves the stdlib-vs-port TODO toward stdlib.
	// human input:
	return 0
}

func stunPseudoHeader(msgType, msgLen uint16, transactionID [12]byte) [20]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L71-L78
	// TODO
	// agent suggestion: BE msgType@0, BE msgLen@2, BE stunMagic@4, transactionID@8:20.
	// human input:
	return [20]byte{}
}

// EncodeStunRequest encodes a STUN request: header + attrs, then optional
// MESSAGE-INTEGRITY (nil integrityKey skips it) and optional FINGERPRINT.
func EncodeStunRequest(msgType uint16, transactionID [12]byte, attrs []byte, integrityKey []byte, includeFingerprint bool) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L82-L118
	// TODO
	// agent suggestion: body=copy(attrs); if integrityKey!=nil compute HMAC-SHA1 over
	// pseudoHeader(msgLen=len(body)+24)||body, append stunAttr(MI, mac); if includeFingerprint
	// crc=ChecksumIEEE(pseudoHeader(len(body)+8)||body)^stunFingerprintXor, append stunAttr(FP, crc);
	// out = msgType||uint16(len(body))||magic||tx||body.
	// human input:
	return nil
}

// CreateNativeSenderSubscription is a native WA sender sub: 1-byte count + BE SSRC.
func CreateNativeSenderSubscription(ssrc uint32) [5]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L121-L126
	// TODO
	// agent suggestion: buf[0]=1; BE ssrc into buf[1:5].
	// human input:
	return [5]byte{}
}

// EncodeXorRelayEndpoint XOR-encodes an IPv4:port into 6 bytes; ok=false if ipv4
// is not exactly four dotted octets.
func EncodeXorRelayEndpoint(ipv4 string, port uint16) ([6]byte, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L129-L144
	// TODO
	// agent suggestion: split ipv4 on '.', parse each as u8 (drop unparseable); require exactly 4;
	// BE (port ^ stunXorPort) into buf[0:2]; buf[2+i] = octet[i] ^ stunXorAddr[i].
	// human input:
	return [6]byte{}, false
}

// createWasmRelayEndpointAttr is the WASM attr 0x0016 value: 00 01 + 6-byte endpoint.
func createWasmRelayEndpointAttr(endpointXor [6]byte) [8]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L147-L152
	// TODO
	// agent suggestion: BE 1 into buf[0:2]; copy endpointXor into buf[2:8].
	// human input:
	return [8]byte{}
}

// BuildWasmStunAllocateRequest builds the WASM/Web DataChannel Allocate: 0x4000
// token + 0x4024 stream desc + 0x0016 endpoint + MI, no FP.
func BuildWasmStunAllocateRequest(transactionID [12]byte, relayToken []byte, endpointXor [6]byte, integrityKey []byte) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L155-L177
	// TODO
	// agent suggestion: attrs = stunAttr(token) || stunAttr(streamDesc, template) ||
	// stunAttr(endpoint, createWasmRelayEndpointAttr); EncodeStunRequest(Allocate, tx, attrs, key, false).
	// human input:
	return nil
}

// BuildWhatsappPing builds the WhatsApp consent ping (type 0x0801, empty body).
func BuildWhatsappPing(transactionID [12]byte) [20]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L180-L186
	// TODO
	// agent suggestion: BE MsgWhatsappPing@0, BE stunMagic@4, transactionID@8:20.
	// human input:
	return [20]byte{}
}

// IsStunPacket reports whether data looks like a STUN packet (top two bits zero).
func IsStunPacket(data []byte) bool {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L188-L190
	// TODO
	// agent suggestion: len>=2 && (data[0]&0xc0)==0x00.
	// human input:
	return false
}

// StunMessageType returns the STUN message type; ok=false if data is too short.
func StunMessageType(data []byte) (uint16, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L192-L194
	// TODO
	// agent suggestion: if len<2 return 0,false; (uint16(data[0]&0x3f)<<8)|uint16(data[1]), true.
	// human input:
	return 0, false
}

// StunTransactionID returns the 12-byte transaction id; ok=false if data < 20 bytes.
func StunTransactionID(data []byte) ([]byte, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L196-L198
	// TODO
	// agent suggestion: if len<20 return nil,false; data[8:20], true.
	// human input:
	return nil, false
}

// IsAllocateOrBindingSuccess reports an Allocate/Binding success response.
func IsAllocateOrBindingSuccess(data []byte) bool {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L200-L208
	// TODO
	// agent suggestion: IsStunPacket && len>=20 && type in {AllocateSuccess, BindingSuccess}.
	// human input:
	return false
}

// IsAllocateError reports an Allocate-error response.
func IsAllocateError(data []byte) bool {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L210-L212
	// TODO
	// agent suggestion: IsStunPacket && type == MsgAllocateError.
	// human input:
	return false
}

// IsWhatsappPong reports a WhatsApp pong; a nil/empty transactionID matches any.
func IsWhatsappPong(data []byte, transactionID []byte) bool {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L214-L222
	// TODO
	// agent suggestion: must be stun + type MsgWhatsappPong; nil/empty tx -> true; else compare
	// StunTransactionID(data) to transactionID.
	// human input:
	return false
}

// StunAttribute is one parsed STUN TLV.
type StunAttribute struct {
	AttrType uint16
	Value    []byte
}

// ParseStunAttributes parses the STUN attributes after the 20-byte header.
func ParseStunAttributes(data []byte) []StunAttribute {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L231-L251
	// TODO
	// agent suggestion: guard stun+len>=20; walk off=20 while off+4<=len: type/len BE, off+=4, break
	// if off+len>len; append {type, copy(data[off:off+len])}; off += len + pad4(len).
	// human input:
	return nil
}

// ParseStunErrorCode parses the numeric error code (class*100+number); ok=false if absent.
func ParseStunErrorCode(data []byte) (uint16, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L254-L276
	// TODO
	// agent suggestion: require type in {MsgAllocateError, 0x0111}; walk attrs to end=min(20+bodyLen,
	// len); on attrErrorCode with len>=4 and off+8<=len return data[off+6]*100 + data[off+7], true.
	// human input:
	return 0, false
}

// pbTag writes a protobuf field tag.
func pbTag(out []byte, field, wire uint32) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L284-L286
	// TODO
	// agent suggestion: binary.AppendUvarint(out, uint64((field<<3)|wire)).
	// human input:
	return nil
}

// pbZigzag zigzag-encodes a signed integer.
func pbZigzag(n int64) uint64 {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L288-L290
	// TODO
	// agent suggestion: uint64((n<<1) ^ (n>>63)) (arithmetic shift).
	// human input:
	return 0
}

// pbLenDelim writes a length-delimited protobuf field.
func pbLenDelim(out []byte, field uint32, b []byte) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L292-L296
	// TODO
	// agent suggestion: out = pbTag(out, field, 2); out = binary.AppendUvarint(out, len(b)); out += b.
	// human input:
	return nil
}

// CreateVoipSenderSubscriptions builds voip.SenderSubscriptions (WASM, attr 0x4000).
func CreateVoipSenderSubscriptions(ssrc uint32) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L299-L310
	// TODO
	// agent suggestion: sender = tag(3,0)+varint(ssrc) + tag(5,0)+varint(0) + tag(6,0)+varint(0);
	// out = pbLenDelim(nil, 1, sender).
	// human input:
	return nil
}

// CreateApkSenderSubscriptions builds wa.voip.SenderSubscriptions (APK, attr 0x4025).
func CreateApkSenderSubscriptions(ssrc uint32, pid *uint32) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L313-L329
	// TODO
	// agent suggestion: ssrcLayers = tag(1,0)+varint(zigzag(ssrc)); if pid: p = tag(1,0)+varint(zigzag(pid))
	// + lenDelim(2,"audio"); ssrcLayers = lenDelim(ssrcLayers, 2, p). ext = lenDelim(nil,1,ssrcLayers);
	// out = lenDelim(nil, 1, ext).
	// human input:
	return nil
}

// CreateApkStreamDescriptors builds wa.voip.StreamDescriptors (APK, attr 0x4024).
func CreateApkStreamDescriptors(ssrc uint32) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L332-L343
	// TODO
	// agent suggestion: sd = lenDelim(1,"audio") + lenDelim(2,"OPUS") + tag(3,0)+varint(zigzag(ssrc))
	// + tag(4,0)+varint(0); out = lenDelim(nil, 1, sd).
	// human input:
	return nil
}

// BuildAndroidStunAllocateRequest builds the APK Allocate: 0x4000 token + 0x4025
// sender subs + 0x4024 stream desc + MI.
func BuildAndroidStunAllocateRequest(transactionID [12]byte, relayToken []byte, ssrc uint32, pid *uint32, integrityKey []byte, includeFingerprint bool) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/stun.rs#L346-L370
	// TODO
	// agent suggestion: attrs = stunAttr(token) || stunAttr(0x4025, CreateApkSenderSubscriptions) ||
	// stunAttr(0x4024, CreateApkStreamDescriptors); EncodeStunRequest(Allocate, tx, attrs, key, includeFingerprint).
	// human input:
	return nil
}
