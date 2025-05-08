package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Frog struct {
	Angle              float32
	CenterX, CenterY   int32
	RecoilCount        int32
	RecoilX1, RecoilY1 int32
	RecoilX2, RecoilY2 int32
	Width, Height      int32
	Bullet, NextBullet *Bullet
	State              FrogState
	StatePercent       float32
	BlinkCount         int32
	Wink               bool
	FireVel            float32
	ShowNextBall       bool
}

type FrogState int32

const (
	FROGSTATE_NORMAL FrogState = iota
	FROGSTATE_FIRING
	FROGSTATE_RELOADING
)

func NewFrog() *Frog {
	return &Frog{
		Angle:        0.0,
		CenterX:      327,
		CenterY:      233,
		RecoilCount:  0,
		RecoilX1:     327,
		RecoilY1:     233,
		RecoilX2:     0,
		RecoilY2:     0,
		Width:        gTextures[Texture_FrogBase].Width,
		Height:       gTextures[Texture_FrogBase].Height,
		Bullet:       nil,
		NextBullet:   nil,
		State:        FROGSTATE_NORMAL,
		StatePercent: 0.0,
		BlinkCount:   0,
		Wink:         false,
		FireVel:      6.0,
		ShowNextBall: true,
	}
}

func (frog *Frog) CalcAngle() {
	if frog.Bullet == nil {
		return
	}

	start := float32(frog.CenterY - 20)
	end := float32(frog.CenterY + 25)
	point_x := float32(frog.CenterX + 1)
	var point_y float32 = 0.0
	if frog.State == FROGSTATE_NORMAL {
		point_y = end
	} else if frog.State == FROGSTATE_RELOADING {
		point_y = start + (end-start)*frog.StatePercent
	} else if frog.StatePercent <= 0.6 {
		point_y = end + (float32(frog.CenterY+10)-end)*frog.StatePercent/0.6
	} else {
		return
	}

	RotateXY(&point_x, &point_y, float64(frog.CenterX), float64(frog.CenterY), float64(frog.Angle))
	frog.Bullet.X = float32(point_x)
	frog.Bullet.Y = float32(point_y)
	frog.Bullet.SetRotation(frog.Angle, true)
}

func (frog *Frog) DoBlink(wink bool) {
	frog.Wink = wink
	frog.BlinkCount = 25
}

func (frog *Frog) Draw() {
	degree := -frog.Angle * rl.Rad2deg
	rl.DrawTexturePro(
		gTextures[Texture_FrogBase],
		rl.NewRectangle(0, 0, float32(frog.Width), float32(frog.Height)),
		rl.NewRectangle(float32(frog.CenterX), float32(frog.CenterY), float32(frog.Width), float32(frog.Height)),
		rl.NewVector2(float32(frog.Width/2), float32(frog.Height/2)),
		degree, rl.White,
	)

	var offset float32
	switch frog.State {
	case FROGSTATE_NORMAL:
		offset = 51
	case FROGSTATE_FIRING:
		offset = frog.StatePercent*30 + (1-frog.StatePercent)*51
	case FROGSTATE_RELOADING:
		offset = frog.StatePercent*51 + (1-frog.StatePercent)*30
	}
	var tongue_center float32 = 17
	if frog.Bullet != nil {
		tongue_center = 54 - offset
	} else {
		offset = 37
	}
	rl.DrawTexturePro(
		gTextures[Texture_FrogTongue],
		rl.NewRectangle(0, 0, float32(gTextures[Texture_FrogTongue].Width), float32(gTextures[Texture_FrogTongue].Height)),
		rl.NewRectangle(float32(frog.CenterX), float32(frog.CenterY), float32(gTextures[Texture_FrogTongue].Width), float32(gTextures[Texture_FrogTongue].Height)),
		rl.NewVector2(16, tongue_center), degree, rl.White,
	)

	if frog.Bullet != nil {
		frog.Bullet.Draw()
	}

	if frog.ShowNextBall {
		if frog.NextBullet != nil && frog.State != FROGSTATE_RELOADING {
			rl.DrawTexturePro(
				gTextures[Texture_BallDots], rl.NewRectangle(float32(frog.NextBullet.Type*15), 0, 15, 15),
				rl.NewRectangle(float32(frog.CenterX), float32(frog.CenterY), 15, 15), rl.NewVector2(7.5, 32),
				degree, rl.White,
			)
		}
	}

	rl.DrawTexturePro(
		gTextures[Texture_FrogImageMask], rl.NewRectangle(0, 0, float32(frog.Width), float32(frog.Height)),
		rl.NewRectangle(float32(frog.CenterX), float32(frog.CenterY), float32(frog.Width), float32(frog.Height)),
		rl.NewVector2(float32(frog.Width/2), float32(frog.Height/2)),
		degree, rl.White,
	)

	if frog.BlinkCount == 0 {
		return
	}
	var blink int32 = 0
	if frog.BlinkCount > 4 && frog.BlinkCount < 20 {
		blink = 1
	} else if frog.BlinkCount > 24 {
		return
	}
	draw_width, draw_height := gTextures[Texture_FrogEye].Width, gTextures[Texture_FrogEye].Height/2
	if frog.Wink {
		draw_width /= 2
	}
	rl.DrawTexturePro(
		gTextures[Texture_FrogEye],
		rl.NewRectangle(0, float32(blink*draw_height), float32(draw_width), float32(draw_height)),
		rl.NewRectangle(float32(frog.CenterX), float32(frog.CenterY), float32(draw_width), float32(draw_height)),
		rl.NewVector2(float32(gTextures[Texture_FrogEye].Width/2), 12),
		degree, rl.White,
	)
}

func (frog *Frog) EmptyBullets() {
	frog.State = FROGSTATE_NORMAL
	frog.NextBullet = nil
	frog.Bullet = nil
}

func (frog *Frog) GetFiredBullet() *Bullet {
	if frog.State == FROGSTATE_FIRING && frog.StatePercent >= 1 {
		bullet := frog.Bullet
		frog.Bullet = nil
		frog.State = FROGSTATE_NORMAL
		return bullet
	}
	return nil
}

func (frog *Frog) NeedsReload() bool {
	return frog.NextBullet == nil || frog.Bullet == nil
}

func (frog *Frog) Reload(theType int32, delay bool, thePower PowerType) {
	bullet := NewBullet()
	bullet.CurCurvePoint = make([]int32, len(globalBoard.CurveList))
	bullet.Type = theType
	bullet.SetPowerType(thePower, false)

	frog.StatePercent = 0
	frog.Bullet = nil
	frog.Bullet = frog.NextBullet
	frog.NextBullet = bullet
	frog.State = FROGSTATE_RELOADING

	if !delay {
		frog.State = FROGSTATE_NORMAL
		frog.StatePercent = 1
	}
	frog.CalcAngle()
}

func RotateXY(x, y *float32, cx, cy, rad float64) {
	ox, oy := float64(*x)-cx, float64(*y)-cy
	*x = float32(cx + ox*math.Cos(rad) + oy*math.Sin(rad))
	*y = float32(cy + oy*math.Cos(rad) - ox*math.Sin(rad))
}

func (frog *Frog) SetAngle(theAngle float32) {
	frog.Angle = theAngle
	frog.CalcAngle()
}

func (frog *Frog) SetPos(theX, theY int32) {
	frog.CenterX, frog.CenterY = theX, theY
	frog.RecoilX1, frog.RecoilY1 = theX, theY
	frog.CalcAngle()
}

func (frog *Frog) StartFire(recoil bool) bool {
	if frog.State != FROGSTATE_NORMAL || frog.Bullet == nil {
		return false
	}

	frog.StatePercent = 0
	frog.State = FROGSTATE_FIRING
	frog.CenterX = frog.RecoilX1
	frog.CenterY = frog.RecoilY1

	bullet := frog.Bullet
	rad := frog.Angle - math.Pi/2
	vx, vy := float32(math.Cos(float64(rad))), -float32(math.Sin(float64(rad)))

	bullet.VelX = vx * frog.FireVel
	bullet.VelY = vy * frog.FireVel
	bullet.X = -40*vx + bullet.X
	bullet.Y = -40*vy + bullet.Y

	frog.RecoilX1 = frog.CenterX
	frog.RecoilY1 = frog.CenterY
	frog.RecoilX2 = int32(float32(frog.CenterX) - vx*6)
	frog.RecoilY2 = int32(float32(frog.CenterY) - vy*6)
	if recoil {
		frog.RecoilCount = 25
	}
	frog.BlinkCount = 25
	frog.CalcAngle()
	return true
}

func (frog *Frog) SwapBullets(playSound bool) {
	if frog.State != FROGSTATE_NORMAL {
		return
	}
	if frog.Bullet == nil || frog.NextBullet == nil {
		return
	}
	if frog.Bullet.Type == frog.NextBullet.Type {
		return
	}
	if playSound {
		rl.PlaySound(gSounds[Sound_FrogSwap])
	}
	bullet := frog.Bullet
	frog.Bullet = frog.NextBullet
	frog.NextBullet = bullet
	frog.CalcAngle()
}

func (frog *Frog) Update() {
	if frog.RecoilCount > 0 {
		frog.RecoilCount--
		if frog.RecoilCount <= 20 {
			if frog.RecoilCount == 1 {
				frog.CenterX = frog.RecoilX1
				frog.CenterY = frog.RecoilY1
			} else {
				if frog.RecoilCount > 14 {
					frog.CenterX = (frog.RecoilCount-16)*(frog.RecoilX1-frog.RecoilX2)/5 + frog.RecoilX2
					frog.CenterY = (frog.RecoilCount-16)*(frog.RecoilY1-frog.RecoilY2)/5 + frog.RecoilY2
				} else {
					frog.CenterX = frog.RecoilX1 + frog.RecoilCount*(frog.RecoilX2-frog.RecoilX1)/15
					frog.CenterY = frog.RecoilY1 + frog.RecoilCount*(frog.RecoilY2-frog.RecoilY1)/15
				}
			}
		}
	}

	if frog.BlinkCount > 0 {
		frog.BlinkCount--
	}

	if frog.State == FROGSTATE_FIRING {
		frog.StatePercent += 0.15
		if frog.StatePercent > 0.6 {
			frog.Bullet.Update()
		}
	} else {
		frog.StatePercent += 0.07
	}

	if frog.StatePercent > 1 {
		frog.StatePercent = 1
		if frog.State == FROGSTATE_RELOADING {
			frog.State = FROGSTATE_NORMAL
		}
	}
	frog.CalcAngle()
}
