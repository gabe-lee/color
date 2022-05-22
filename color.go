package color

import (
	math "github.com/gabe-lee/genmath"
)

const (
	minFF      = 0
	maxFF      = 1
	min8       = 0
	max8       = 3
	min16      = 0
	max16      = 15
	min32      = 0
	max32      = 255
	floats0to1 = 1056964610
	lumaR      = 0.2126
	lumaG      = 0.7152
	lumaB      = 0.0722
	lumaCheck  = lumaR + lumaB + lumaG

	SpectrumFF    = floats0to1 * floats0to1 * floats0to1
	Spectrum32_24 = (max32 + 1) * (max32 + 1) * (max32 + 1)
	Spectrum16_12 = (max16 + 1) * (max16 + 1) * (max16 + 1)
	Spectrum8_6   = (max8 + 1) * (max8 + 1) * (max8 + 1)
)

type ColorFF [4]float32
type Color32 uint32
type Color24 [3]uint8
type Color16 uint16
type Color8 uint8

type IColor interface {
	RGBA() (r uint8, g uint8, b uint8, a uint8)
	ToColorFF() ColorFF
}

var _ IColor = (*Color32)(nil)
var _ IColor = (*Color16)(nil)
var _ IColor = (*Color8)(nil)

func NewColorHex(hexColor string) ColorFF {
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
	return c32.ToColorFF()
}

func NewColorHSVA(h float32, s float32, v float32, a float32) ColorFF {
	h = math.Clamp(0, h, 360)
	s = math.Clamp(0, s, 1)
	v = math.Clamp(0, v, 1)
	a = math.Clamp(0, a, 1)
	if s <= 0 {
		return ColorFF{v, v, v, a}
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
	return ColorFF{r + foundation, g + foundation, b + foundation, a}
}

func NewColorRGBA(r float32, g float32, b float32, a float32) ColorFF {
	return ColorFF{r, g, b, a}.Clamp()
}

func (c ColorFF) RGBA() (r float32, g float32, b float32, a float32) {
	return c[0], c[1], c[2], c[3]
}

func (c ColorFF) HSVA() (h float32, s float32, v float32, a float32) {
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

func (c ColorFF) Hex() string {
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

func (c ColorFF) Luma() float32 {
	return (c[0] * lumaR) + (c[1] * lumaG) + (c[2] * lumaB)
}
func (c ColorFF) Add(other ColorFF) ColorFF {
	return ColorFF{c[0] + other[0], c[1] + other[1], c[2] + other[2], c[3]}.cClamp()
}
func (c ColorFF) Subtract(other ColorFF) ColorFF {
	return ColorFF{c[0] - other[0], c[1] - other[1], c[2] - other[2], c[3]}.cClamp()
}
func (c ColorFF) Multiply(other ColorFF) ColorFF {
	return ColorFF{c[0] * other[0], c[1] * other[1], c[2] * other[2], c[3]}.cClamp()
}
func (c ColorFF) Dilute(other ColorFF) ColorFF {
	return ColorFF{c[0], c[1], c[2], math.Clamp(minFF, c[3]*other[3], maxFF)}
}
func (c ColorFF) Condense(other ColorFF) ColorFF {
	return ColorFF{c[0], c[1], c[2], math.Clamp(minFF, c[3]+other[3], maxFF)}
}
func (c ColorFF) Divide(other ColorFF) ColorFF {
	return ColorFF{c[0] / other[0], c[1] / other[1], c[2] / other[2], c[3]}.cClamp()
}
func (c ColorFF) Blend(ratio float32, other ColorFF) ColorFF {
	return ColorFF{lerp(c[0], other[0], ratio), lerp(c[1], other[1], ratio), lerp(c[2], other[2], ratio), c[3]}.cClamp()
}
func (c ColorFF) BlendWithAlpha(ratio float32, other ColorFF) ColorFF {
	return ColorFF{lerp(c[0], other[0], ratio), lerp(c[1], other[1], ratio), lerp(c[2], other[2], ratio), lerp(c[3], other[3], ratio)}.Clamp()
}
func (c ColorFF) AlphaAdjustedBlend(other ColorFF, blendFunc func(ColorFF) ColorFF) ColorFF {
	after := blendFunc(other)
	ratio := other[3]
	return c.Blend(ratio, after)
}
func (c ColorFF) Invert() ColorFF {
	return ColorFF{maxFF - c[0], maxFF - c[1], maxFF - c[2], c[3]}.cClamp()
}
func (c ColorFF) Screen(other ColorFF) ColorFF {
	return c.Invert().Multiply(other.Invert()).Invert()
}
func (c ColorFF) Dodge(other ColorFF) ColorFF {
	return c.Divide(other.Invert())
}
func (c ColorFF) Burn(other ColorFF) ColorFF {
	return c.Invert().Divide(other).Invert()
}
func (c ColorFF) Overlay(other ColorFF) ColorFF {
	l := c.Luma()
	if l < 0.5 {
		return ColorFF{overlow(c[0], other[0]), overlow(c[1], other[1]), overlow(c[2], other[2]), c[3]}.cClamp()
	}
	return ColorFF{overhi(c[0], other[0]), overhi(c[1], other[1]), overhi(c[2], other[2]), c[3]}.cClamp()
}
func (c ColorFF) HardLight(other ColorFF) ColorFF {
	l := other.Luma()
	if l < 0.5 {
		return ColorFF{overlow(c[0], other[0]), overlow(c[1], other[1]), overlow(c[2], other[2]), c[3]}.cClamp()
	}
	return ColorFF{overhi(c[0], other[0]), overhi(c[1], other[1]), overhi(c[2], other[2]), c[3]}.cClamp()
}
func (c ColorFF) SoftLight(other ColorFF) ColorFF {
	return ColorFF{soft(c[0], other[0]), soft(c[1], other[1]), soft(c[2], other[2]), c[3]}.cClamp()
}
func (c ColorFF) VividLight(other ColorFF) ColorFF {
	l := other.Luma()
	if l < 0.5 {
		return c.Burn(other)
	}
	return c.Dodge(other)
}
func (c ColorFF) Clamp() ColorFF {
	return ColorFF{math.Clamp(minFF, c[0], maxFF), math.Clamp(minFF, c[1], maxFF), math.Clamp(minFF, c[2], maxFF), math.Clamp(minFF, c[3], maxFF)}
}
func (c ColorFF) cClamp() ColorFF {
	return ColorFF{math.Clamp(minFF, c[0], maxFF), math.Clamp(minFF, c[1], maxFF), math.Clamp(minFF, c[2], maxFF), c[3]}
}

func (c ColorFF) ToColor32() Color32 {
	r, g, b, a := c.RGBA()
	rr := math.RoundClamp(min32, r*max32, max32)
	gg := math.RoundClamp(min32, g*max32, max32)
	bb := math.RoundClamp(min32, b*max32, max32)
	aa := math.RoundClamp(min32, a*max32, max32)
	return Color32(rr)<<24 | Color32(gg)<<16 | Color32(bb)<<8 | Color32(aa)
}

func (c ColorFF) ToColor24() Color24 {
	return c.ToColor32().ToColor24()
}

func (c ColorFF) ToColor16() Color16 {
	r, g, b, a := c.RGBA()
	rr := math.RoundClamp(min16, r*max16, max16)
	gg := math.RoundClamp(min16, g*max16, max16)
	bb := math.RoundClamp(min16, b*max16, max16)
	aa := math.RoundClamp(min16, a*max16, max16)
	return Color16(rr)<<12 | Color16(gg)<<8 | Color16(bb)<<4 | Color16(aa)
}

func (c ColorFF) ToColor8() Color8 {
	r, g, b, a := c.RGBA()
	rr := math.RoundClamp(min8, r*max8, max8)
	gg := math.RoundClamp(min8, g*max8, max8)
	bb := math.RoundClamp(min8, b*max8, max8)
	aa := math.RoundClamp(min8, a*max8, max8)
	return Color8(rr)<<6 | Color8(gg)<<4 | Color8(bb)<<2 | Color8(aa)
}

func (c Color32) RGBA() (r uint8, g uint8, b uint8, a uint8) {
	r = uint8((c & (max32 << 24)) >> 24)
	g = uint8((c & (max32 << 16)) >> 16)
	b = uint8((c & (max32 << 8)) >> 8)
	a = uint8(c & (max32))
	return r, g, b, a
}

func (c Color32) ToColorFF() ColorFF {
	r, g, b, a := c.RGBA()
	rr := math.Clamp(minFF, float32(r)/max32, maxFF)
	gg := math.Clamp(minFF, float32(g)/max32, maxFF)
	bb := math.Clamp(minFF, float32(b)/max32, maxFF)
	aa := math.Clamp(minFF, float32(a)/max32, maxFF)
	return ColorFF{rr, gg, bb, aa}
}

func (c Color32) ToColor24() Color24 {
	r, g, b, _ := c.RGBA()
	return Color24{r, g, b}
}

func (c Color16) RGBA() (r uint8, g uint8, b uint8, a uint8) {
	r = uint8((c & (max16 << 12)) >> 12)
	g = uint8((c & (max16 << 8)) >> 8)
	b = uint8((c & (max16 << 4)) >> 4)
	a = uint8(c & (max16))
	return r, g, b, a
}

func (c Color16) ToColorFF() ColorFF {
	r, g, b, a := c.RGBA()
	rr := math.Clamp(minFF, float32(r)/max16, maxFF)
	gg := math.Clamp(minFF, float32(g)/max16, maxFF)
	bb := math.Clamp(minFF, float32(b)/max16, maxFF)
	aa := math.Clamp(minFF, float32(a)/max16, maxFF)
	return ColorFF{rr, gg, bb, aa}
}

func (c Color8) RGBA() (r uint8, g uint8, b uint8, a uint8) {
	r = uint8((c & (max8 << 6)) >> 6)
	g = uint8((c & (max8 << 4)) >> 4)
	b = uint8((c & (max8 << 2)) >> 2)
	a = uint8(c & (max8))
	return r, g, b, a
}

func (c Color8) ToColorFF() ColorFF {
	r, g, b, a := c.RGBA()
	rr := math.Clamp(minFF, float32(r)/max8, maxFF)
	gg := math.Clamp(minFF, float32(g)/max8, maxFF)
	bb := math.Clamp(minFF, float32(b)/max8, maxFF)
	aa := math.Clamp(minFF, float32(a)/max8, maxFF)
	return ColorFF{rr, gg, bb, aa}
}

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
