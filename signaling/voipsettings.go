package signaling

import (
	"errors"

	"github.com/rs/zerolog"
)

// errNotImplemented marks a scaffolded body whose logic the human has not yet
// directed. Stubbed parsers return it so a KAT cannot fake a pass.
var errNotImplemented = errors.New("signaling: not implemented")

// VoipSettings is the codec-relevant subset of the server's <voip_settings> JSON
// blob, delivered inline on an inbound <offer> and in the <ack> of an outbound
// offer. meowcaller reads it to choose the per-call audio codec.
//
// This is meowcaller-original glue: the whatsapp-rust reference does not parse
// voip_settings to pick a codec — it steers onto RFC Opus by advertising only
// <audio rate=8000> (see BuildAccept's audio_rates). Reading use_mlow_codec_v1 is
// a meowcaller-specific lever, so the parser carries no // Source of truth: port.
type VoipSettings struct {
	// UseMlowCodecV1 mirrors encode.use_mlow_codec_v1: true => Meta's 16 kHz MLow
	// codec, false => RFC Opus. Absent => true (MLow), matching current behavior.
	UseMlowCodecV1 bool
	// FrameMs mirrors encode.frame_ms (audio frame duration in ms); 0 if absent.
	FrameMs int
	// TargetBitrate mirrors rc.target_bitrate in bits/s; 0 if absent.
	TargetBitrate int
	// Present reports whether a non-empty voip_settings blob was parsed.
	Present bool
}

// ParseVoipSettings parses the <voip_settings> JSON content into the codec-relevant
// subset. An empty blob yields the zero VoipSettings; malformed JSON is an error.
func ParseVoipSettings(raw []byte, log ...zerolog.Logger) (*VoipSettings, error) {
	// TODO
	// agent suggestion: json.Unmarshal into an intermediate struct whose `encode`
	// and `rc` objects hold the stringly-typed values as strings, then convert:
	// UseMlowCodecV1 = encode.use_mlow_codec_v1 != "false" (default true when the
	// key is absent), FrameMs/TargetBitrate via strconv.Atoi (0 on absent/parse
	// fail), Present = len(raw) > 0. Return (&VoipSettings{...}, nil).
	// human input:
	_ = pickLog(log)
	return nil, errNotImplemented
}
