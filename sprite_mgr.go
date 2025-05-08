package main

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type SpriteMgr struct {
	BackgroundImage      rl.Texture2D
	UpdateCnt            int32
	InSpace, SpaceScroll bool
	HoleInfos            []HoleInfo
	HoleFlashes          []HoleFlash
}

type HoleInfo struct {
	X, Y            int32
	Frame           int32
	TotalBrightness int32
	Rotation        float32
	PercentOpen     []float32
	Brightness      []int32
}

type HoleFlash struct {
	UpdateCnt, CurveIndex int32
}

func NewSpriteMgr() *SpriteMgr {
	return &SpriteMgr{
		SpaceScroll: true,
	}
}

func (mgr *SpriteMgr) AddHoleFlash(theCurveIndex, theStagger int32) {
	mgr.HoleFlashes = append(mgr.HoleFlashes, HoleFlash{-theStagger, theCurveIndex})
}

func (mgr *SpriteMgr) DrawBackground() {
	if mgr.InSpace {

	} else {
		rl.DrawTexture(mgr.BackgroundImage, 0, 0, rl.White)
	}

	for i := range mgr.HoleInfos {
		mgr.DrawHole(i, rl.White)
		total_brightness := mgr.HoleInfos[i].TotalBrightness
		if total_brightness > 0 {
			rl.BeginBlendMode(rl.BlendAdditive)
			mgr.DrawHole(i, rl.NewColor(uint8(total_brightness), uint8(total_brightness), uint8(total_brightness), 255))
			rl.EndBlendMode()
		}
	}
}

func (mgr *SpriteMgr) DrawHole(theHoleIndex int, tint color.RGBA) {
	size := gTextures[Texture_HoleCover].Width
	hole_texture := gTextures[Texture_Hole]
	hole_cover := gTextures[Texture_HoleCover]
	hole_info := &mgr.HoleInfos[theHoleIndex]
	rl.DrawTexturePro(hole_texture, rect(0, 0, hole_texture.Width, hole_texture.Height),
		rect(hole_info.X, hole_info.Y, hole_texture.Width, hole_texture.Height), vec2(hole_texture.Width/2, hole_texture.Height/2),
		-hole_info.Rotation*rl.Rad2deg, tint)
	rl.DrawTexturePro(hole_cover, rect(0, hole_info.Frame*size, size, size),
		rect(hole_info.X, hole_info.Y, size, size), vec2(size/2, size/2),
		-hole_info.Rotation*rl.Rad2deg, tint)
}

func (mgr *SpriteMgr) PlaceHole(theX, theY int32, theRotation float32) {
	rotation := float64(theRotation)
	for rotation < 0 {
		rotation += math.Pi * 2
	}
	for rotation > math.Pi*2 {
		rotation -= math.Pi * 2
	}

	if math.Abs(rotation) < 0.2 {
		rotation = 0
	}
	if math.Abs(rotation-math.Pi/2) < 0.2 {
		rotation = math.Pi / 2
	}
	if math.Abs(rotation-math.Pi) < 0.2 {
		rotation = math.Pi
	}
	if math.Abs(rotation-math.Pi*1.5) < 0.2 {
		rotation = math.Pi * 1.5
	}
	if math.Abs(rotation-math.Pi*2) < 0.2 {
		rotation = 0
	}

	var i int32 = 0
	for i := range mgr.HoleInfos {
		hole := &mgr.HoleInfos[i]
		if (hole.Y-theY)*(hole.Y-theY)+(hole.X-theX)*(hole.X-theX) < 400 {
			break
		}
	}

	if i == int32(len(mgr.HoleInfos)) {
		hole := HoleInfo{}
		hole.X, hole.Y = theX, theY
		hole.Rotation = float32(rotation)
		hole.PercentOpen = append(hole.PercentOpen, 0.0)
		hole.Brightness = append(hole.Brightness, 0)
		mgr.HoleInfos = append(mgr.HoleInfos, hole)
	}
}

func (mgr *SpriteMgr) Update() {
	mgr.UpdateCnt++
	mgr.UpdateHoles()
}

func (mgr *SpriteMgr) UpdateHole(theCurveIndex int32, thePercentOpen float32) {
	hole := &mgr.HoleInfos[theCurveIndex]
	hole.PercentOpen[theCurveIndex] = thePercentOpen

	the_max := thePercentOpen
	for i := range hole.PercentOpen {
		the_max = max(hole.PercentOpen[i], the_max)
	}

	num_rows := gTextures[Texture_HoleCover].Height / gTextures[Texture_HoleCover].Width
	hole.Frame = int32(float32(num_rows) * the_max)
	if hole.Frame >= num_rows {
		hole.Frame = num_rows - 1
	}
}

func (mgr *SpriteMgr) UpdateHoles() {
	for i := range mgr.HoleFlashes {
		flash := &mgr.HoleFlashes[i]
		flash.UpdateCnt++
		if flash.UpdateCnt < 0 {
			continue
		}

		var brightness int32 = 0
		if flash.UpdateCnt < 30 {
			brightness = 255 * flash.UpdateCnt / 30
		} else if flash.UpdateCnt < 60 {
			brightness = 255 - (flash.UpdateCnt-30)*255/30
		}

		hole := &mgr.HoleInfos[flash.CurveIndex]
		hole.Brightness[flash.CurveIndex] = brightness

		for k := range hole.Brightness {
			if brightness < hole.Brightness[k] {
				brightness = hole.Brightness[k]
			}
		}
		hole.TotalBrightness = brightness
	}
}
