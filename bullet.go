package main

import "slices"

type Bullet struct {
	Ball
	VelX, VelY                 float32
	HitBall                    *Ball
	HitInFront, HasSetPrevBall bool
	HitX, HitY                 float32
	DestX, DestY               float32
	HitPercent                 float32
	MergeSpeed                 float32
	GapInfos                   []*GapInfo
	CurCurvePoint              []int32
}

type GapInfo struct {
	CurveIndex, Dist, Id int32
}

func NewBullet() *Bullet {
	tmp := &Bullet{}
	tmp.Ball = *NewBall()
	tmp.MergeSpeed = 0.05
	return tmp
}

func (bullet *Bullet) BeforeDestroy() {
	bullet.Ball.BeforeDestroy()
	bullet.SetBallInfo(nil)
}

func (bullet *Bullet) AddGapInfo(theCurve, theDist, theBallId int32) bool {
	for i := range bullet.GapInfos {
		if bullet.GapInfos[i].Id == theBallId {
			return false
		}
	}
	bullet.GapInfos = append(bullet.GapInfos, &GapInfo{theCurve, theDist, theBallId})
	return true
}

func (bullet *Bullet) CheckSetHitBallToPrevBall(ballList []*Ball) {
	if bullet.HasSetPrevBall || bullet.HitBall == nil {
		return
	}
	prev_ball := bullet.HitBall.GetPrevBall(false, ballList)
	if prev_ball == nil {
		return
	}
	if prev_ball.CollidesWithPhysically(&bullet.Ball, 0) && prev_ball.ClearCount == 0 {
		bullet.HasSetPrevBall = true
		bullet.SetBallInfo(nil)
		bullet.HitBall = prev_ball
		bullet.HitInFront = true
		bullet.HitX = bullet.X
		bullet.HitY = bullet.Y
		bullet.HitPercent = 0
		bullet.SetBallInfo(bullet)
	}
}

func (bullet *Bullet) Draw() {
	way_point := bullet.WayPoint
	bullet.WayPoint = 0
	bullet.Ball.Draw()
	bullet.WayPoint = way_point
}

func (bullet *Bullet) GetCurCurvePoint(theCurveNum int32) int32 {
	return bullet.CurCurvePoint[theCurveNum]
}

func (bullet *Bullet) GetMinGapDist() int32 {
	var min_dist int32 = 0
	for i := range bullet.GapInfos {
		if min_dist == 0 || bullet.GapInfos[i].Dist < min_dist {
			min_dist = bullet.GapInfos[i].Dist
		}
	}
	return min_dist
}

func (bullet *Bullet) GetPushBall(list []*Ball) *Ball {
	if bullet.HitBall == nil {
		return nil
	}
	push_ball := bullet.HitBall
	if bullet.HitInFront {
		push_ball = bullet.HitBall.GetNextBall(false, list)
	}
	if push_ball != nil && push_ball.CollidesWithPhysically(&bullet.Ball, 0) {
		return push_ball
	}
	return nil
}

func (bullet *Bullet) MergeFully() {
	bullet.HitPercent = 1
	bullet.Update()
}

func (bullet *Bullet) RemoveGapInfoForBall(theBallId int32) {
	for i := 0; i != len(bullet.GapInfos); {
		if bullet.GapInfos[i].Id == theBallId {
			bullet.GapInfos[i] = nil
			bullet.GapInfos = slices.Delete(bullet.GapInfos, i, i+1)
		} else {
			i++
		}
	}
}

func (bullet *Bullet) SetBallInfo(theBullet *Bullet) {
	if bullet.HitBall != nil {
		bullet.HitBall.Bullet = theBullet
	}
}

func (bullet *Bullet) SetCurCurvePoint(theCurveNum, thePoint int32) {
	bullet.CurCurvePoint[theCurveNum] = thePoint
}

func (bullet *Bullet) SetHitBall(theBall *Ball, hitInFront bool) {
	bullet.SetBallInfo(nil)
	bullet.HasSetPrevBall = false
	bullet.HitBall = theBall
	bullet.HitX = bullet.X
	bullet.HitY = bullet.Y
	bullet.HitPercent = 0
	bullet.HitInFront = hitInFront
	bullet.SetBallInfo(bullet)
}

func (bullet *Bullet) Update() {
	if bullet.HitBall == nil {
		bullet.X += bullet.VelX
		bullet.Y += bullet.VelY
	} else if bullet.ClearCount == 0 {
		bullet.HitPercent += bullet.MergeSpeed
		if bullet.HitPercent > 1 {
			bullet.HitPercent = 1
		}
		bullet.X = bullet.HitX + bullet.HitPercent*(bullet.DestX-bullet.HitX)
		bullet.Y = bullet.HitY + bullet.HitPercent*(bullet.DestY-bullet.HitY)
	}
}

func (bullet *Bullet) UpdateHitPos() {
	bullet.HitX, bullet.HitY = bullet.X, bullet.Y
}
