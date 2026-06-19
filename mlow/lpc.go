package mlow

import "math"

const (
	SmplLPCOrder  = 16
	SmplLPCBufLen = 448
	SmplLPCNFFT   = 512
	SmplFLen      = SmplLPCNFFT/2 + 1
)

// smplPI is the truncated literal the reference uses (not math.Pi) — load-bearing
// for bit-faithful window/NLSF math.
//
// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/674e85164b35ca19115dfebcf605708d15951ee7/wacore/src/voip/mlow/smpl_lpc.rs#L25
const smplPI = 3.1415926535897

const (
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/674e85164b35ca19115dfebcf605708d15951ee7/wacore/src/voip/mlow/smpl_lpc.rs#L411-L414
	lsfCosTabSzFix         = 128
	binDivStepsA2NLSFFix   = 3
	maxIterationsA2NLSFFix = 16
	silkInt16Max           = 32767
)

// smplWindowLPC20 applies the 20 ms LPC analysis window to a raw analysis buffer,
// producing the windowed buffer the autocorrelation FFT consumes. useLongWin
// selects the 64-tap vs 32-tap trailing cosine taper.
func smplWindowLPC20(input *[SmplLPCBufLen]float32, useLongWin bool) [SmplLPCBufLen]float32 {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/674e85164b35ca19115dfebcf605708d15951ee7/wacore/src/voip/mlow/smpl_lpc.rs#L55-L90
	// TODO
	// agent suggestion: gen_sin_win(264) over the head, copy the 120-sample middle
	// verbatim, gen_cos_win taper over the trailing 64 (or 32, zeroing the last 32
	// for the short window). Trig in f64 cast to f32, using the truncated SMPL_PI
	// literal 3.1415926535897, not math.Pi.
	// human input:
	return [SmplLPCBufLen]float32{}
}

// smplLPCAnalyzeWithF2 runs the full LPC analysis over a windowed buffer: returns
// the post-bandwidth-expansion monic LPC A[0..16] (A[0]=1) and the power spectrum
// F2[0..256] that the pitch and signal-mode paths consume.
func smplLPCAnalyzeWithF2(windowed *[SmplLPCBufLen]float32) ([SmplLPCOrder + 1]float32, [SmplFLen]float32) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/674e85164b35ca19115dfebcf605708d15951ee7/wacore/src/voip/mlow/smpl_lpc.rs#L255-L283
	// TODO
	// agent suggestion: zero-pad to 512, forward real FFT, power spectrum F2, cast
	// to f64, brute_dct → R[0..16], ac2rc_dbl (Schur, reg=5e-7) → rc, rc2a → monic
	// A, bwe_expand (0.9999^i). BLOCKED: needs a 512-pt real FFT (rfft_forward_
	// ordered) that has no module/datasheet yet — see chat.
	// human input:
	return [SmplLPCOrder + 1]float32{}, [SmplFLen]float32{}
}

// smplLPCInterpol returns the per-subframe interpolated LPC predictor coefficients
// (interpolation index 0) and the carried last-subframe NLSF. nlsf2a is the
// decoder's NLSF→A conversion, supplied by the caller.
func smplLPCInterpol(
	lsf, prevLSF []float32,
	nlsf2a func(nlsf []float32) []float32,
) (predcoefs [4][SmplLPCOrder + 1]float32, ilsf [SmplLPCOrder]float32) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/674e85164b35ca19115dfebcf605708d15951ee7/wacore/src/voip/mlow/smpl_lpc.rs#L358-L367
	// TODO
	// agent suggestion: delegate to smplLPCInterpolIdx with interpolIdx=0.
	// human input:
	return predcoefs, ilsf
}

// smplLPCInterpolIdx is smplLPCInterpol for an explicit interpolation-weight row.
func smplLPCInterpolIdx(
	lsf, prevLSF []float32,
	interpolIdx int,
	nlsf2a func(nlsf []float32) []float32,
) (predcoefs [4][SmplLPCOrder + 1]float32, ilsf [SmplLPCOrder]float32) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/674e85164b35ca19115dfebcf605708d15951ee7/wacore/src/voip/mlow/smpl_lpc.rs#L370-L407
	// TODO
	// agent suggestion: pick the interp row (clamp idx to 1); seed prev from prevLSF
	// when its last entry is non-zero else from lsf; per subframe interpolate
	// (1-w)*prev + w*lsf (or copy lsf when w==1), nlsf2a → A, force A[0]=1, then
	// lpc_stabilize via repeated bandwidth expansion until stable.
	// human input:
	return predcoefs, ilsf
}

func silkRshiftRound(a, shift int32) int32 {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/674e85164b35ca19115dfebcf605708d15951ee7/wacore/src/voip/mlow/smpl_lpc.rs#L419-L425
	if shift == 1 {
		return (a >> 1) + (a & 1)
	}
	return ((a >> (shift - 1)) + 1) >> 1
}

func silkSmlaww(a32, b32, c32 int32) int32 {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/674e85164b35ca19115dfebcf605708d15951ee7/wacore/src/voip/mlow/smpl_lpc.rs#L428-L434
	return int32(int64(a32) + ((int64(b32) * int64(c32)) >> 16))
}

func silkDiv32(a, b int32) int32 {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/674e85164b35ca19115dfebcf605708d15951ee7/wacore/src/voip/mlow/smpl_lpc.rs#L437-L439
	return a / b
}

// silkBwexpander32 chirp-expands the Q16 LPC coefficients in place.
func silkBwexpander32(ar []int32, d int, chirpQ16 int32) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/674e85164b35ca19115dfebcf605708d15951ee7/wacore/src/voip/mlow/smpl_lpc.rs#L448-L457
	chirp := chirpQ16
	chirpMinusOne := chirpQ16 - 65536
	for i := 0; i < d-1; i++ {
		ar[i] = int32((int64(chirp) * int64(ar[i])) >> 16)
		mul := chirp * chirpMinusOne
		chirp += silkRshiftRound(mul, 16)
	}
	ar[d-1] = int32((int64(chirp) * int64(ar[d-1])) >> 16)
}

func silkA2NLSFTransPoly(p []int32, dd int) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/674e85164b35ca19115dfebcf605708d15951ee7/wacore/src/voip/mlow/smpl_lpc.rs#L459-L466
	for k := 2; k <= dd; k++ {
		for n := dd; n >= k+1; n-- {
			p[n-2] -= p[n]
		}
		p[k-2] -= p[k] << 1
	}
}

func silkA2NLSFEvalPoly(p []int32, x int32, dd int) int32 {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/674e85164b35ca19115dfebcf605708d15951ee7/wacore/src/voip/mlow/smpl_lpc.rs#L468-L475
	xQ16 := x << 4
	y32 := p[dd]
	for n := dd - 1; n >= 0; n-- {
		y32 = silkSmlaww(p[n], y32, xQ16)
	}
	return y32
}

func silkA2NLSFInit(aQ16, p, q []int32, dd int) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/674e85164b35ca19115dfebcf605708d15951ee7/wacore/src/voip/mlow/smpl_lpc.rs#L477-L490
	p[dd] = 1 << 16
	q[dd] = 1 << 16
	for k := 0; k < dd; k++ {
		p[k] = -aQ16[dd-k-1] - aQ16[dd+k]
		q[k] = -aQ16[dd-k-1] + aQ16[dd+k]
	}
	for k := dd; k >= 1; k-- {
		p[k-1] -= p[k]
		q[k-1] += q[k]
	}
	silkA2NLSFTransPoly(p, dd)
	silkA2NLSFTransPoly(q, dd)
}

// silkA2NLSF converts monic whitening coefficients (Q16) to NLSF (Q15). It mutates
// aQ16 (bandwidth expansion on non-convergence). d is the even filter order.
func silkA2NLSF(nlsf, aQ16 []int32, d int) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/674e85164b35ca19115dfebcf605708d15951ee7/wacore/src/voip/mlow/smpl_lpc.rs#L494-L589
	dd := d >> 1
	p := make([]int32, dd+1)
	q := make([]int32, dd+1)
	silkA2NLSFInit(aQ16, p, q, dd)

	useQ := false
	poly := func() []int32 {
		if useQ {
			return q
		}
		return p
	}
	xlo := silkLSFCosTabFIXQ12[0]
	ylo := silkA2NLSFEvalPoly(poly(), xlo, dd)

	var rootIx int
	if ylo < 0 {
		nlsf[0] = 0
		useQ = true
		ylo = silkA2NLSFEvalPoly(q, xlo, dd)
		rootIx = 1
	}
	k := 1
	var iter, thr int32
	for {
		xhi := silkLSFCosTabFIXQ12[k]
		yhi := silkA2NLSFEvalPoly(poly(), xhi, dd)

		if (ylo <= 0 && yhi >= thr) || (ylo >= 0 && yhi <= -thr) {
			if yhi == 0 {
				thr = 1
			} else {
				thr = 0
			}
			xloL, yloL, xhiL := xlo, ylo, xhi
			ffrac := int32(-256)
			for m := int32(0); m < binDivStepsA2NLSFFix; m++ {
				xmid := silkRshiftRound(xloL+xhiL, 1)
				ymid := silkA2NLSFEvalPoly(poly(), xmid, dd)
				if (yloL <= 0 && ymid >= 0) || (yloL >= 0 && ymid <= 0) {
					xhiL = xmid
					yhi = ymid
				} else {
					xloL = xmid
					yloL = ymid
					ffrac += 128 >> m
				}
			}
			absYloL := yloL
			if absYloL < 0 {
				absYloL = -absYloL
			}
			if absYloL < 65536 {
				den := yloL - yhi
				nom := (yloL << (8 - binDivStepsA2NLSFFix)) + (den >> 1)
				if den != 0 {
					ffrac += silkDiv32(nom, den)
				}
			} else {
				ffrac += silkDiv32(yloL, (yloL-yhi)>>(8-binDivStepsA2NLSFFix))
			}
			nlsf[rootIx] = min((int32(k)<<8)+ffrac, silkInt16Max)

			rootIx++
			if rootIx >= d {
				break
			}
			useQ = rootIx&1 != 0
			xlo = silkLSFCosTabFIXQ12[k-1]
			ylo = (1 - (int32(rootIx) & 2)) << 12
		} else {
			k++
			xlo = xhi
			ylo = yhi
			thr = 0
			if k > lsfCosTabSzFix {
				iter++
				if iter > maxIterationsA2NLSFFix {
					nlsf[0] = silkDiv32(1<<15, int32(d)+1)
					for kk := 1; kk < d; kk++ {
						nlsf[kk] = nlsf[kk-1] + nlsf[0]
					}
					return
				}
				silkBwexpander32(aQ16, d, int32(65536-(1<<iter)))
				silkA2NLSFInit(aQ16, p, q, dd)
				useQ = false
				xlo = silkLSFCosTabFIXQ12[0]
				ylo = silkA2NLSFEvalPoly(p, xlo, dd)
				if ylo < 0 {
					nlsf[0] = 0
					useQ = true
					ylo = silkA2NLSFEvalPoly(q, xlo, dd)
					rootIx = 1
				} else {
					rootIx = 0
				}
				k = 1
			}
		}
	}
}

// smplA2NLSF16 converts post-BWE float LPC A[0..16] (A[0]=1) into the analysis
// NLSF in radians (0..pi) via the fixed-point silk forward A→NLSF.
func smplA2NLSF16(a []float32) [SmplLPCOrder]float32 {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/674e85164b35ca19115dfebcf605708d15951ee7/wacore/src/voip/mlow/smpl_lpc.rs#L592-L604
	var aQ16 [SmplLPCOrder]int32
	for i := range SmplLPCOrder {
		aQ16[i] = int32(math.Round(float64(-a[i+1] * 65536.0)))
	}
	var lsfQ15 [SmplLPCOrder]int32
	silkA2NLSF(lsfQ15[:], aQ16[:], SmplLPCOrder)
	var nlsf [SmplLPCOrder]float32
	for i := range SmplLPCOrder {
		nlsf[i] = float32(lsfQ15[i]) / 32768.0 * smplPI
	}
	return nlsf
}
