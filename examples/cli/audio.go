package main

import (
	"encoding/binary"
	"sync"

	"github.com/gen2brain/malgo"
)

const (
	sampleRate  = 16000
	frameSamps  = 960 // 60 ms @ 16 kHz — one MLow frame
	numChannels = 1
)

// audio owns a cross-platform miniaudio context (CoreAudio / WASAPI / ALSA /
// PulseAudio, picked by the OS). It hands out a mic frame source and a speaker sink,
// both in 16 kHz mono s16 — the codec's native format — so no resampling is needed.
type audio struct {
	ctx *malgo.AllocatedContext
}

func newAudio() (*audio, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, err
	}
	return &audio{ctx: ctx}, nil
}

func (a *audio) close() {
	_ = a.ctx.Uninit()
	a.ctx.Free()
}

// openMic starts capture and returns a channel of fixed 960-sample frames plus a
// stop func. The device callback delivers arbitrary chunk sizes, so we re-window
// them into exact 60 ms frames; if the consumer falls behind, frames are dropped
// (the call keeps real-time rather than accumulating latency).
func (a *audio) openMic() (<-chan []int16, func(), error) {
	cfg := malgo.DefaultDeviceConfig(malgo.Capture)
	cfg.Capture.Format = malgo.FormatS16
	cfg.Capture.Channels = numChannels
	cfg.SampleRate = sampleRate
	cfg.Alsa.NoMMap = 1

	frames := make(chan []int16, 16)
	var acc []int16 // touched only by the (single) capture callback thread
	onData := func(_, in []byte, count uint32) {
		for i := 0; i+1 < len(in); i += 2 {
			acc = append(acc, int16(binary.LittleEndian.Uint16(in[i:])))
		}
		for len(acc) >= frameSamps {
			f := make([]int16, frameSamps)
			copy(f, acc[:frameSamps])
			acc = acc[frameSamps:]
			select {
			case frames <- f:
			default: // consumer slow: drop to stay real-time
			}
		}
	}
	dev, err := malgo.InitDevice(a.ctx.Context, cfg, malgo.DeviceCallbacks{Data: onData})
	if err != nil {
		return nil, nil, err
	}
	if err := dev.Start(); err != nil {
		dev.Uninit()
		return nil, nil, err
	}
	return frames, func() { _ = dev.Stop(); dev.Uninit() }, nil
}

// openSpeaker starts playback and returns a channel onto which decoded 16 kHz mono
// frames are pushed, plus a stop func. A small jitter buffer rides out scheduling
// gaps; underruns are zero-filled (silence) rather than glitching.
func (a *audio) openSpeaker() (chan<- []int16, func(), error) {
	cfg := malgo.DefaultDeviceConfig(malgo.Playback)
	cfg.Playback.Format = malgo.FormatS16
	cfg.Playback.Channels = numChannels
	cfg.SampleRate = sampleRate

	in := make(chan []int16, 64)
	var (
		mu  sync.Mutex
		buf []int16
	)
	done := make(chan struct{})
	go func() {
		for f := range in {
			mu.Lock()
			buf = append(buf, f...)
			mu.Unlock()
		}
		close(done)
	}()

	onData := func(out, _ []byte, count uint32) {
		need := int(count)
		mu.Lock()
		n := min(need, len(buf))
		for i := range n {
			binary.LittleEndian.PutUint16(out[i*2:], uint16(buf[i]))
		}
		buf = buf[n:]
		mu.Unlock()
		for i := n * 2; i < need*2; i++ {
			out[i] = 0 // zero-fill underrun
		}
	}
	dev, err := malgo.InitDevice(a.ctx.Context, cfg, malgo.DeviceCallbacks{Data: onData})
	if err != nil {
		return nil, nil, err
	}
	if err := dev.Start(); err != nil {
		dev.Uninit()
		return nil, nil, err
	}
	stop := func() {
		close(in)
		<-done
		_ = dev.Stop()
		dev.Uninit()
	}
	return in, stop, nil
}

// pcmToFloat converts s16 mono samples to the codec's [-1, 1) float range.
func pcmToFloat(pcm []int16) []float32 {
	out := make([]float32, len(pcm))
	for i, s := range pcm {
		out[i] = float32(s) / 32768
	}
	return out
}

// floatToPCM converts the codec's float output back to clamped s16 mono.
func floatToPCM(f []float32) []int16 {
	out := make([]int16, len(f))
	for i, s := range f {
		v := s * 32768
		switch {
		case v > 32767:
			v = 32767
		case v < -32768:
			v = -32768
		}
		out[i] = int16(v)
	}
	return out
}
