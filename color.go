package color

import (
	math "github.com/gabe-lee/genmath"
)

const (
	minF       = 0
	maxF       = 1
	min8       = 0
	max8       = 3
	min16      = 0
	max16      = 15
	min32      = 0
	max32      = 255
	min64      = 0
	max64      = 65535
	floats0to1 = 1056964610
	lumaR      = 0.2126
	lumaG      = 0.7152
	lumaB      = 0.0722
	lumaCheck  = lumaR + lumaB + lumaG
	epsilon    = 0.0001

	SpectrumF  = floats0to1 * floats0to1 * floats0to1
	Spectrum64 = (max64 + 1) * (max64 + 1) * (max64 + 1)
	Spectrum32 = (max32 + 1) * (max32 + 1) * (max32 + 1)
	Spectrum16 = (max16 + 1) * (max16 + 1) * (max16 + 1)
	Spectrum8  = (max8 + 1) * (max8 + 1) * (max8 + 1)
)

type ColorFA [4]float32
type ColorF [3]float32
type Color64 uint64
type Color48 [3]uint16
type Color32 uint32
type Color24 [3]uint8
type Color16 uint16
type Color8 uint8

/******************
	COLOR_FA
*******************/

func NewColorHex(hexColor string) ColorFA {
	runes := []rune(hexColor)
	for len(runes) < 4 {
		runes = append(runes, '0')
	}
	if len(runes) == 4 {
		runes = []rune{runes[0], runes[0], runes[1], runes[1], runes[2], runes[2], runes[3], runes[3]}
	}
	for len(runes) < 8 {
		runes = append(runes, '0')
	}
	if len(runes) > 8 {
		runes = runes[0:8]
	}
	c32 := Color32(0)
	off := 28
	for _, r := range runes {
		v := rune2col[r]
		c32 |= v << off
		off -= 4
	}
	return c32.ToColorFA()
}

func NewColorHSVA(h float32, s float32, v float32, a float32) ColorFA {
	h = math.Clamp(0, h, 360)
	s = math.Clamp(0, s, 1)
	v = math.Clamp(0, v, 1)
	a = math.Clamp(0, a, 1)
	if s <= 0 {
		return ColorFA{v, v, v, a}
	}
	var r, g, b, chroma, sector, blend, foundation float32
	chroma = v * s
	foundation = v - chroma
	sector = math.Clamp(0, h/60.0, 6)
	blend = chroma * (1 - math.Abs(math.FMod(sector, 2.0)-1))
	if sector < 1 {
		r, g, b = chroma, blend, 0
	} else if sector < 2 {
		r, g, b = blend, chroma, 0
	} else if sector < 3 {
		r, g, b = 0, chroma, blend
	} else if sector < 4 {
		r, g, b = 0, blend, chroma
	} else if sector < 5 {
		r, g, b = blend, 0, chroma
	} else { // sector 6
		r, g, b = chroma, 0, blend
	}
	return ColorFA{r + foundation, g + foundation, b + foundation, a}
}

func NewColorRGBA(r float32, g float32, b float32, a float32) ColorFA {
	return ColorFA{r, g, b, a}.Clamp()
}

func (c ColorFA) RGBA() (r float32, g float32, b float32, a float32) {
	return c[0], c[1], c[2], c[3]
}

func (c ColorFA) HSVA() (h float32, s float32, v float32, a float32) {
	r, g, b, a := c.RGBA()
	var min, max, chroma, sector float32
	var maxCase, rMax, gMax, bMax byte = 0, 0, 1, 2
	if r <= g && r <= b {
		min = r
	} else if g <= r && g <= b {
		min = g
	} else if b <= r && b <= g {
		min = b
	} else {
		return 0, 0, 0, 0
	}
	if r >= g && r >= b {
		max = r
		maxCase = rMax
	} else if g >= r && g >= b {
		max = g
		maxCase = gMax
	} else if b >= r && b >= g {
		max = b
		maxCase = bMax
	} else {
		return 0, 0, 0, 0
	}
	if max <= 0 {
		return 0, 0, 0, a
	}
	v = max
	chroma = max - min
	if chroma <= 0 {
		return 0, 0, v, a
	}
	s = chroma / v
	switch maxCase {
	case rMax:
		sector = (g - b) / chroma
	case gMax:
		sector = ((b - r) / chroma) + 2.0
	case bMax:
		fallthrough
	default:
		sector = ((r - g) / chroma) + 4.0
	}
	if sector < 0 {
		sector += 6
	}
	h = math.FMod(sector*60.0, 360.0)
	return h, s, v, a
}

func (c ColorFA) Hex() string {
	runes := make([]rune, 8)
	c32 := c.ToColor32()
	off := 28
	for i := 0; i < 8; i += 1 {
		v := (c32 & (15 << off)) >> off
		r := col2rune[v]
		runes[i] = r
		off -= 4
	}
	return string(runes)
}

func (c ColorFA) Red() float32 {
	return c[0]
}
func (c ColorFA) Green() float32 {
	return c[1]
}
func (c ColorFA) Blue() float32 {
	return c[2]
}
func (c ColorFA) Alpha() float32 {
	return c[3]
}
func (c ColorFA) Hue() float32 {
	h, _, _, _ := c.HSVA()
	return h
}
func (c ColorFA) Sat() float32 {
	_, s, _, _ := c.HSVA()
	return s
}
func (c ColorFA) Val() float32 {
	_, _, v, _ := c.HSVA()
	return v
}

func (c ColorFA) SetRed(red float32) ColorFA {
	return ColorFA{red, c[1], c[2], c[3]}
}
func (c ColorFA) SetGreen(green float32) ColorFA {
	return ColorFA{c[0], green, c[2], c[3]}
}
func (c ColorFA) SetBlue(blue float32) ColorFA {
	return ColorFA{c[0], c[1], blue, c[3]}
}
func (c ColorFA) SetAlpha(alpha float32) ColorFA {
	return ColorFA{c[0], c[1], c[2], alpha}
}
func (c ColorFA) SetHue(hue float32) ColorFA {
	_, s, v, a := c.HSVA()
	return NewColorHSVA(hue, s, v, a)
}
func (c ColorFA) SetSat(sat float32) ColorFA {
	h, _, v, a := c.HSVA()
	return NewColorHSVA(h, sat, v, a)
}
func (c ColorFA) SetVal(val float32) ColorFA {
	h, s, _, a := c.HSVA()
	return NewColorHSVA(h, s, val, a)
}
func (c ColorFA) SetSatVal(sat float32, val float32) ColorFA {
	h, _, _, a := c.HSVA()
	return NewColorHSVA(h, sat, val, a)
}
func (c ColorFA) SetHueVal(hue float32, val float32) ColorFA {
	_, s, _, a := c.HSVA()
	return NewColorHSVA(hue, s, val, a)
}
func (c ColorFA) SetHueSat(hue float32, sat float32) ColorFA {
	_, _, v, a := c.HSVA()
	return NewColorHSVA(hue, sat, v, a)
}
func (c ColorFA) SetHueSatVal(hue float32, sat float32, val float32) ColorFA {
	return NewColorHSVA(hue, sat, val, c[3])
}
func (c ColorFA) Luma() float32 {
	return (c[0] * lumaR) + (c[1] * lumaG) + (c[2] * lumaB)
}
func (c ColorFA) Lighten(amount float32) ColorFA {
	if amount == 0 {
		return c
	}
	luma := c.Luma()
	if amount+luma >= 1 {
		return ColorFA{1, 1, 1, c[3]}
	}
	if amount+luma <= 0 {
		return ColorFA{0, 0, 0, c[3]}
	}
	t1 := c[0] + c[1] + c[2]
	ratioR := c[0] / t1
	ratioG := c[1] / t1
	ratioB := c[2] / t1
	epsilonR := epsilon * ratioR
	epsilonG := epsilon * ratioG
	epsilonB := epsilon * ratioB
	lightR := epsilonR * lumaR
	lightG := epsilonG * lumaG
	lightB := epsilonB * lumaB
	epsilonLight := lightR + lightG + lightB
	mult := amount / epsilonLight
	deltaR := mult * epsilonR
	deltaG := mult * epsilonG
	deltaB := mult * epsilonB
	return ColorFA{c[0] + deltaR, c[1] + deltaG, c[2] + deltaB, c[3]}.cClamp()
}
func (c ColorFA) Darken(amount float32) ColorFA {
	return c.Lighten(-amount)
}
func (c ColorFA) Illuminate(other ColorFA) ColorFA {
	oLuma := other.Luma()
	return c.Lighten(oLuma)
}
func (c ColorFA) Deluminate(other ColorFA) ColorFA {
	oLuma := other.Luma()
	return c.Lighten(-oLuma)
}
func (c ColorFA) Add(other ColorFA) ColorFA {
	return ColorFA{c[0] + other[0], c[1] + other[1], c[2] + other[2], c[3]}.cClamp()
}
func (c ColorFA) Subtract(other ColorFA) ColorFA {
	return ColorFA{c[0] - other[0], c[1] - other[1], c[2] - other[2], c[3]}.cClamp()
}
func (c ColorFA) Multiply(other ColorFA) ColorFA {
	return ColorFA{c[0] * other[0], c[1] * other[1], c[2] * other[2], c[3]}.cClamp()
}
func (c ColorFA) Dilute(other ColorFA) ColorFA {
	return ColorFA{c[0], c[1], c[2], math.Clamp(minF, c[3]*other[3], maxF)}
}
func (c ColorFA) Condense(other ColorFA) ColorFA {
	return ColorFA{c[0], c[1], c[2], math.Clamp(minF, c[3]+other[3], maxF)}
}
func (c ColorFA) Divide(other ColorFA) ColorFA {
	return ColorFA{c[0] / other[0], c[1] / other[1], c[2] / other[2], c[3]}.cClamp()
}
func (c ColorFA) Blend(ratio float32, other ColorFA) ColorFA {
	return ColorFA{lerp(c[0], other[0], ratio), lerp(c[1], other[1], ratio), lerp(c[2], other[2], ratio), c[3]}.cClamp()
}
func (c ColorFA) BlendWithAlpha(ratio float32, other ColorFA) ColorFA {
	return ColorFA{lerp(c[0], other[0], ratio), lerp(c[1], other[1], ratio), lerp(c[2], other[2], ratio), lerp(c[3], other[3], ratio)}.Clamp()
}
func (c ColorFA) AlphaAdjustedBlend(other ColorFA, blendFunc func(ColorFA) ColorFA) ColorFA {
	after := blendFunc(other)
	ratio := other[3]
	return c.Blend(ratio, after)
}
func (c ColorFA) Invert() ColorFA {
	return ColorFA{maxF - c[0], maxF - c[1], maxF - c[2], c[3]}.cClamp()
}
func (c ColorFA) Screen(other ColorFA) ColorFA {
	return c.Invert().Multiply(other.Invert()).Invert()
}
func (c ColorFA) Dodge(other ColorFA) ColorFA {
	return c.Divide(other.Invert())
}
func (c ColorFA) Burn(other ColorFA) ColorFA {
	return c.Invert().Divide(other).Invert()
}
func (c ColorFA) Overlay(other ColorFA) ColorFA {
	l := c.Luma()
	if l < 0.5 {
		return ColorFA{overlow(c[0], other[0]), overlow(c[1], other[1]), overlow(c[2], other[2]), c[3]}.cClamp()
	}
	return ColorFA{overhi(c[0], other[0]), overhi(c[1], other[1]), overhi(c[2], other[2]), c[3]}.cClamp()
}
func (c ColorFA) HardLight(other ColorFA) ColorFA {
	l := other.Luma()
	if l < 0.5 {
		return ColorFA{overlow(c[0], other[0]), overlow(c[1], other[1]), overlow(c[2], other[2]), c[3]}.cClamp()
	}
	return ColorFA{overhi(c[0], other[0]), overhi(c[1], other[1]), overhi(c[2], other[2]), c[3]}.cClamp()
}
func (c ColorFA) SoftLight(other ColorFA) ColorFA {
	return ColorFA{soft(c[0], other[0]), soft(c[1], other[1]), soft(c[2], other[2]), c[3]}.cClamp()
}
func (c ColorFA) VividLight(other ColorFA) ColorFA {
	l := other.Luma()
	if l < 0.5 {
		return c.Burn(other)
	}
	return c.Dodge(other)
}
func (c ColorFA) LightestLuma(other ColorFA) ColorFA {
	if c.Luma() > other.Luma() {
		return c
	}
	return other
}
func (c ColorFA) DarkestLuma(other ColorFA) ColorFA {
	if c.Luma() < other.Luma() {
		return c
	}
	return other
}
func (c ColorFA) LightestComponent(other ColorFA) ColorFA {
	var r, g, b float32
	if c[0]*lumaR > other[0]*lumaR {
		r = c[0]
	} else {
		r = other[0]
	}
	if c[1]*lumaG > other[1]*lumaG {
		g = c[1]
	} else {
		g = other[1]
	}
	if c[2]*lumaB > other[2]*lumaB {
		b = c[2]
	} else {
		b = other[2]
	}
	return ColorFA{r, g, b, c[3]}
}
func (c ColorFA) DarkestComponent(other ColorFA) ColorFA {
	var r, g, b float32
	if c[0]*lumaR < other[0]*lumaR {
		r = c[0]
	} else {
		r = other[0]
	}
	if c[1]*lumaG < other[1]*lumaG {
		g = c[1]
	} else {
		g = other[1]
	}
	if c[2]*lumaB < other[2]*lumaB {
		b = c[2]
	} else {
		b = other[2]
	}
	return ColorFA{r, g, b, c[3]}
}
func (c ColorFA) LargestComponent(other ColorFA) ColorFA {
	return ColorFA{math.Max(c[0], other[0]), math.Max(c[1], other[1]), math.Max(c[2], other[2]), c[3]}
}
func (c ColorFA) LargestAlpha(other ColorFA) ColorFA {
	return ColorFA{c[0], c[1], c[2], math.Max(c[3], other[3])}
}
func (c ColorFA) SmallestComponent(other ColorFA) ColorFA {
	return ColorFA{math.Min(c[0], other[0]), math.Min(c[1], other[1]), math.Min(c[2], other[2]), c[3]}
}
func (c ColorFA) SmallestAlpha(other ColorFA) ColorFA {
	return ColorFA{c[0], c[1], c[2], math.Min(c[3], other[3])}
}
func (c ColorFA) Clamp() ColorFA {
	return ColorFA{math.Clamp(minF, c[0], maxF), math.Clamp(minF, c[1], maxF), math.Clamp(minF, c[2], maxF), math.Clamp(minF, c[3], maxF)}
}
func (c ColorFA) cClamp() ColorFA {
	return ColorFA{math.Clamp(minF, c[0], maxF), math.Clamp(minF, c[1], maxF), math.Clamp(minF, c[2], maxF), c[3]}
}

func (c ColorFA) ToColorF() ColorF {
	return ColorF{c[0], c[1], c[2]}
}

func (c ColorFA) ToColor64() Color64 {
	r, g, b, a := c.RGBA()
	rr := math.RoundClamp(min64, r*max64, max64)
	gg := math.RoundClamp(min64, g*max64, max64)
	bb := math.RoundClamp(min64, b*max64, max64)
	aa := math.RoundClamp(min64, a*max64, max64)
	return Color64(rr)<<48 | Color64(gg)<<32 | Color64(bb)<<16 | Color64(aa)
}

func (c ColorFA) ToColor48() Color48 {
	return c.ToColor64().ToColor48()
}

func (c ColorFA) ToColor32() Color32 {
	r, g, b, a := c.RGBA()
	rr := math.RoundClamp(min32, r*max32, max32)
	gg := math.RoundClamp(min32, g*max32, max32)
	bb := math.RoundClamp(min32, b*max32, max32)
	aa := math.RoundClamp(min32, a*max32, max32)
	return Color32(rr)<<24 | Color32(gg)<<16 | Color32(bb)<<8 | Color32(aa)
}

func (c ColorFA) ToColor24() Color24 {
	return c.ToColor32().ToColor24()
}

func (c ColorFA) ToColor16() Color16 {
	r, g, b, a := c.RGBA()
	rr := math.RoundClamp(min16, r*max16, max16)
	gg := math.RoundClamp(min16, g*max16, max16)
	bb := math.RoundClamp(min16, b*max16, max16)
	aa := math.RoundClamp(min16, a*max16, max16)
	return Color16(rr)<<12 | Color16(gg)<<8 | Color16(bb)<<4 | Color16(aa)
}

func (c ColorFA) ToColor8() Color8 {
	r, g, b, a := c.RGBA()
	rr := math.RoundClamp(min8, r*max8, max8)
	gg := math.RoundClamp(min8, g*max8, max8)
	bb := math.RoundClamp(min8, b*max8, max8)
	aa := math.RoundClamp(min8, a*max8, max8)
	return Color8(rr)<<6 | Color8(gg)<<4 | Color8(bb)<<2 | Color8(aa)
}

/******************
	COLOR_F
*******************/

func (c ColorF) RGBA() (r float32, g float32, b float32, a float32) {
	return c[0], c[1], c[2], maxF
}

func (c ColorF) ToColorFA() ColorFA {
	return ColorFA{c[0], c[1], c[2], maxF}
}

/******************
	COLOR_64
*******************/

func (c Color64) RGBA() (r uint16, g uint16, b uint16, a uint16) {
	r = uint16((c & (max64 << 48)) >> 48)
	g = uint16((c & (max64 << 32)) >> 32)
	b = uint16((c & (max64 << 16)) >> 16)
	a = uint16(c & (max64))
	return r, g, b, a
}

func (c Color64) ToColorFA() ColorFA {
	r, g, b, a := c.RGBA()
	rr := math.Clamp(minF, float32(r)/max64, maxF)
	gg := math.Clamp(minF, float32(g)/max64, maxF)
	bb := math.Clamp(minF, float32(b)/max64, maxF)
	aa := math.Clamp(minF, float32(a)/max64, maxF)
	return ColorFA{rr, gg, bb, aa}
}

func (c Color64) ToColor48() Color48 {
	r, g, b, _ := c.RGBA()
	return Color48{r, g, b}
}

/******************
	COLOR_48
*******************/

func (c Color48) RGBA() (r uint16, g uint16, b uint16, a uint16) {
	return c[0], c[1], c[2], max64
}

func (c Color48) ToColor64() Color64 {
	return Color64(c[0])<<48 | Color64(c[1])<<32 | Color64(c[2])<<16 | max64
}

func (c Color48) ToColorFA() ColorFA {
	r := math.Clamp(minF, float32(c[0])/max64, maxF)
	g := math.Clamp(minF, float32(c[1])/max64, maxF)
	b := math.Clamp(minF, float32(c[2])/max64, maxF)
	return ColorFA{r, g, b, maxF}
}

/******************
	COLOR_32
*******************/

func (c Color32) RGBA() (r uint8, g uint8, b uint8, a uint8) {
	r = uint8((c & (max32 << 24)) >> 24)
	g = uint8((c & (max32 << 16)) >> 16)
	b = uint8((c & (max32 << 8)) >> 8)
	a = uint8(c & (max32))
	return r, g, b, a
}

func (c Color32) ToColorFA() ColorFA {
	r, g, b, a := c.RGBA()
	rr := math.Clamp(minF, float32(r)/max32, maxF)
	gg := math.Clamp(minF, float32(g)/max32, maxF)
	bb := math.Clamp(minF, float32(b)/max32, maxF)
	aa := math.Clamp(minF, float32(a)/max32, maxF)
	return ColorFA{rr, gg, bb, aa}
}

func (c Color32) ToColor24() Color24 {
	r, g, b, _ := c.RGBA()
	return Color24{r, g, b}
}

/******************
	COLOR_24
*******************/

func (c Color24) RGBA() (r uint8, g uint8, b uint8, a uint8) {
	return c[0], c[1], c[2], 255
}

func (c Color24) ToColor32() Color32 {
	return Color32(c[0])<<24 | Color32(c[1])<<16 | Color32(c[2])<<8 | max32
}

func (c Color24) ToColorFA() ColorFA {
	r := math.Clamp(minF, float32(c[0])/max32, maxF)
	g := math.Clamp(minF, float32(c[1])/max32, maxF)
	b := math.Clamp(minF, float32(c[2])/max32, maxF)
	return ColorFA{r, g, b, maxF}
}

/******************
	COLOR_16
*******************/

func (c Color16) RGBA() (r uint8, g uint8, b uint8, a uint8) {
	r = uint8((c & (max16 << 12)) >> 12)
	g = uint8((c & (max16 << 8)) >> 8)
	b = uint8((c & (max16 << 4)) >> 4)
	a = uint8(c & (max16))
	return r, g, b, a
}

func (c Color16) ToColorFA() ColorFA {
	r, g, b, a := c.RGBA()
	rr := math.Clamp(minF, float32(r)/max16, maxF)
	gg := math.Clamp(minF, float32(g)/max16, maxF)
	bb := math.Clamp(minF, float32(b)/max16, maxF)
	aa := math.Clamp(minF, float32(a)/max16, maxF)
	return ColorFA{rr, gg, bb, aa}
}

/******************
	COLOR_8
*******************/

func (c Color8) RGBA() (r uint8, g uint8, b uint8, a uint8) {
	r = uint8((c & (max8 << 6)) >> 6)
	g = uint8((c & (max8 << 4)) >> 4)
	b = uint8((c & (max8 << 2)) >> 2)
	a = uint8(c & (max8))
	return r, g, b, a
}

func (c Color8) ToColorFA() ColorFA {
	r, g, b, a := c.RGBA()
	rr := math.Clamp(minF, float32(r)/max8, maxF)
	gg := math.Clamp(minF, float32(g)/max8, maxF)
	bb := math.Clamp(minF, float32(b)/max8, maxF)
	aa := math.Clamp(minF, float32(a)/max8, maxF)
	return ColorFA{rr, gg, bb, aa}
}

/******************
	INTERNAL
*******************/

var rune2col = map[rune]Color32{
	'0': 0,
	'1': 1,
	'2': 2,
	'3': 3,
	'4': 4,
	'5': 5,
	'6': 6,
	'7': 7,
	'8': 8,
	'9': 9,
	'A': 10,
	'a': 10,
	'B': 11,
	'b': 11,
	'C': 12,
	'c': 12,
	'D': 13,
	'd': 13,
	'E': 14,
	'e': 14,
	'F': 15,
	'f': 15,
}
var col2rune = [...]rune{
	0:  '0',
	1:  '1',
	2:  '2',
	3:  '3',
	4:  '4',
	5:  '5',
	6:  '6',
	7:  '7',
	8:  '8',
	9:  '9',
	10: 'A',
	11: 'B',
	12: 'C',
	13: 'D',
	14: 'E',
	15: 'F',
}

func lerp(a float32, b float32, ratio float32) float32 {
	diff := b - a
	return a + (diff * ratio)
}

func overlow(a float32, b float32) float32 {
	return 2 * a * b
}
func overhi(a float32, b float32) float32 {
	return 1 - (2 * (1 - a) * (1 - b))
}

func soft(a float32, b float32) float32 {
	return ((1 - (2 * b)) * (a * a)) + (2 * b * a)
}
