package main

import (
	"cmp"
	"image/color"
	"os"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/tidwall/gjson"
)

type FontType int32

const (
	FontType_Float FontType = iota
)

var gFonts map[FontType]BitmapFont = make(map[FontType]BitmapFont)

type BitmapFont struct {
	Layers []FontLayer
}

type FontLayer struct {
	Name, ImageName       string
	Ascent, AscentPadding int32
	SpaceWidth, ZOrder    int32
	Mapping               map[byte]*CharShape
	Kerning               *map[[2]byte]int32
}

type CharShape struct {
	SourceRect rl.Rectangle
	Offset     [2]int32
	Width      int32
}

func (font BitmapFont) DrawText(text string, x, y int32, theColor color.RGBA) {
	layers := slices.Clone(font.Layers)
	slices.SortStableFunc(layers, func(a, b FontLayer) int { return cmp.Compare(a.ZOrder, b.ZOrder) })
	for i := range layers {
		cx, cy := x, y
		texture := gFontTextures[layers[i].ImageName]
		var last_char byte = 0
		for k := range text {
			if text[k] != ' ' {
				the_shape := layers[i].Mapping[text[k]]
				offset_x, offset_y := the_shape.Offset[0], the_shape.Offset[1]
				kern := layers[i].Kerning
				if kern != nil {
					if kern_val, found := (*kern)[[2]byte{last_char, text[k]}]; found {
						offset_x += kern_val
					}
				}
				rl.DrawTextureRec(texture, the_shape.SourceRect, vec2(cx+offset_x, cy-(layers[i].Ascent-the_shape.Offset[1]-offset_y)), theColor)
				cx += the_shape.Width
				cy += offset_y
			} else {
				cx += layers[i].SpaceWidth
			}
			last_char = text[k]
		}
	}
}

func (font BitmapFont) StringWidth(text string) int32 {
	var max_width int32 = 0
	for i := range font.Layers {
		var total_width int32 = 0
		var last_char byte = 0
		for k := range text {
			if text[k] != ' ' {
				the_shape := font.Layers[i].Mapping[text[k]]
				total_width += the_shape.Width
				kern := font.Layers[i].Kerning
				if kern != nil {
					if kern_val, found := (*kern)[[2]byte{last_char, text[k]}]; found {
						total_width += kern_val
					}
				}
			} else {
				total_width += font.Layers[i].SpaceWidth
			}
			last_char = text[k]
		}
		max_width = max(max_width, total_width)
	}
	return max_width
}

var gFontTextures map[string]rl.Texture2D = make(map[string]rl.Texture2D)

func DestroyFontTextures() {
	for i := range gFontTextures {
		rl.UnloadTexture(gFontTextures[i])
	}
}

func InitFonts() {
	gFonts[FontType_Float] = loadFont("./fonts/CancunFloat14.json")
}

func loadFont(filePath string) BitmapFont {
	raw, _ := os.ReadFile(filePath)
	json := string(raw)
	layers := gjson.Get(json, "Layers").Array()
	font := BitmapFont{Layers: make([]FontLayer, len(layers))}
	for k := range layers {
		obj := layers[k].Map()
		layer := FontLayer{
			Name:          obj["Name"].String(),
			ImageName:     obj["Image"].String(),
			Ascent:        int32(obj["Ascent"].Int()),
			AscentPadding: int32(obj["AscentPadding"].Int()),
			SpaceWidth:    int32(obj["SpaceWidth"].Int()),
			Mapping:       make(map[byte]*CharShape),
		}

		if _, found := gFontTextures[layer.ImageName]; !found {
			gFontTextures[layer.ImageName] = rl.LoadTexture("./fonts/" + layer.ImageName + ".png")
		}

		char_list := gjson.Get(json, obj["Chars"].String()).Array()
		cache_chars := make([]byte, len(char_list))
		for i := range char_list {
			char := char_list[i].String()[0]
			cache_chars[i] = char
			layer.Mapping[char] = new(CharShape)
		}

		char_widths := gjson.Get(json, obj["CharWidths"].String()).Array()
		for i := range char_widths {
			layer.Mapping[cache_chars[i]].Width = int32(char_widths[i].Int())
		}

		char_offsets := gjson.Get(json, obj["CharOffsets"].String()).Array()
		for i := range char_offsets {
			tmp := char_offsets[i].Array()
			layer.Mapping[cache_chars[i]].Offset = [2]int32{int32(tmp[0].Int()), int32(tmp[1].Int())}
		}

		char_srcrects := gjson.Get(json, obj["CharSrcRects"].String()).Array()
		for i := range char_srcrects {
			tmp := char_srcrects[i].Array()
			layer.Mapping[cache_chars[i]].SourceRect = rl.NewRectangle(
				float32(tmp[0].Int()), float32(tmp[1].Int()), float32(tmp[2].Int()), float32(tmp[3].Int()),
			)
		}

		if kerning, found := obj["Kerning"]; found {
			tmp_map := make(map[[2]byte]int32)
			pairs_and_values := kerning.Array()
			pairs := gjson.Get(json, pairs_and_values[0].String()).Array()
			values := gjson.Get(json, pairs_and_values[1].String()).Array()
			for i := range pairs {
				pair_str, value_int := pairs[i].String(), values[i].Int()
				tmp_map[[2]byte{pair_str[0], pair_str[1]}] = int32(value_int)
			}
			layer.Kerning = &tmp_map
		}

		font.Layers[k] = layer
	}
	return font
}
