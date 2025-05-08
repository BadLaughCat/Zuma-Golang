package main

import (
	"image/color"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type ParticleMgr struct {
	Board            *Board
	HadUpdate        bool
	SparkleList      [MaxPriority + 1][]Sparkle
	ExplosionList    []Explosion
	FloatingTextList []FloatingText
}

type Sparkle struct {
	X, Y, VX, VY               float32
	Color                      color.RGBA
	Frame, Duration, UpdateCnt int32
}

type Explosion struct {
	X, Y,
	Radius, CurRadius int32
	Color, CurColor color.RGBA
	UpdateCnt       int32
}

type FloatingText struct {
	Text                          string
	Font                          FontType
	X, Y                          int32
	Color                         color.RGBA
	ScoreInc, Duration, UpdateCnt int32
	Fade                          bool
}

func (mgr *ParticleMgr) AddExplosion(x, y, theRadius int32, theColor color.RGBA, theStagger int32) {
	mgr.ExplosionList = append(mgr.ExplosionList, Explosion{
		x, y, theRadius, 0, theColor, color.RGBA{0, 0, 0, 0}, -theStagger,
	})
}

func (mgr *ParticleMgr) AddFloatingText(x, y int32, theColor color.RGBA, theText string, theFont FontType, theStagger, theScoreInc, theDuration int32, fade bool) {
	mgr.FloatingTextList = append(mgr.FloatingTextList, FloatingText{
		theText, theFont, x, y, theColor, theScoreInc, theDuration, -theStagger, fade,
	})
}

func (mgr *ParticleMgr) AddSparkle(x, y, vx, vy float32, thePriority, theDuration, theStagger int32, theColor color.RGBA) {
	sparkle := Sparkle{}
	cols := gTextures[Texture_Sparkle].Width / gTextures[Texture_Sparkle].Height
	sparkle.X = x - float32((gTextures[Texture_Sparkle].Width/cols)/2)
	sparkle.Y = y - float32(gTextures[Texture_Sparkle].Height/2)
	sparkle.VX, sparkle.VY = vx, vy
	if theDuration > 0 {
		sparkle.Duration = theDuration
	} else {
		sparkle.Duration = 2 * cols
	}

	sparkle.Frame = 0
	sparkle.UpdateCnt = -theStagger
	sparkle.Color = theColor
	mgr.SparkleList[thePriority] = append(mgr.SparkleList[thePriority], sparkle)
}

func AddTextsToMgr(texts []string, theFont FontType, theMgr *ParticleMgr, theColor color.RGBA, x, y, theStagger, theScoreInc int32) {
	if len(texts) == 0 {
		return
	}

	font := gFonts[theFont]
	texture := gFontTextures[font.Layers[0].ImageName]
	var total_width, total_height int32 = 0, 0
	for i := range texts {
		text_width := font.StringWidth(texts[i])
		if text_width >= total_width {
			total_width = text_width
		}
		total_height += texture.Height
	}

	text_x, text_y := x-total_width/2, y+total_height/2
	if total_width+text_x > GameWidth-40 {
		text_x = GameWidth - 40 - total_width
	}
	if total_height-text_y > GameHeight-40 {
		text_y = GameHeight - 40 - total_height
	}
	if text_y < 100 {
		text_y = 100
	}
	if text_x < 40 {
		text_x = 40
	}
	for i := range texts {
		text_width := font.StringWidth(texts[i])
		theMgr.AddFloatingText(text_x+(total_width-text_width)/2, text_y, theColor, texts[i], theFont, theStagger, theScoreInc, 100, false)
		text_y += texture.Height
	}
}

func (mgr *ParticleMgr) Draw(thePriority int32) {
	mgr.DrawSparkles(thePriority)
}

func (mgr *ParticleMgr) DrawExplosions() {
	for i := range mgr.ExplosionList {
		target := &mgr.ExplosionList[i]
		if target.UpdateCnt < 0 {
			continue
		}
		if target.Radius > 0 {
			rl.DrawCircle(target.X, target.Y, float32(target.CurRadius), target.CurColor)
		} else {
			width, height := gTextures[Texture_Explosion].Width, gTextures[Texture_Explosion].Height/17
			rl.BeginBlendMode(rl.BlendAdditive)
			rl.DrawTextureRec(gTextures[Texture_Explosion], rect(0, (target.UpdateCnt>>2)*height, width, height),
				vec2(target.X-width/2, target.Y-height/2), rl.White)
			rl.EndBlendMode()
		}
	}
}

func (mgr *ParticleMgr) DrawFloatingText() {
	for i := range mgr.FloatingTextList {
		target := &mgr.FloatingTextList[i]
		if target.UpdateCnt < 0 {
			continue
		}
		the_color := target.Color
		if target.Fade {
			the_color.A = uint8(255 - 255*target.UpdateCnt/target.Duration)
		}
		gFonts[target.Font].DrawText(target.Text, target.X, target.Y, the_color)
	}
}

func (mgr *ParticleMgr) DrawSparkles(thePriority int32) {
	rl.BeginBlendMode(rl.BlendAdditive)
	list := &mgr.SparkleList[thePriority]
	for i := range *list {
		if (*list)[i].UpdateCnt < 0 {
			continue
		}

		texture := gTextures[Texture_Sparkle]
		cols := texture.Width / texture.Height
		rl.DrawTextureRec(texture,
			rl.NewRectangle(float32((*list)[i].Frame*texture.Width/cols), 0, float32(texture.Width/cols), float32(texture.Height)),
			rl.NewVector2((*list)[i].X, (*list)[i].Y), (*list)[i].Color)
	}
	rl.EndBlendMode()
}

func (mgr *ParticleMgr) DrawTopMost() {
	mgr.DrawSparkles(MaxPriority)
	mgr.DrawExplosions()
	mgr.DrawFloatingText()
}

func (mgr *ParticleMgr) Update() bool {
	mgr.HadUpdate = false
	mgr.UpdateSparkles()
	mgr.UpdateExplosions()
	mgr.UpdateFloatingText()
	return mgr.HadUpdate
}

func (mgr *ParticleMgr) UpdateExplosions() {
	for i := 0; i < len(mgr.ExplosionList); {
		target := &mgr.ExplosionList[i]
		target.UpdateCnt++
		if target.UpdateCnt < 0 {
			i++
			continue
		}
		if target.UpdateCnt > 50 {
			mgr.HadUpdate = true
			mgr.ExplosionList = slices.Delete(mgr.ExplosionList, i, i+1)
			continue
		}

		mgr.HadUpdate = true
		target.CurRadius = target.Radius
		alpha := 200 - 4*target.UpdateCnt
		cur_color := target.Color
		if alpha > 0 {
			cur_color.A = uint8(alpha)
		} else {
			cur_color = color.RGBA{0, 0, 0, 1}
		}
		target.CurColor = cur_color
		i++
	}
}

func (mgr *ParticleMgr) UpdateFloatingText() {
	for i := 0; i < len(mgr.FloatingTextList); {
		target := &mgr.FloatingTextList[i]
		target.UpdateCnt++
		if target.UpdateCnt < 0 {
			i++
			continue
		}
		if target.UpdateCnt == 1 {
			if target.ScoreInc > 0 {
				mgr.Board.IncScore(target.ScoreInc, true)
			}
		}
		if target.UpdateCnt > target.Duration {
			mgr.HadUpdate = true
			mgr.FloatingTextList = slices.Delete(mgr.FloatingTextList, i, i+1)
			i++
			continue
		}
		mgr.HadUpdate = true
		target.Y--
		i++
	}
}

func (mgr *ParticleMgr) UpdateSparkles() {
	for i := range MaxPriority + 1 {
		list := &mgr.SparkleList[i]
		for k := 0; k < len(*list); {
			(*list)[k].UpdateCnt++
			if (*list)[k].UpdateCnt < 0 {
				k++
				continue
			}

			mgr.HadUpdate = true
			cols := gTextures[Texture_Sparkle].Width / gTextures[Texture_Sparkle].Height
			(*list)[k].Frame = ((*list)[k].UpdateCnt >> 1) % cols
			if (*list)[k].UpdateCnt >= (*list)[k].Duration {
				*list = slices.Delete(*list, k, k+1)
				continue
			} else {
				(*list)[k].X += (*list)[k].VX
				(*list)[k].Y += (*list)[k].VY
				k++
			}
		}
	}
}
