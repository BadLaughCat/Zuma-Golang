package main

import (
	"image/color"
	"math"
	"math/rand"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var globalIdGen int32 = 0
var globalBallColors [6]color.RGBA = [6]color.RGBA{
	{25, 128, 255, 255}, {255, 255, 0, 255}, {255, 0, 0, 255},
	{0, 255, 0, 255}, {255, 0, 255, 255}, {255, 255, 255, 255},
}
var globalBrightBallColors [6]color.RGBA = [6]color.RGBA{
	{128, 255, 255, 255}, {255, 255, 64, 255}, {255, 170, 170, 255},
	{128, 255, 128, 255}, {255, 128, 255, 255}, {255, 255, 255, 255},
}
var globalDarkBallColors [6]color.RGBA = [6]color.RGBA{
	{35, 22, 121, 255}, {96, 81, 10, 255}, {160, 15, 20, 255},
	{32, 68, 34, 255}, {86, 22, 67, 255}, {56, 27, 34, 255},
}
var globalTextBallColors [6]color.RGBA = [6]color.RGBA{
	{45, 139, 255, 255}, {255, 255, 0, 255}, {255, 0, 0, 255},
	{0, 255, 0, 255}, {255, 0, 255, 255}, {255, 255, 255, 255},
}

type Ball struct {
	Id                       int32
	Type                     int32
	X, Y                     float32
	WayPoint                 float32
	Rotation, DestRotation   float32
	RotationInc              float32
	Particles                *[60]Particle
	PowerType, DestPowerType PowerType
	StartFrame               int32
	CollidesWithNext         bool
	NeedCheckCollision       bool
	Bullet                   *Bullet
	ClearCount               int32
	SuckCount                int32
	SuckPending              bool
	BackwardsCount           int32
	BackwardsSpeed           float32
	PowerCount               int32
	PowerFade                int32
	ComboCount, ComboScore   int32
	GapCount, GapBonus       int32
}

type PowerType int32

const (
	PowerType_Bomb PowerType = iota
	PowerType_SlowDown
	PowerType_Accuracy
	PowerType_MoveBackwards
	PowerType_Max
	PowerType_None PowerType = 4
)

type Particle struct {
	X, Y, VX, VY float32
	Size         int32
}

func NewBall() *Ball {
	tmp := &Ball{}
	globalIdGen++
	tmp.Id = globalIdGen
	tmp.PowerType = PowerType_Max
	tmp.DestPowerType = PowerType_Max
	return tmp
}

func (ball *Ball) BeforeDestroy() {
	ball.Particles = nil
}

func (ball *Ball) CollidesWith(theBall *Ball, thePad int32) bool {
	return math.Abs(float64(int32(ball.WayPoint)-int32(theBall.WayPoint))) < float64(2*(thePad+DefaultBallRadius))
}

func (ball *Ball) CollidesWithPhysically(theBall *Ball, thePad int32) bool {
	dx, dy := theBall.X-ball.X, theBall.Y-ball.Y
	r := float32(thePad + DefaultBallRadius)
	return dx*dx+dy*dy < r*(r*4)
}

func (ball *Ball) DoDraw() {
	if ball.PowerType == PowerType_None {
		the_texture := gTextures[Texture_BlueBall+TextureKey(ball.Type)]
		num_rows := the_texture.Height / the_texture.Width
		frame := (ball.StartFrame + int32(ball.WayPoint)) % num_rows
		rl.DrawTexturePro(the_texture, rl.NewRectangle(0, float32(frame*the_texture.Width), float32(the_texture.Width), float32(the_texture.Width)),
			rl.NewRectangle(ball.X, ball.Y, float32(DefaultBallRadius*2), float32(DefaultBallRadius*2)), rl.NewVector2(float32(DefaultBallRadius), float32(DefaultBallRadius)),
			-ball.Rotation*rl.Rad2deg, rl.White,
		)
	} else {
		ball.DrawPower()
	}
}

func (ball *Ball) Draw() {
	if ball.ClearCount != 0 {
		ball.DrawExplosion()
	} else {
		ball.DoDraw()
		if ball.PowerFade != 0 && (ball.PowerFade&0x10) != 0 {
			rl.BeginBlendMode(rl.BlendAdditive)
			ball.DoDraw()
		}
		if globalBallBlink {
			rl.BeginBlendMode(rl.BlendAdditive)
			ball.DoDraw()
		}
	}

	rl.EndBlendMode()
}

func (ball *Ball) DrawBomb() {
	texture := gTextures[Texture_BlueBomb+TextureKey(ball.Type)]
	x, y := ball.X-float32(texture.Width/2), ball.Y-float32(texture.Height/2)
	rl.DrawTextureV(texture, rl.NewVector2(x, y), rl.White)

	var alpha int32 = globalBoard.StateCount
	if alpha%50 <= 9 {
		alpha = 0
	} else if alpha%50 <= 24 {
		alpha = (200*(alpha%50) - 2000) / 15
	} else if alpha%50 <= 34 {
		alpha = 200
	} else if alpha%50 <= 49 {
		alpha = 200 - (200*(alpha%50)-7000)/15
	}

	rl.BeginBlendMode(rl.BlendAdditive)
	color := rl.NewColor(uint8(alpha), uint8(alpha), uint8(alpha), 255)
	light_texture := gTextures[Texture_BlueLight+TextureKey(ball.Type)]
	rl.DrawTextureV(light_texture, rl.NewVector2(x+7, y+9), color)

	rl.EndBlendMode()
}

func (ball *Ball) DrawExplosion() {
	width, height := gTextures[Texture_BallExplosion].Width, gTextures[Texture_BallExplosion].Height
	image_rows := height / width
	cell_height := height / image_rows
	img_x := ball.X - float32(width)/2
	img_y := ball.Y - float32(cell_height)/2

	cel := ball.ClearCount / 3
	if cel < image_rows {
		angle := float32(ball.StartFrame) * math.Pi / 25
		rl.DrawTexturePro(gTextures[Texture_BallExplosion], rl.NewRectangle(0, float32(cel*cell_height), float32(width), float32(cell_height)),
			rl.NewRectangle(img_x, img_y, float32(width), float32(cell_height)), rl.NewVector2(0, 0),
			angle, globalBrightBallColors[ball.Type],
		)
	}

	if ball.Particles != nil {
		for i := range 60 {
			particle := &(*ball.Particles)[i]
			color := globalBrightBallColors[ball.Type]
			if ball.ClearCount > 20 {
				color.A = uint8(255 - (255*ball.ClearCount-5100)/20)
			}
			rl.DrawRectangleRec(rl.NewRectangle(
				float32(ball.ClearCount)*particle.VX+particle.X, float32(ball.ClearCount)*particle.VY+particle.Y,
				float32(particle.Size), float32(particle.Size),
			), color)
		}
	}
}

func (ball *Ball) DrawPower() {
	switch ball.PowerType {
	case PowerType_Bomb:
		ball.DrawBomb()
	case PowerType_SlowDown:
		ball.DrawStandardPower(Texture_BlueSlow, Texture_SlowLight)
	case PowerType_Accuracy:
		ball.DrawStandardPower(Texture_BlueAccuracy, Texture_AccuracyLight)
	case PowerType_MoveBackwards:
		ball.DrawStandardPower(Texture_BlueBackwards, Texture_BackwardsLight)
	}
}

func (ball *Ball) DrawShadow() {
	if ball.ClearCount == 0 {
		rl.DrawTextureV(gTextures[Texture_BallShadow], rl.NewVector2(
			ball.X-float32(gTextures[Texture_BallShadow].Width/2)-3,
			ball.Y-float32(gTextures[Texture_BallShadow].Height/2)+5,
		), rl.NewColor(0, 0, 0, 128))
	}
}

func (ball *Ball) DrawStandardPower(theBallImageId, theBlinkImageId TextureKey) {
	ball_texture := gTextures[theBallImageId+TextureKey(ball.Type)]
	blink_texture := gTextures[theBlinkImageId]

	rl.DrawTexturePro(ball_texture, rect(0, 0, ball_texture.Width, ball_texture.Height),
		rl.NewRectangle(ball.X, ball.Y, float32(DefaultBallRadius*2), float32(DefaultBallRadius*2)),
		vec2(DefaultBallRadius, DefaultBallRadius), -(ball.Rotation+math.Pi/2)*rl.Rad2deg, rl.White)

	var alpha int32 = 0
	time := globalBoard.StateCount % 100
	if time < 20 {
		alpha = 0
	} else if time < 50 {
		alpha = (255*time - 5100) / 30
	} else if time < 70 {
		alpha = 255
	} else if time < 100 {
		alpha = 255 - (255*time-17850)/30
	}

	ball_color := globalDarkBallColors[ball.Type]
	ball_color.A = uint8(alpha)
	rl.DrawTexturePro(blink_texture, rect(0, 0, blink_texture.Width, blink_texture.Height),
		rl.NewRectangle(ball.X, ball.Y, float32(blink_texture.Width), float32(blink_texture.Height)),
		vec2(blink_texture.Width/2, blink_texture.Height/2), -(ball.Rotation+math.Pi/2)*rl.Rad2deg, ball_color)

	rl.EndBlendMode()
}

func (ball *Ball) GetCollidesWithPrev(list []*Ball) bool {
	prev_ball := ball.GetPrevBall(false, list)
	if prev_ball != nil {
		return prev_ball.CollidesWithNext
	}
	return false
}

func (ball *Ball) GetNextBall(mustCollide bool, list []*Ball) *Ball {
	if len(list) == 0 {
		return nil
	} else {
		index := slices.Index(list, ball)
		if index == -1 || index == len(list)-1 {
			return nil
		}
		index++
		if !mustCollide || ball.CollidesWithNext {
			return list[index]
		}
		return nil
	}
}

func (ball *Ball) GetPowerTypeWussy() PowerType {
	if ball.PowerType == PowerType_None {
		return ball.DestPowerType
	}
	return ball.PowerType
}

func (ball *Ball) GetPrevBall(mustCollide bool, list []*Ball) *Ball {
	if len(list) == 0 {
		return nil
	}
	index := slices.Index(list, ball)
	if index < 1 {
		return nil
	}
	index--
	if !mustCollide {
		return list[index]
	} else {
		if list[index].CollidesWithNext {
			return list[index]
		}
		return nil
	}
}

func (ball *Ball) InsertInList(theList *[]*Ball, index int) {
	*theList = slices.Insert(*theList, index, ball)
}

func (ball *Ball) Intersects(p1, v1 rl.Vector3, t *float32) bool {
	delta := rl.NewVector2(p1.X-ball.X, p1.Y-ball.Y)
	a := rl.Vector2LengthSqr(rl.NewVector2(v1.Y, v1.X))
	b := v1.X * delta.X
	b += v1.Y * delta.Y
	b = b + b
	disc := b*b - (rl.Vector2LengthSqr(delta)-float32(DefaultBallRadius*DefaultBallRadius))*a*4
	if disc < 0 {
		return false
	}
	disc = float32(math.Sqrt(float64(disc)))

	*t = (-b - disc) / (2 * a)
	return true
}

func (ball *Ball) RandomizeFrame() {
	ball.StartFrame = rand.Int31n(50)
}

func (ball *Ball) SetCollidesWithPrev(collidesWithPrev bool, list []*Ball) {
	prev_ball := ball.GetPrevBall(false, list)
	if prev_ball != nil {
		prev_ball.CollidesWithNext = collidesWithPrev
	}
}

func (ball *Ball) SetFrame(theFrame int32) {
	the_texture := gTextures[TextureKey(ball.Type+int32(Texture_BlueBall))]
	num_rows := the_texture.Height / the_texture.Width
	ball.StartFrame = num_rows - int32(float32(theFrame)+ball.WayPoint)%num_rows
}

func (ball *Ball) SetPowerType(theType PowerType, delay bool) {
	if theType == ball.PowerType {
		return
	}
	if delay {
		ball.DestPowerType = theType
		ball.PowerFade = 100
	} else {
		ball.DestPowerType = PowerType_None
		ball.PowerType = theType
	}
}

func (ball *Ball) SetRotation(theRot float32, immediate bool) {
	if immediate {
		ball.Rotation = theRot
		return
	}

	for math.Abs(float64(theRot-ball.Rotation)) > math.Pi {
		if theRot > ball.Rotation {
			theRot -= 6.2831802
		} else {
			theRot += 6.2831802
		}
	}

	ball.DestRotation = theRot
	ball.RotationInc = 0.104719669
	if theRot < ball.Rotation {
		ball.RotationInc = -ball.RotationInc
	}
}

func (ball *Ball) StartClearCount(inTunnel bool) {
	if ball.ClearCount != 0 {
		return
	}
	ball.ClearCount = 1

	if !inTunnel {
		if ball.Particles == nil {
			ball.Particles = new([60]Particle)
		}
		for i := range 60 {
			ptcl := &(*ball.Particles)[i]
			angle := float64(rand.Int31n(360)) * rl.Deg2rad
			speed := float32(rand.Int31n(500)) / 500
			ptcl.VX = float32(math.Sin(angle)) * speed
			ptcl.VY = float32(math.Cos(angle)) * speed

			rnd := float32(rand.Int31n(30))
			ptcl.X = rnd*ptcl.VX + ball.X
			ptcl.Y = rnd*ptcl.VY + ball.Y
			ptcl.Size = 1
			if rand.Int31n(10) < 2 {
				ptcl.Size++
			}
		}
	}
}

func (ball *Ball) UpdateCollisionInfo(thePad int32, list []*Ball) {
	prev_ball := ball.GetPrevBall(false, list)
	next_ball := ball.GetNextBall(false, list)
	if prev_ball != nil {
		prev_ball.CollidesWithNext = prev_ball.CollidesWith(ball, thePad)
	}
	if next_ball != nil {
		ball.CollidesWithNext = next_ball.CollidesWith(ball, thePad)
	} else {
		ball.CollidesWithNext = false
	}
}

func (ball *Ball) UpdateRotation() {
	if ball.PowerFade > 0 {
		ball.PowerFade--
		if ball.PowerFade == 0 {
			ball.PowerType = ball.DestPowerType
			ball.DestPowerType = PowerType_None
			if ball.PowerType != PowerType_None {
				ball.PowerCount = 2000
			}
		}
	}
	if ball.PowerCount > 0 {
		ball.PowerCount--
		if ball.PowerCount <= 0 && ball.PowerType != PowerType_None {
			ball.DestPowerType = PowerType_None
			ball.PowerFade = 100
		}
	}
	if ball.RotationInc != 0 {
		ball.Rotation += ball.RotationInc
		if ball.RotationInc > 0 && ball.Rotation > ball.DestRotation {
			ball.Rotation = ball.DestRotation
			ball.RotationInc = 0
		} else if ball.RotationInc < 0 && ball.Rotation < ball.DestRotation {
			ball.Rotation = ball.DestRotation
			ball.RotationInc = 0
		}
	}
}
