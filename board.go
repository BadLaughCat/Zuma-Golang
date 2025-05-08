package main

import (
	"fmt"
	"image/color"
	"math"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Board struct {
	Frog                      *Frog
	BallColorMap              map[int32]int32
	BulletList                []*Bullet
	CurveList                 []Curve
	BallDrawer                *BallDrawer
	ParticleMgr               *ParticleMgr
	SpriteMgr                 *SpriteMgr
	SoundMgr                  *SoundMgr
	StateCount                int32
	LastExplosionTick         uint32
	LastBallClickTick         uint32
	AccuracyCount             int32
	ShowGuide                 bool
	DoGuide                   bool
	RecalcGuide               bool
	Guide                     [4]rl.Vector2
	GuideCenter               rl.Vector3
	LevelBeginning            bool
	IsWinning                 bool
	IsEndless                 bool
	HasReachedTarget          bool
	NumClearsInARow           int32
	CurInARowBonus            int32
	NumCleared                int32
	ClearedXSum               int32
	ClearedYSum               int32
	CurComboCount             int32
	CurComboScore             int32
	Score, ScoreDisplay       int32
	ScoreTarget               int32
	LevelBeginScore           int32
	Lives, LivesBlinkCount    int32
	CurBarSize, TargetBarSize int32
	BarBlinkCount             int32
	FlashCount                int32
	LevelEndFrame             int32
	NeedComboCount            []*Ball
	LevelDesc                 *LevelDesc
	LevelStats                GameStats
	GameState                 GameState
}

type GameStats struct {
	TimePlayed                int32
	NumBallsCleared           int32
	NumGemsCleared            int32
	NumGaps                   int32
	NumCombos                 int32
	MaxCombo, MaxComboScore   int32
	MaxInARow, MaxInARowScore int32
}

type GameState int32

const (
	GameState_None GameState = iota
	GameState_Playing
	GameState_Losing
	GameState_LevelUp
	GameState_LevelBegin
)

func NewBoard() *Board {
	tmp := &Board{
		BallColorMap: make(map[int32]int32),
		Frog:         NewFrog(),
		Lives:        3,
		BallDrawer:   new(BallDrawer),
		SpriteMgr:    NewSpriteMgr(),
		SoundMgr:     InitSoundManager(),
		LevelDesc:    NewLevelDesc(),
	}
	tmp.ParticleMgr = &ParticleMgr{Board: tmp}
	return tmp
}

func (b *Board) ActivatePower(theBall *Ball) {
	switch theBall.GetPowerTypeWussy() {
	case PowerType_Bomb:
		ticks := b.GetTickCount()
		if ticks-b.LastExplosionTick > 250 {
			b.LastExplosionTick = ticks
		}
	case PowerType_MoveBackwards:
	case PowerType_SlowDown:
	case PowerType_Accuracy:
		b.AccuracyCount = 2000
		b.DoAccuracy(true)
	}

	for i := range b.CurveList {
		b.CurveList[i].ActivatePower(theBall)
	}
}

func (b *Board) AdvanceFreeBullet(index *int) {
	bullet := b.BulletList[*index]
	bullet.Update()

	for i := range b.CurveList {
		if b.CurveList[i].CheckCollision(bullet) {
			b.BulletList[*index] = nil
			b.BulletList = slices.Delete(b.BulletList, *index, *index+1)
			return
		}
	}

	for i := range b.CurveList {
		b.CurveList[i].CheckGapShot(bullet)
	}

	if bullet.X >= 0 && bullet.Y >= 0 &&
		bullet.X-float32(DefaultBallRadius) < float32(GameWidth) &&
		bullet.Y-float32(DefaultBallRadius) < float32(GameHeight) {
		*index++
	} else {
		b.ResetInARowBonus()
		bullet = nil
		b.BulletList[*index] = nil
		b.BulletList = slices.Delete(b.BulletList, *index, *index+1)
	}
}

func (b *Board) DoAccuracy(accuracy bool) {
	b.RecalcGuide, b.ShowGuide, b.DoGuide = accuracy, accuracy, accuracy
	if accuracy {
		b.Frog.FireVel = 15
	} else {
		b.AccuracyCount = 0
		b.Frog.FireVel = b.LevelDesc.FireSpeed
	}
}

func (b *Board) CheckReload() {
	if len(b.BallColorMap) == 0 {
		return
	}
	bullet := b.Frog.Bullet
	if bullet != nil {
		if _, found := b.BallColorMap[bullet.Type]; !found {
			b.Frog.Bullet.Type = randomly_get_map_key(b.BallColorMap)
		}
	}

	bullet = b.Frog.NextBullet
	if bullet != nil {
		if _, found := b.BallColorMap[bullet.Type]; !found {
			b.Frog.NextBullet.Type = randomly_get_map_key(b.BallColorMap)
		}
	}

	for b.Frog.NeedsReload() {
		b.Frog.Reload(randomly_get_map_key(b.BallColorMap), true, PowerType_Max)
	}
}

func (b *Board) Draw() {
	b.DrawPlaying()
	b.DrawText()
	b.DrawOverlay()
}

func (b *Board) DrawBullets() {
	for i := range b.BulletList {
		b.BulletList[i].DrawShadow()
	}
	for i := range b.BulletList {
		b.BulletList[i].Draw()
	}
}

func (b *Board) DrawOverlay() {
	b.ParticleMgr.DrawTopMost()
}

func (b *Board) DrawPlaying() {
	b.SpriteMgr.DrawBackground()
	b.BallDrawer.Reset()
	for i := range b.CurveList {
		b.CurveList[i].DrawBalls(b.BallDrawer)
	}
	b.BallDrawer.Draw(b.ParticleMgr)
	b.Frog.Draw()

	if b.ShowGuide {
		var alpha int32 = 120
		if b.AccuracyCount <= 300 {
			alpha = (120*b.AccuracyCount)/300 + 8
		}
		color := color.RGBA{0, 255, 255, uint8(alpha)}
		rl.DrawTriangle(b.Guide[0], b.Guide[1], b.Guide[2], color)
		rl.DrawTriangle(b.Guide[2], b.Guide[3], b.Guide[0], color)
	}

	b.DrawBullets()
}

func (b *Board) DrawText() {
	text := ""
	more_than_3 := false
	var show_count int32 = 0

	rl.DrawRectangle(25, 3, 80, 23, color.RGBA{19, 50, 9, 255})
	if (b.LivesBlinkCount & 0x10) == 0 {
		lives := b.Lives - 1
		if b.GameState == GameState_Losing {
			lives = b.Lives
		}

		if b.IsEndless {
			text = "Survival"
			text_width := rl.MeasureText(text, 16)
			rl.DrawText(text, 64-text_width/2, 6, 16, color.RGBA{255, 255, 0, 255})
			more_than_3, show_count = false, 0
		} else {
			if lives == 0 {
				text = "Last Life"
				text_width := rl.MeasureText(text, 16)
				rl.DrawText(text, 64-text_width/2, 6, 16, color.RGBA{255, 255, 0, 255})
				more_than_3, show_count = false, 0
			} else if lives <= 3 {
				more_than_3, show_count = false, lives
			} else {
				more_than_3, show_count = true, 1
			}
		}

		var frog_x int32 = 28
		for range show_count {
			rl.DrawTexture(gTextures[Texture_Life], frog_x, 3, rl.White)
			frog_x += 26
		}

		if more_than_3 {
			rl.DrawText(fmt.Sprintf("x%d", lives), frog_x+4, 6, 16, color.RGBA{255, 255, 0, 255})
		}
	}
}

func (b *Board) GetTickCount() uint32 {
	return 10 * uint32(b.StateCount)
}

func (b *Board) IncScore(theInc int32, delayDisplay bool) {
	if theInc <= 0 {
		return
	}
	score := b.Score
	b.Score += theInc
	if score/50000 < (score+theInc)/50000 && !b.IsEndless && !b.IsWinning {
		b.Lives += (score+theInc)/50000 - score/50000
		b.LivesBlinkCount = 150
		rl.PlaySound(gSounds[Sound_ExtraLife])
		b.SoundMgr.AddSound(gSounds[Sound_ExtraLife], 30, 0, 0)
		b.SoundMgr.AddSound(gSounds[Sound_ExtraLife], 60, 0, 0)
	}
	if !delayDisplay {
		b.ScoreDisplay = b.Score
	}
}

func (b *Board) PlayBallClick(theSound SoundKey) {
	tick := b.GetTickCount()
	if tick-b.LastBallClickTick >= 250 {
		rl.PlaySound(gSounds[theSound])
		b.LastBallClickTick = tick
	}
}

func (b *Board) ResetInARowBonus() {
	if b.NumClearsInARow > b.LevelStats.MaxInARow {
		b.LevelStats.MaxInARow = b.NumClearsInARow
		b.LevelStats.MaxInARowScore = b.CurInARowBonus
	}
	b.NumClearsInARow, b.CurInARowBonus = 0, 0
}

func (b *Board) StartLevel() {
	b.GameState = GameState_Playing
	b.StateCount = 0
	b.Frog.FireVel = b.LevelDesc.FireSpeed
	//b.Frog.SetPos(b.LevelDesc.FrogX, b.LevelDesc.FrogY)
	b.SoundMgr.PlayLoop(LoopType_RollIn)
	b.LevelBeginning = true
}

func (b *Board) Update() {
	b.Frog.Update()
	b.SpriteMgr.Update()
	b.ParticleMgr.Update()
	b.SoundMgr.Update()

	b.StateCount++

	b.UpdatePlaying()
	if b.DoGuide {
		b.UpdateGuide()
	}
	b.UpdateMiscStuff()
}

func (b *Board) UpdateBallColorMap(theBall *Ball, added bool) {
	if added {
		b.BallColorMap[theBall.Type]++
	} else {
		if _, found := b.BallColorMap[theBall.Type]; found {
			b.BallColorMap[theBall.Type]--
			if b.BallColorMap[theBall.Type] <= 0 {
				delete(b.BallColorMap, theBall.Type)
			}
		}
	}
}

func (b *Board) UpdateBullets() {
	for i := 0; i != len(b.BulletList); {
		b.AdvanceFreeBullet(&i)
	}
}

func (b *Board) UpdateGuide() {
	angle := b.Frog.Angle - math.Pi/2
	dx, dy := float32(math.Sin(float64(angle))), float32(math.Cos(float64(angle)))
	dx2, dy2 := dx, dy
	dx3, dy3 := dx*16, dy*16

	center := rl.NewVector3(float32(b.Frog.CenterX)+dy*50, float32(b.Frog.CenterY)-dx*50, 0)
	g1 := rl.NewVector3(center.X-dx3, center.Y-dy3, 0)
	g2 := rl.NewVector3(center.X+dx3, center.Y+dy3, 0)
	v1 := rl.NewVector3(float32(math.Cos(float64(angle))), -float32(math.Sin(float64(angle))), 0)
	var t float32 = 10000000

	var ball *Ball = nil
	for i := range b.CurveList {
		intersect_ball := b.CurveList[i].CheckBallIntersection(g1, v1, &t)
		if intersect_ball != nil {
			ball = intersect_ball
		}
	}
	for i := range b.CurveList {
		intersect_ball := b.CurveList[i].CheckBallIntersection(g2, v1, &t)
		if intersect_ball != nil {
			ball = intersect_ball
		}
	}

	if ball == nil {
		t = 1000 / rl.Vector3Length(v1)
	}

	guide := rl.Vector3Add(center, rl.Vector3Scale(v1, t))
	if !b.RecalcGuide && b.ShowGuide && rl.Vector3Length(rl.Vector3Subtract(b.GuideCenter, guide)) < 20 {
		return
	}

	b.GuideCenter = guide
	b.ShowGuide = true
	b.RecalcGuide = false

	b.Guide[0].X = g1.X + dx3*0.5
	b.Guide[0].Y = g1.Y + dy3*0.5
	b.Guide[1].X = g2.X - dx3*0.5
	b.Guide[1].Y = g2.Y - dy3*0.5
	b.Guide[2].X = guide.X + dx2
	b.Guide[2].Y = guide.Y + dy2
	b.Guide[3].X = guide.X - dx2
	b.Guide[3].Y = guide.Y - dy2
}

func (b *Board) UpdateMiscStuff() {
	if b.ScoreDisplay != b.Score {
		if b.ScoreDisplay < b.Score {
			b.ScoreDisplay += int32(float32(b.Score-b.ScoreDisplay)/50.0 + 2.0)
		}
		if b.ScoreDisplay > b.Score {
			b.ScoreDisplay = b.Score
		}
	}

	if b.CurBarSize != b.TargetBarSize {
		if b.CurBarSize < b.TargetBarSize {
			b.CurBarSize++
		}
		if b.CurBarSize > b.TargetBarSize {
			b.CurBarSize--
		}
	}

	if b.BarBlinkCount > 0 {
		b.BarBlinkCount--
	}

	if b.FlashCount > 0 {
		globalBallBlink = true
		b.FlashCount--
		if b.FlashCount == 0 {
			globalBallBlink = false
		}
	}

	if b.LivesBlinkCount > 0 {
		b.LivesBlinkCount--
	}

	if b.AccuracyCount > 0 {
		b.AccuracyCount--
		if b.AccuracyCount == 0 {
			b.DoAccuracy(false)
		}
	}
}

func (b *Board) UpdatePlaying() {
	bullet := b.Frog.GetFiredBullet()
	if bullet != nil {
		b.BulletList = append(b.BulletList, bullet)
		b.CheckReload()
	}

	b.UpdateBullets()

	if b.LevelBeginning {
		still_starting := false
		for i := range len(b.CurveList) {
			if !b.CurveList[i].HasReachedCruisingSpeed() {
				still_starting = true
			}
		}
		if !still_starting || b.StateCount > 500 {
			b.LevelBeginning = false
			b.SoundMgr.StopLoop(LoopType_RollIn)
		}
	}

	if !b.HasReachedTarget && b.CurBarSize == 256 && b.Score >= b.ScoreTarget {
		b.BarBlinkCount = 224
		b.FlashCount = 50
		if b.IsEndless {

		} else {
			b.HasReachedTarget = true
			for i := range b.CurveList {
				b.CurveList[i].SetStopAddingBalls(true)
			}
		}
	}

	score_diff := b.ScoreTarget - b.LevelBeginScore
	if score_diff > 0 {
		score_remaining := b.ScoreTarget - b.Score
		if score_remaining < 0 {
			if b.LevelEndFrame != 0 {
				if b.StateCount-3000 == b.LevelEndFrame {
					for i := range b.CurveList {
						b.LevelDesc.CurveDescs[i].PowerUpFreq[PowerType_Bomb] = 500
						b.LevelDesc.CurveDescs[i].PowerUpFreq[PowerType_SlowDown] = 0
						b.LevelDesc.CurveDescs[i].PowerUpFreq[PowerType_Accuracy] = 0
						b.LevelDesc.CurveDescs[i].PowerUpFreq[PowerType_MoveBackwards] = 0
						b.LevelDesc.CurveDescs[i].AccelerationRate = 0.0003
					}
				}
			} else if len(b.BallColorMap) == 2 {
				b.LevelEndFrame = b.StateCount
			}
		}
		b.TargetBarSize = 256 - (score_remaining * 256 / score_diff)
	}

	for i := range b.CurveList {
		b.CurveList[i].UpdatePlaying()
	}

	if b.StateCount > 50 {
		b.CheckReload()
	}
}
