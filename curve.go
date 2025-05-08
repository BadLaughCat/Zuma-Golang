package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type BallDrawer struct {
	NumBalls, NumShadows [MaxPriority]int32
	Balls, Shadows       [MaxPriority][1024]*Ball
}

func (drawer *BallDrawer) Draw(theSpriteMgr *SpriteMgr, theParticleMgr *ParticleMgr) {
	for i := range MaxPriority {
		theSpriteMgr.DrawSprites(i)
		theParticleMgr.Draw(i)
		for k := range drawer.NumShadows[i] {
			drawer.Shadows[i][k].DrawShadow()
		}
		for k := range drawer.NumBalls[i] {
			drawer.Balls[i][k].Draw()
		}
	}
}

func (drawer *BallDrawer) Reset() {
	for i := range MaxPriority {
		drawer.NumBalls[i] = 0
		drawer.NumShadows[i] = 0
	}
}

type Curve struct {
	Board                   *Board
	BulletList              []*Bullet
	BallList                []*Ball
	PendingBalls            []*Ball
	WayPointMgr             *WayPointMgr
	LevelDesc               *LevelDesc
	CurveDesc               *CurveDesc
	CurveIndex              int32
	SpriteMgr               *SpriteMgr
	LastPowerUpFrame        [PowerType_Max]int32
	StopTime                int32
	SlowCount               int32
	BackwardCount           int32
	TotalBalls              int32
	AdvanceSpeed            float32
	FirstChainEnd           int32
	DangerPoint             int32
	PathLightEndFrame       int32
	LastClearedBallPoint    int32
	LastPathShowTick        uint32
	FirstBallMovedBackwards bool
	HaveSets                bool
	HadPowerUp              bool
	StopAddingBalls         bool
	InDanger                bool
}

var globalGotPowerUp [PowerType_Max]bool = [PowerType_Max]bool{false, false, false, false}

func (curve *Curve) ActivateBomb(theBall *Ball) {
	color := globalBallColors[theBall.Type]
	x, y := int32(theBall.X), int32(theBall.Y)
	curve.Board.ParticleMgr.AddExplosion(x, y, 0, color, 5)
	v19 := gTextures[Texture_Explosion].Width / 3

	var a6 int32 = 7
	for i := v19; i < 100; i += v19 {
		var v21 float32 = 0
		for v21 < math.Pi*2 {
			curve.Board.ParticleMgr.AddExplosion(
				x+(rand.Int31n(21)-10)+int32(math.Sin(float64(v21))*float64(i)),
				y+(rand.Int31n(21)-10)+int32(math.Cos(float64(v21))*float64(i)),
				0, color, a6,
			)
			v21 += float32(v19) / float32(i)
		}
		a6 += 4
	}

	curve.Board.ParticleMgr.AddExplosion(x, y, 0, color, 0)

	for i := range curve.BallList {
		ball := curve.BallList[i]
		if ball.ClearCount == 0 && ball.CollidesWithPhysically(theBall, 45) {
			ball.ComboScore, ball.ComboCount = curve.Board.CurComboScore, curve.Board.CurComboCount
			curve.Board.NeedComboCount = append(curve.Board.NeedComboCount, ball)
			curve.StartClearCount(ball)
			curve.Board.ParticleMgr.AddExplosion(int32(ball.X), int32(ball.Y), 0, color, 0)
		}
	}
}

func (curve *Curve) ActivatePower(theBall *Ball) {
	power_type := theBall.GetPowerTypeWussy()
	globalGotPowerUp[power_type] = true

	if power_type == PowerType_Bomb {
		curve.ActivateBomb(theBall)
	} else if power_type == PowerType_MoveBackwards {
		if len(curve.BallList) != 0 {
			curve.BackwardCount = 300
		}
	} else if power_type == PowerType_SlowDown {
		if curve.SlowCount < 1000 {
			curve.SlowCount = 800
		}
	}
}

func (curve *Curve) AddBall() {
	if len(curve.PendingBalls) == 0 {
		if curve.CurveDesc.NumBalls != 0 || curve.StopAddingBalls {
			return
		}
		curve.AddPendingBall()
	}

	ball := curve.PendingBalls[0]
	curve.WayPointMgr.SetWayPoint(ball, 1)

	if len(curve.BallList) != 0 {
		front_ball := curve.BallList[0]
		if ball.WayPoint > front_ball.WayPoint || front_ball.CollidesWith(ball, 0) {
			return
		}
	}

	curve.Board.UpdateBallColorMap(ball, true)

	ball.InsertInList(&curve.BallList, 0)
	ball.UpdateCollisionInfo(5+int32(curve.AdvanceSpeed), curve.BallList)
	ball.NeedCheckCollision = true
	ball.SetRotation(curve.WayPointMgr.GetRotationForPoint(int(ball.WayPoint)), true)
	ball.BackwardsCount = 0
	ball.SuckCount = 0
	ball.GapBonus, ball.GapCount = 0, 0
	ball.ComboScore, ball.ComboCount = 0, 0

	curve.PendingBalls[0].BeforeDestroy()
	curve.PendingBalls[0] = nil
	curve.PendingBalls = curve.PendingBalls[1:]
}

func (curve *Curve) AddPendingBall() {
	var new_color, prev_color, num_colors int32 = 0, 0, curve.CurveDesc.NumColors
	ball := NewBall()
	ball.RandomizeFrame()

	if len(curve.PendingBalls) != 0 {
		prev_color = curve.PendingBalls[len(curve.PendingBalls)-1].Type
	} else if len(curve.BallList) != 0 {
		prev_color = curve.BallList[0].Type
	} else {
		prev_color = rand.Int31n(num_colors)
	}

	if prev_color >= num_colors {
		prev_color = rand.Int31n(num_colors)
	}

	max_single := curve.CurveDesc.MaxSingle
	if rand.Int31n(100) <= curve.CurveDesc.BallRepeat {
		new_color = prev_color
	} else if max_single < 10 && curve.GetNumPendingSingles(1) == 1 && (max_single == 0 || curve.GetNumPendingSingles(10) > max_single) {
		new_color = prev_color
	} else {
		for new_color == prev_color {
			new_color = rand.Int31n(num_colors)
		}
	}
	ball.Type = new_color
	curve.PendingBalls = append(curve.PendingBalls, ball)
}

func (curve *Curve) AddPowerUp(thePower PowerType) {
	ball_idx := rand.Intn(len(curve.BallList))
	ball := curve.BallList[ball_idx]
	if ball.PowerType == PowerType_None && ball.DestPowerType == PowerType_None {
		ball.SetPowerType(thePower, true)
	}
}

func (curve *Curve) AdvanceBackwardBalls() {
	curve.FirstBallMovedBackwards = false
	if len(curve.BallList) == 0 {
		return
	}

	collided := false
	var backwards_speed float32 = 0
	if curve.BackwardCount != 0 {
		last_one := curve.BallList[len(curve.BallList)-1]
		last_one.BackwardsSpeed, last_one.BackwardsCount = 1, 1
	}

	iter_index := len(curve.BallList) - 1
	for {
		ball := curve.BallList[iter_index]
		backwards_count := ball.BackwardsCount
		if backwards_count > 0 {
			backwards_speed = ball.BackwardsSpeed
			curve.WayPointMgr.SetWayPoint(ball, ball.WayPoint-backwards_speed)
			ball.BackwardsCount--
			collided = true
		}

		iter_index--
		if iter_index == -1 {
			break
		}
		next_ball := curve.BallList[iter_index]

		if collided {
			if next_ball.CollidesWithNext {
				curve.WayPointMgr.SetWayPoint(next_ball, next_ball.WayPoint-backwards_speed)
			} else {
				way_off_next := ball.WayPoint - float32(DefaultBallRadius)*2
				if next_ball.WayPoint > way_off_next {
					next_ball.CollidesWithNext = true
					collided = true
					curve.Board.PlayBallClick(Sound_BallClick1)
					backwards_speed = next_ball.WayPoint - way_off_next
					next_ball.WayPoint = way_off_next
				} else {
					collided = false
				}
			}
		}
	}

	if collided {
		curve.FirstBallMovedBackwards = true
		if curve.StopTime < 20 {
			curve.StopTime = 20
		}
	}
}

func (curve *Curve) AdvanceBalls() {
	if len(curve.BallList) == 0 {
		return
	}

	max_speed := curve.CurveDesc.Speed
	if curve.CurveDesc.AccelerationRate != 0 {
		curve.CurveDesc.CurAcceleration += curve.CurveDesc.AccelerationRate
		max_speed += curve.CurveDesc.CurAcceleration
		if max_speed > curve.CurveDesc.MaxSpeed {
			max_speed = curve.CurveDesc.MaxSpeed
		}
	}
	if curve.SlowCount != 0 {
		max_speed /= 4
	}
	if curve.FirstChainEnd >= curve.DangerPoint-curve.CurveDesc.SlowDistance {
		if curve.FirstChainEnd < curve.DangerPoint {
			dist := float32(curve.FirstChainEnd-(curve.DangerPoint-curve.CurveDesc.SlowDistance)) / float32(curve.CurveDesc.SlowDistance)
			max_speed = (1-dist)*max_speed + dist*max_speed/float32(curve.CurveDesc.SlowFactor)
		} else {
			max_speed /= curve.CurveDesc.SlowFactor
		}
	}
	if curve.AdvanceSpeed > max_speed {
		curve.AdvanceSpeed -= 0.1
	}
	if curve.AdvanceSpeed < max_speed {
		curve.AdvanceSpeed += 0.005
		if curve.AdvanceSpeed >= max_speed {
			curve.AdvanceSpeed = max_speed
		}
	}

	ball := curve.BallList[0]
	next_way_point := ball.WayPoint
	if !curve.FirstBallMovedBackwards && curve.StopTime == 0 {
		curve.WayPointMgr.SetWayPoint(ball, next_way_point+curve.AdvanceSpeed)
	}

	var first_chain_end *Ball = nil
	iter_index := 0
	for iter_index < len(curve.BallList) {
		ball = curve.BallList[iter_index]
		iter_index++
		if iter_index == len(curve.BallList) {
			break
		}

		next_ball := curve.BallList[iter_index]
		next_way_point = next_ball.WayPoint
		way_point := ball.WayPoint
		if way_point > next_way_point-float32(DefaultBallRadius*2) {
			curve.WayPointMgr.SetWayPoint(next_ball, way_point+float32(DefaultBallRadius*2))
			if !ball.CollidesWithNext {
				ball.CollidesWithNext = true
				curve.Board.PlayBallClick(Sound_BallClick1)
			}
			ball.NeedCheckCollision = false
		}
		if first_chain_end == nil {
			if !ball.CollidesWithNext {
				first_chain_end = ball
			}
		}
	}

	if first_chain_end == nil {
		first_chain_end = curve.BallList[len(curve.BallList)-1]
	}

	curve.FirstChainEnd = int32(first_chain_end.WayPoint)

	if curve.FirstChainEnd >= curve.DangerPoint {
		tick := curve.Board.GetTickCount()
		max_time := 100 + 4000*(curve.GetCurveLength()-curve.FirstChainEnd)/(curve.GetCurveLength()-curve.DangerPoint)
		frame := curve.Board.StateCount
		if frame >= curve.PathLightEndFrame && tick-curve.LastPathShowTick >= uint32(max_time) {
			curve.LastPathShowTick = tick
			curve.PathLightEndFrame = frame + curve.DrawPathSparkles(curve.FirstChainEnd, 0, false)
		}
	}
	curve.InDanger = int32(curve.BallList[len(curve.BallList)-1].WayPoint) >= curve.DangerPoint
}

func (curve *Curve) AdvanceBullets() {
	i := 0
	for i != len(curve.BulletList) {
		curve.AdvanceMergingBullet(&i)
	}
}

func (curve *Curve) AdvanceMergingBullet(index *int) {
	bul := curve.BulletList[*index]
	bul.CheckSetHitBallToPrevBall(curve.BallList)
	hit_ball := bul.HitBall
	curve.WayPointMgr.SetWayPoint(&bul.Ball, hit_ball.WayPoint)
	curve.WayPointMgr.FindFreeWayPoint(hit_ball, &bul.Ball, bul.HitInFront, 0)
	bul.DestX, bul.DestY = bul.X, bul.Y
	bul.Update()

	push_ball := bul.GetPushBall(curve.BallList)
	if push_ball != nil {
		num := 1.0 - bul.HitPercent
		f := float32(-DefaultBallRadius) * num / 2
		point := push_ball.WayPoint
		percent := float32(DefaultBallRadius+DefaultBallRadius) * (bul.HitPercent * bul.HitPercent)
		curve.WayPointMgr.FindFreeWayPoint(&bul.Ball, bul.GetPushBall(curve.BallList), true, int32(f))
		if push_ball.WayPoint-bul.WayPoint > percent {
			end_point := bul.WayPoint + percent
			if end_point > point {
				curve.WayPointMgr.SetWayPoint(push_ball, end_point)
			} else {
				curve.WayPointMgr.SetWayPoint(push_ball, point)
			}
		}
		push_ball.NeedCheckCollision = true
	}

	if bul.HitPercent >= 1 {
		ball_iter := slices.Index(curve.BallList, hit_ball)
		if bul.HitInFront {
			ball_iter++
		}

		new_ball := NewBall()
		new_ball.SetRotation(bul.Rotation, true)
		new_ball.Type = bul.Type
		new_ball.SetPowerType(bul.PowerType, false)
		curve.WayPointMgr.SetWayPoint(new_ball, bul.WayPoint)
		new_ball.SetFrame(0)
		new_ball.InsertInList(&curve.BallList, ball_iter)
		curve.Board.UpdateBallColorMap(new_ball, true)

		min_gap_dist, num_gaps := bul.GetMinGapDist(), int32(len(bul.GapInfos))

		bul.BeforeDestroy()
		bul = nil

		curve.BulletList[*index] = nil
		curve.BulletList = slices.Delete(curve.BulletList, *index, *index+1)
		curve.TotalBalls++

		prev_ball, next_ball := new_ball.GetPrevBall(false, curve.BallList), new_ball.GetNextBall(false, curve.BallList)
		new_ball.UpdateCollisionInfo(5, curve.BallList)
		new_ball.NeedCheckCollision = true

		if prev_ball != nil && new_ball.GetCollidesWithPrev(curve.BallList) {
			prev_ball.NeedCheckCollision = true
		}
		if min_gap_dist > 0 {
			min_gap_dist -= DefaultBallRadius * 4
			if min_gap_dist < 0 {
				min_gap_dist = 0
			}

			var bonus_rate int32 = 500
			if curve.Board.IsEndless {
				bonus_rate = 250
			}
			gap_bonus := (MaxGapSize - min_gap_dist) * bonus_rate / MaxGapSize
			gap_bonus = (gap_bonus / 10) * 10
			if gap_bonus < 10 {
				gap_bonus = 10
			}
			if num_gaps > 1 {
				gap_bonus *= num_gaps
			}
			new_ball.GapBonus, new_ball.GapCount = gap_bonus, num_gaps
		}

		curve.Board.NumClearsInARow++
		if !curve.CheckSet(new_ball) {
			curve.Board.NumClearsInARow--

			if prev_ball != nil && !prev_ball.CollidesWithNext && prev_ball.Type == new_ball.Type &&
				prev_ball.Bullet == nil && prev_ball.ClearCount == 0 {
				new_ball.SuckPending, new_ball.SuckCount = true, 1
			} else if next_ball != nil && !new_ball.CollidesWithNext && next_ball.Type == new_ball.Type &&
				next_ball.Bullet == nil && next_ball.ClearCount == 0 {
				new_ball.SuckPending = true
				if next_ball.SuckCount <= 0 {
					next_ball.SuckCount = 1
				}
			} else {
				curve.Board.ResetInARowBonus()
				new_ball.GapBonus, new_ball.GapCount = 0, 0
			}
		}
	} else {
		*index++
	}
}

func (curve *Curve) CanFire() bool {
	if len(curve.BallList) == 0 {
		return true
	}
	return curve.BallList[len(curve.BallList)-1].WayPoint < float32(len(curve.WayPointMgr.WayPoints)-1)
}

func (curve *Curve) CheckBallIntersection(p1, v1 rl.Vector3, t *float32) *Ball {
	iter_index := 0
	var intersect_ball *Ball = nil
	for iter_index != len(curve.BallList) {
		ball := curve.BallList[iter_index]
		if !curve.WayPointMgr.InTunnel1(int(ball.WayPoint)) {
			var t2 float32 = 0
			if ball.Intersects(p1, v1, &t2) {
				if t2 < *t && t2 > 0 {
					*t = t2
					intersect_ball = ball
				}
			}
		}
		iter_index++
	}
	return intersect_ball
}

func (curve *Curve) CheckCollision(theBullet *Bullet) bool {
	bullet := theBullet
	var ball *Ball = nil
	flag := false

	for i := 0; i != len(curve.BulletList); i++ {
		bullet = curve.BulletList[i]
		if theBullet.CollidesWithPhysically(&bullet.Ball, 0) {
			bullet.Update()
			curve.AdvanceMergingBullet(&i)
			break
		}
	}

	ball_index := 0
	for ball_index = 0; ; ball_index++ {
		if ball_index == len(curve.BallList) {
			return false
		}

		ball = curve.BallList[ball_index]
		if ball.CollidesWithPhysically(&theBullet.Ball, 0) && ball.Bullet == nil && ball.ClearCount == 0 {
			prev_ball := ball.GetPrevBall(true, curve.BallList)
			if prev_ball == nil || prev_ball.Bullet == nil {
				next_ball := ball.GetNextBall(true, curve.BallList)
				if next_ball == nil || next_ball.Bullet == nil {
					v := rl.NewVector3(ball.X, ball.Y, 0)
					impliedObject := rl.NewVector3(theBullet.X, theBullet.Y, 0)
					v2 := curve.WayPointMgr.CalcPerpendicular(ball.WayPoint)

					flag = rl.Vector3CrossProduct(rl.Vector3Subtract(impliedObject, v), v2).Z < 0
					if !curve.WayPointMgr.InTunnel2(ball, flag) {
						break
					}
				}
			}
		}
	}

	if ball_index != len(curve.BallList) {
		theBullet.SetHitBall(ball, flag)
		theBullet.MergeSpeed = curve.CurveDesc.MergeSpeed

		next_ball2 := ball.GetNextBall(false, curve.BallList)
		if !flag {
			theBullet.RemoveGapInfoForBall(ball.Id)
		} else if next_ball2 != nil {
			theBullet.RemoveGapInfoForBall(next_ball2.Id)
		}
		rl.PlaySound(gSounds[Sound_BallClick2])
		curve.BulletList = append(curve.BulletList, theBullet)
		return true
	}
	return false
}

func (curve *Curve) CheckGapShot(theBullet *Bullet) bool {
	bul_radius := DefaultBallRadius
	bul_diameter := bul_radius * 2
	bul_diameter_sq := float32(bul_diameter) * float32(bul_diameter)
	bul_x, bul_y := theBullet.X, theBullet.Y
	num_way_points := curve.WayPointMgr.GetNumPoints()
	ball_idx := theBullet.GetCurCurvePoint(curve.CurveIndex)

	if ball_idx > 0 && ball_idx < num_way_points {
		way_point := &curve.WayPointMgr.WayPoints[ball_idx]
		if bul_diameter_sq > (way_point.Y-bul_y)*(way_point.Y-bul_y)+(way_point.X-bul_x)*(way_point.X-bul_x) {
			return false
		}
		theBullet.SetCurCurvePoint(curve.CurveIndex, 0)
	}

	for i := int32(1); i < num_way_points; i += bul_diameter {
		way_point := &curve.WayPointMgr.WayPoints[i]
		if !way_point.InTunnel && bul_diameter_sq > (way_point.Y-bul_y)*(way_point.Y-bul_y)+(way_point.X-bul_x)*(way_point.X-bul_x) {
			theBullet.SetCurCurvePoint(curve.CurveIndex, i)
			for k := range curve.BallList {
				ball := curve.BallList[k]
				if int32(ball.WayPoint) > i {
					prev_ball := ball.GetPrevBall(false, curve.BallList)
					if prev_ball == nil {
						return false
					}
					ball_dist := int32(ball.WayPoint - prev_ball.WayPoint)
					if ball_dist <= 0 {
						return false
					}
					return theBullet.AddGapInfo(curve.CurveIndex, ball_dist, ball.Id)
				}
			}
			return false
		}
	}

	return false
}

func (curve *Curve) CheckSet(theBall *Ball) bool {
	curve.HadPowerUp = false
	var prev_end *Ball = nil
	var next_end *Ball = nil
	combo_count := theBall.ComboCount

	count := curve.GetNumInARow(theBall, theBall.Type, &next_end, &prev_end)
	if count < 3 {
		return false
	}

	curve.Board.NumCleared = 0
	curve.Board.ClearedXSum = 0
	curve.Board.ClearedYSum = 0
	curve.Board.CurComboCount = combo_count
	curve.Board.CurComboScore = theBall.ComboScore
	curve.Board.NeedComboCount = make([]*Ball, 0)

	for i := range PowerType_Max {
		globalGotPowerUp[i] = false
	}

	var gap_bonus, num_gaps int32 = 0, 0
	end_ball, ball := next_end.GetNextBall(false, curve.BallList), prev_end
	for ball != end_ball {
		if ball.SuckPending {
			ball.SuckPending = false
			curve.Board.NumClearsInARow++
		}

		curve.StartClearCount(ball)
		gap_bonus += ball.GapBonus
		if ball.GapCount > num_gaps {
			num_gaps = ball.GapCount
		}
		ball.GapBonus, ball.GapCount = 0, 0
		ball = ball.GetNextBall(false, curve.BallList)
	}

	curve.DoScoring(theBall, curve.Board.NumCleared, combo_count, gap_bonus, num_gaps)

	if curve.Board.CurComboCount > curve.Board.LevelStats.MaxCombo ||
		curve.Board.CurComboCount == curve.Board.LevelStats.MaxCombo &&
			curve.Board.CurComboScore >= curve.Board.LevelStats.MaxComboScore {
		curve.Board.LevelStats.MaxCombo = curve.Board.CurComboCount
		curve.Board.LevelStats.MaxComboScore = curve.Board.CurComboScore
	}

	ball = prev_end
	for ball != end_ball {
		ball.ComboScore, ball.ComboCount = curve.Board.CurComboScore, combo_count
		ball = ball.GetNextBall(false, curve.BallList)
	}

	iter_index := 0
	for ; iter_index < len(curve.Board.NeedComboCount); iter_index++ {
		ball := curve.Board.NeedComboCount[iter_index]
		ball.ComboScore, ball.ComboCount = curve.Board.CurComboScore, combo_count
	}
	curve.Board.NeedComboCount = make([]*Ball, 0)

	if !curve.HadPowerUp {
		var destroy_sound rl.Sound
		switch combo_count {
		case 0:
			destroy_sound = gSounds[Sound_BallDestroyed1]
		case 1:
			destroy_sound = gSounds[Sound_BallDestroyed2]
		case 2:
			destroy_sound = gSounds[Sound_BallDestroyed3]
		case 3:
			destroy_sound = gSounds[Sound_BallDestroyed4]
		default:
			destroy_sound = gSounds[Sound_BallDestroyed5]
		}
		rl.PlaySound(destroy_sound)

		combo_sound := gSounds[Sound_Combo]
		rl.SetSoundPitch(combo_sound, float32(math.Pow(1.0594630943592952645618252949463, float64(2*combo_count))))
		rl.SetSoundVolume(combo_sound, min(1.0, float32(combo_count)*0.2+0.4))
		rl.PlaySound(combo_sound)
	}

	curve.Board.CurComboScore, curve.Board.CurComboCount = 0, 0
	return true
}

func (curve *Curve) ClearPendingSucks(theEndBall *Ball) {
	if theEndBall == nil {
		return
	}

	ball, collided := theEndBall, true
	for ball != nil {
		if ball.SuckPending {
			ball.SuckPending = false
			curve.Board.ResetInARowBonus()
			ball.GapBonus, ball.GapCount = 0, 0
		}

		ball = ball.GetPrevBall(false, curve.BallList)
		if ball == nil {
			return
		}
		if !ball.CollidesWithNext {
			collided = false
		}
		if !collided && ball.SuckCount != 0 {
			return
		}
	}
}

func (curve *Curve) DeleteBall(theBall *Ball) {
	bullet := theBall.Bullet
	if bullet != nil {
		bullet.MergeFully()
		found_index := slices.Index(curve.BulletList, bullet)
		if found_index > -1 {
			curve.AdvanceMergingBullet(&found_index)
		}
	}

	curve.DeleteBullet(bullet)
	theBall.SetCollidesWithPrev(false, curve.BallList)
	theBall.BeforeDestroy()
	theBall = nil
}

func (curve *Curve) DeleteBullet(theBullet *Bullet) {
	if theBullet == nil {
		return
	}
	found_index := slices.Index(curve.BulletList, theBullet)
	if found_index > -1 {
		curve.BulletList[found_index] = nil
		curve.BulletList = slices.Delete(curve.BulletList, found_index, found_index+1)
	}
	theBullet.BeforeDestroy()
	theBullet = nil
}

func (curve *Curve) DoScoring(theBall *Ball, theNumBalls, theComboCount, theGapBonus, theNumGaps int32) {
	if theNumBalls == 0 {
		return
	}

	text_list := make([]string, 0)
	num_points := 100*theComboCount + 10*theNumBalls + theGapBonus
	in_a_row := false
	var row_bonus int32 = 0

	if curve.Board.NumClearsInARow > 4 && theComboCount == 0 {
		row_bonus = 10*curve.Board.NumClearsInARow + 50
		num_points += row_bonus
		curve.Board.CurInARowBonus += row_bonus
		in_a_row = true
	}

	curve.Board.CurComboScore += num_points
	curve.Board.IncScore(num_points, true)

	if theComboCount > 0 {
		curve.Board.LevelStats.NumCombos++
	}
	if theGapBonus > 0 {
		curve.Board.LevelStats.NumGaps++
	}

	text_list = append(text_list, fmt.Sprintf("+%d", num_points))
	if theComboCount > 0 {
		text_list = append(text_list, fmt.Sprintf("COMBO x%d", theComboCount+1))
	}
	if theGapBonus > 0 {
		gap_string := "GAP BONUS"
		if theNumGaps > 3 {
			gap_string = fmt.Sprintf("%dx GAP BONUS", theNumGaps)
		} else if theNumGaps == 3 {
			gap_string = "TRIPLE GAP BONUS"
		} else if theNumGaps == 2 {
			gap_string = "DOUBLE GAP BONUS"
		}
		text_list = append(text_list, gap_string)
		for i := range theNumGaps {
			curve.Board.SoundMgr.AddSound(gSounds[Sound_GapBonus], i*15, 0, float32(i+1))
		}
	}

	if in_a_row {
		text_list = append(text_list, fmt.Sprintf("CHAIN BONUS x%d", curve.Board.NumClearsInARow))
		rl.SetSoundPitch(gSounds[Sound_Chain], float32(curve.Board.NumClearsInARow)-5)
		curve.Board.SoundMgr.AddSound(gSounds[Sound_Chain], 0, 0, float32(curve.Board.NumClearsInARow-5))
	}

	clr_x, clr_y := curve.Board.ClearedXSum/theNumBalls, curve.Board.ClearedYSum/theNumBalls
	if globalGotPowerUp[PowerType_SlowDown] {
		text_list = append(text_list, "SLOWDOWN Ball")
	}
	if globalGotPowerUp[PowerType_MoveBackwards] {
		text_list = append(text_list, "BACKWARDS Ball")
	}
	if globalGotPowerUp[PowerType_Accuracy] {
		text_list = append(text_list, "ACCURACY Ball")
	}
	AddTextsToMgr(text_list, FontType_Float, curve.Board.ParticleMgr, globalTextBallColors[theBall.Type], clr_x, clr_y, 0, 0)
}

func (curve *Curve) DrawBalls(theDrawer *BallDrawer) {
	for i := range curve.BallList {
		ball := curve.BallList[i]
		priority := curve.WayPointMgr.GetPriority(ball)
		next_ball := ball.GetNextBall(true, curve.BallList)

		next_priority := priority
		if next_ball != nil && priority > curve.WayPointMgr.GetPriority(next_ball) {
			next_priority = curve.WayPointMgr.GetPriority(next_ball)
		}
		num_balls := theDrawer.NumBalls[priority]
		theDrawer.NumBalls[priority]++
		theDrawer.Balls[priority][num_balls] = ball
		num_shadows := theDrawer.NumShadows[next_priority]
		theDrawer.NumShadows[next_priority]++
		theDrawer.Shadows[next_priority][num_shadows] = ball
	}
	for i := range curve.BulletList {
		bullet := curve.BulletList[i]
		priority := curve.WayPointMgr.GetPriority3(bullet)
		num_balls := theDrawer.NumBalls[priority]
		theDrawer.NumBalls[priority]++
		theDrawer.Balls[priority][num_balls] = &bullet.Ball
		num_shadows := theDrawer.NumShadows[priority]
		theDrawer.NumShadows[priority]++
		theDrawer.Shadows[priority][num_shadows] = &bullet.Ball
	}
}

func (curve *Curve) DrawPathSparkles(theStartPoint, theStagger int32, addSound bool) int32 {
	path_highlight_wp := theStartPoint
	forward_pitch := ((curve.CurveIndex ^ 1) & 1) != 0
	path_highlight_pitch := -20
	if forward_pitch {
		path_highlight_pitch = 0
	}
	var sound_ctr int32 = 0

	for path_highlight_wp < curve.WayPointMgr.GetNumPoints() {
		var sparkle_x, sparkle_y, sparkle_priority int32 = 0, 0, 0
		curve.GetPoint(path_highlight_wp, &sparkle_x, &sparkle_y, &sparkle_priority)
		curve.Board.ParticleMgr.AddSparkle(float32(sparkle_x), float32(sparkle_y), 0, 0, sparkle_priority, 0, theStagger, color.RGBA{255, 255, 0, 255})
		if addSound && sound_ctr%25 == 0 {
			if forward_pitch {
				if path_highlight_pitch > -20 {
					path_highlight_pitch--
				}
			} else {
				if path_highlight_pitch < 0 {
					path_highlight_pitch++
				}
			}
			curve.Board.SoundMgr.AddSound(gSounds[Sound_LightTrail], theStagger, 0, float32(path_highlight_pitch)*0.8)
		}
		path_highlight_wp += 11
		theStagger++
		sound_ctr++
	}

	if addSound {
		curve.Board.SoundMgr.AddSound(gSounds[Sound_LightTrailEnd], theStagger, 0, 0)
	}
	curve.SpriteMgr.AddHoleFlash(curve.CurveIndex, theStagger)

	return theStagger + 60
}

func (curve *Curve) GetCurveLength() int32 {
	return curve.WayPointMgr.GetNumPoints()
}

func (curve *Curve) GetFarthestBallPercent() int32 {
	if len(curve.BallList) == 0 {
		return 0
	}

	way_point := curve.BallList[len(curve.BallList)-1].WayPoint
	return int32(way_point * 100 / float32(curve.WayPointMgr.GetNumPoints()))
}

func (curve *Curve) GetNumInARow(theBall *Ball, theColor int32, theNextEnd, thePrevEnd **Ball) int32 {
	if theBall.Type != theColor {
		return 0
	}
	ball, color := theBall, theColor
	var count int32 = 1

	next_end := ball
	for {
		next_ball := next_end.GetNextBall(true, curve.BallList)
		if next_ball == nil || next_ball.Type != color {
			break
		}
		next_end = next_ball
		count++
	}

	prev_end := ball
	for {
		prev_ball := prev_end.GetPrevBall(true, curve.BallList)
		if prev_ball == nil || prev_ball.Type != color {
			break
		}
		prev_end = prev_ball
		count++
	}

	if theNextEnd != nil {
		*theNextEnd = next_end
	}
	if thePrevEnd != nil {
		*thePrevEnd = prev_end
	}
	return count
}

func (curve *Curve) GetNumPendingSingles(theNumGroups int32) int32 {
	var num_groups, prev_color, num_singles, group_count int32 = 0, -1, 0, 0
	index := len(curve.PendingBalls) - 1
	for index != -1 {
		ball := curve.PendingBalls[index]
		if num_groups > theNumGroups {
			break
		}
		GetNumPendingSinglesHelper(ball.Type, &num_groups, &prev_color, &num_singles, &group_count)
		index++
	}
	for i := range curve.PendingBalls {
		GetNumPendingSinglesHelper(curve.PendingBalls[i].Type, &num_groups, &prev_color, &num_singles, &group_count)
	}
	return num_singles
}

func GetNumPendingSinglesHelper(color int32, numGroups, prevColor, numSingles, groupCount *int32) {
	if color == *prevColor {
		*groupCount++
	} else {
		if *groupCount == 1 {
			*numSingles++
		}
		*groupCount = 1
		*numGroups++
		*prevColor = color
	}
}

func (curve *Curve) GetPoint(thePoint int32, x, y, pri *int32) {
	if thePoint < 0 {
		thePoint = 0
	}
	if thePoint >= curve.WayPointMgr.GetNumPoints() {
		thePoint = curve.WayPointMgr.GetEndPoint()
	}

	way_point := &curve.WayPointMgr.WayPoints[thePoint]
	*x = int32(way_point.X)
	*y = int32(way_point.Y)
	*pri = int32(way_point.Priority)
}

func (curve *Curve) HasReachedCruisingSpeed() bool {
	return curve.AdvanceSpeed-curve.CurveDesc.Speed < 0.1
}

func (curve *Curve) RemoveBallsAtFront() {
	index := 0
	for index < len(curve.BallList) {
		ball := curve.BallList[index]
		if ball.WayPoint >= 1 {
			break
		}
		index++
		curve.DeleteBullet(ball.Bullet)
		curve.BallList = slices.DeleteFunc(curve.BallList, func(b *Ball) bool { return b == ball })

		if ball.ClearCount == 0 {
			curve.Board.UpdateBallColorMap(ball, false)
		}
		if ball.ClearCount != 0 || curve.StopAddingBalls {
			curve.DeleteBall(ball)
		} else {
			curve.PendingBalls = slices.Insert(curve.PendingBalls, 0, ball)
		}
	}
}

func (curve *Curve) RollBallsIn() {
	speed := curve.CurveDesc.Speed
	var start_distance int32 = 50
	if !curve.Board.IsEndless {
		start_distance = curve.CurveDesc.StartDistance
	}

	way_point := float32(start_distance * curve.WayPointMgr.GetNumPoints() / 100)
	way_point -= float32(curve.FirstChainEnd) / float32(curve.WayPointMgr.GetNumPoints())

	if curve.FirstChainEnd <= 0 || way_point > 0 {
		curve.AdvanceSpeed = speed + ((float32(math.Sqrt(float64((speed+20)*(speed+20)+way_point*20*4)))-(speed+20))*0.5+18)*0.1
	} else {
		curve.AdvanceSpeed = curve.CurveDesc.Speed
	}
}

func (curve *Curve) SetFarthestBall(thePoint int32) {
	last_point := curve.DangerPoint
	if last_point < 0 {
		last_point = 0
	}

	var percent_open float32 = 0
	if last_point <= thePoint {
		percent_open = float32(thePoint-last_point) / float32(curve.WayPointMgr.GetNumPoints()-last_point)
	}
	curve.SpriteMgr.UpdateHole(curve.CurveIndex, percent_open)
}

func (curve *Curve) SetStopAddingBalls(stop bool) {
	if curve.StopAddingBalls == stop {
		return
	}

	if curve.GetFarthestBallPercent() > 50 {
		curve.BackwardCount = curve.CurveDesc.ZumaBack
		curve.SlowCount = curve.CurveDesc.ZumaSlow
	}

	curve.StopAddingBalls = stop
	if stop {
		for i := range curve.PendingBalls {
			curve.PendingBalls[i] = nil
		}
		curve.PendingBalls = make([]*Ball, 0)
	}
}

func (curve *Curve) SetupLevel(theDesc *LevelDesc, theSpriteMgr *SpriteMgr, theCurveIndex int32) {
	curve.LevelDesc = theDesc
	curve.CurveDesc = &theDesc.CurveDescs[theCurveIndex]
	curve.SpriteMgr = theSpriteMgr
	curve.WayPointMgr.LoadCurve(theDesc.CurveDescs[theCurveIndex].FilePath)
	curve.CurveIndex = theCurveIndex

	skull_rotation := float32(curve.CurveDesc.SkullRotation)
	if skull_rotation >= 0 {
		skull_rotation = skull_rotation * math.Pi / 180
	}

	var hole_x, hole_y int32 = 0, 0
	if len(curve.WayPointMgr.WayPoints) != 0 {
		curve.WayPointMgr.CalcPerpendicularForPoint(int(curve.WayPointMgr.GetEndPoint()))
		point := &curve.WayPointMgr.WayPoints[len(curve.WayPointMgr.WayPoints)-1]
		hole_x, hole_y = int32(point.X), int32(point.Y)
		if skull_rotation < 0 {
			skull_rotation = point.Rotation
		}
	}

	curve.SpriteMgr.PlaceHole(curve.CurveIndex, hole_x, hole_y, skull_rotation)
	curve.DangerPoint = curve.WayPointMgr.GetNumPoints() - curve.CurveDesc.DangerDistance
	if curve.DangerPoint >= curve.WayPointMgr.GetNumPoints() {
		curve.DangerPoint = curve.WayPointMgr.GetEndPoint()
	}
}

func (curve *Curve) StartClearCount(theBall *Ball) {
	if theBall.ClearCount > 0 {
		return
	}
	curve.Board.UpdateBallColorMap(theBall, false)
	curve.Board.LevelStats.NumBallsCleared++
	curve.Board.NumCleared++

	curve.Board.ClearedXSum += int32(theBall.X)
	curve.Board.ClearedYSum += int32(theBall.Y)
	curve.LastClearedBallPoint = int32(theBall.WayPoint)
	if theBall.SuckPending {
		theBall.SuckPending = false
	}

	theBall.StartClearCount(curve.WayPointMgr.InTunnel1(int(theBall.WayPoint)))
	if theBall.GetPowerTypeWussy() != PowerType_None {
		curve.Board.ActivatePower(theBall)
		curve.HadPowerUp = true
	}
}

func (curve *Curve) StartLevel() {
	curve.LevelDesc = curve.Board.LevelDesc
	curve.CurveDesc = &curve.LevelDesc.CurveDescs[curve.CurveIndex]
	curve.LastPathShowTick = curve.Board.GetTickCount() - 1000000

	for i := range PowerType_Max {
		curve.LastPowerUpFrame[i] = curve.Board.StateCount - 1000
	}

	curve.SpriteMgr.UpdateHole(curve.CurveIndex, 0)

	num_balls := curve.CurveDesc.NumBalls
	if num_balls == 0 {
		num_balls = 10
	}

	for range num_balls {
		curve.AddPendingBall()
	}

	curve.TotalBalls = curve.CurveDesc.NumBalls
	curve.RollBallsIn()
}

func (curve *Curve) UpdateBallRotation() {
	for i := range curve.BallList {
		curve.BallList[i].UpdateRotation()
	}
	for i := range curve.BulletList {
		curve.BulletList[i].UpdateRotation()
	}
}

func (curve *Curve) UpdatePlaying() {
	balls_at_beginning := len(curve.BallList) == 0 || curve.BallList[len(curve.BallList)-1].WayPoint < 50.0
	if curve.StopTime > 0 {
		curve.StopTime--
		if balls_at_beginning {
			curve.StopTime = 0
		}
		if curve.StopTime == 0 {
			curve.AdvanceSpeed = 0
		}
	}
	if curve.SlowCount > 0 {
		curve.SlowCount--
		if balls_at_beginning {
			curve.SlowCount = 0
		}
	}
	if curve.BackwardCount > 0 {
		curve.BackwardCount--
		if balls_at_beginning {
			curve.BackwardCount = 0
		}
	}

	curve.AddBall()
	curve.UpdateBallRotation()
	curve.AdvanceBullets()
	curve.UpdateSuckingBalls()
	curve.AdvanceBalls()
	curve.AdvanceBackwardBalls()
	curve.RemoveBallsAtFront()
	curve.UpdateSets()
	curve.UpdatePowerUps()
	if len(curve.BallList) == 0 {
		curve.SetFarthestBall(0)
	} else {
		curve.SetFarthestBall(int32(curve.BallList[len(curve.BallList)-1].WayPoint))
	}
}

func (curve *Curve) UpdatePowerUps() {
	if len(curve.BallList) == 0 {
		return
	}

	for i := range PowerType_Max {
		freq := curve.CurveDesc.PowerUpFreq[i]
		if freq > 0 && rand.Int31n(freq) == 0 && freq < curve.Board.StateCount-curve.LastPowerUpFrame[i] {
			curve.AddPowerUp(i)
			curve.LastPowerUpFrame[i] = curve.Board.StateCount
		}
	}
}

func (curve *Curve) UpdateSets() {
	curve.HaveSets = false
	index, len := 0, len(curve.BallList)
	for index < len {
		ball := curve.BallList[index]
		clear_count := ball.ClearCount
		if clear_count > 0 {
			curve.HaveSets = true
		}
		if clear_count < 40 {
			if clear_count > 0 {
				ball.ClearCount = clear_count + 1
			}
			index++
		} else {
			next_ball, prev_ball := ball.GetNextBall(false, curve.BallList), ball.GetPrevBall(false, curve.BallList)
			if next_ball != nil && next_ball.ClearCount == 0 && prev_ball != nil && next_ball.Type == prev_ball.Type {
				next_ball.SuckCount = 10
				next_ball.ComboScore, next_ball.ComboCount = ball.ComboScore, ball.ComboCount+1
			}
			if index == 0 {
				curve.AdvanceSpeed = 0
				if curve.StopTime < 40 {
					curve.StopTime = 40
				}
			}
			curve.DeleteBall(ball)
			curve.BallList[index] = nil
			curve.BallList = slices.Delete(curve.BallList, index, index+1)
			index++
			len--
		}
	}
}

func (curve *Curve) UpdateSuckingBalls() {
	iter_index := 0
	for iter_index != len(curve.BallList) {
		ball := curve.BallList[iter_index]
		suck_count := ball.SuckCount
		if suck_count <= 0 {
			iter_index++
			continue
		}

		var next_ball *Ball = nil
		suck := float32(suck_count / 8)
		for iter_index != len(curve.BallList) {
			next_ball = curve.BallList[iter_index]
			iter_index++
			next_ball.SuckCount = 0
			curve.WayPointMgr.SetWayPoint(next_ball, next_ball.WayPoint-suck)

			bullet := next_ball.Bullet
			if bullet != nil {
				push_ball := bullet.GetPushBall(curve.BallList)
				if push_ball != nil {
					curve.WayPointMgr.FindFreeWayPoint(push_ball, &bullet.Ball, false, 0)
				}
				bullet.UpdateHitPos()
			}
			if !next_ball.CollidesWithNext {
				break
			}
		}

		ball.SuckCount = suck_count + 1
		prev_ball := ball.GetPrevBall(false, curve.BallList)
		if prev_ball == nil {
			ball.SuckCount = 0
			continue
		}

		new_way_point := (ball.WayPoint - float32(DefaultBallRadius)) - float32(DefaultBallRadius)
		if prev_ball.WayPoint > new_way_point {
			curve.WayPointMgr.SetWayPoint(prev_ball, new_way_point)
			curve.Board.PlayBallClick(Sound_BallClick1)
			prev_ball.CollidesWithNext = true
			ball.SuckCount = 0
			if !curve.CheckSet(ball) {
				ball.ComboScore, ball.ComboCount = 0, 0
			}
			if next_ball.BackwardsCount == 0 {
				next_ball.BackwardsCount = 30
				backwards_speed := float32(ball.ComboCount) * 1.5
				if backwards_speed <= 0.5 {
					backwards_speed = 0.5
				}
				next_ball.BackwardsSpeed = backwards_speed
			}
			curve.ClearPendingSucks(next_ball)
		}
	}
}
