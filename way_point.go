package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type WayPoint struct {
	PathPoint
	HasPerpendicular      bool
	Perpendicular         rl.Vector3
	HaveAvgRotation       bool
	Rotation, AvgRotation float32
}

type WayPointMgr struct {
	WayPoints []WayPoint
}

func (mgr *WayPointMgr) CalcAvgRotationForPoint(theWayPoint int) {
	target := &mgr.WayPoints[theWayPoint]
	if target.HaveAvgRotation {
		return
	}

	mgr.CalcPerpendicularForPoint(theWayPoint)
	target.HaveAvgRotation = true
	target.AvgRotation = target.Rotation

	first, last := theWayPoint-10, theWayPoint+10
	if first < 0 {
		first = 0
	}
	if last >= len(mgr.WayPoints) {
		last = len(mgr.WayPoints) - 1
	}

	mgr.CalcPerpendicularForPoint(first)

	for i := first + 1; i < last; i++ {
		mgr.CalcPerpendicularForPoint(i)

		dr := mgr.WayPoints[i].Rotation - mgr.WayPoints[i-1].Rotation
		if dr > 0 {
			for dr > math.Pi {
				dr = dr - 2*math.Pi
			}
		} else if dr < 0 {
			for dr < -math.Pi {
				dr = dr + 2*math.Pi
			}
		}

		if dr > 0.1 || dr < -0.1 {
			mix := 1 - float32(i-first)/float32(last-first)
			target.AvgRotation = mgr.WayPoints[first].Rotation + mix*dr
			return
		}
	}
}

func (mgr *WayPointMgr) CalcPerpendicular(theWayPoint float32) rl.Vector3 {
	way_point := int(theWayPoint)
	if way_point < 0 {
		way_point = 0
	}
	if way_point >= len(mgr.WayPoints) {
		way_point = len(mgr.WayPoints) - 1
	}
	mgr.CalcPerpendicularForPoint(way_point)
	return mgr.WayPoints[way_point].Perpendicular
}

func (mgr *WayPointMgr) CalcPerpendicularForPoint(theWayPoint int) {
	target := &mgr.WayPoints[theWayPoint]
	if target.HasPerpendicular {
		return
	}

	target2 := target
	if theWayPoint+1 < len(mgr.WayPoints) {
		target2 = &mgr.WayPoints[theWayPoint+1]
	} else {
		target2 = &mgr.WayPoints[theWayPoint-1]
	}

	target.Perpendicular = rl.Vector3Normalize(rl.NewVector3(target2.Y-target.Y, target.X-target2.X, 0))
	target.Rotation = float32(math.Acos(float64(rl.Vector3DotProduct(target.Perpendicular, rl.NewVector3(1, 0, 0)))))

	if target.Perpendicular.Y > 0 {
		target.Rotation = -target.Rotation
	}
	if target.Rotation < 0 {
		target.Rotation += 2 * math.Pi
	}
	target.HasPerpendicular = true
}

func (mgr *WayPointMgr) GetEndPoint() int32 {
	return int32(len(mgr.WayPoints)) - 1
}

func (mgr *WayPointMgr) GetNumPoints() int32 {
	return int32(len(mgr.WayPoints))
}

func (mgr *WayPointMgr) FindFreeWayPoint(existBall, newBall *Ball, inFront bool, thePad int32) {
	itr, way_point := 1, int(existBall.WayPoint)
	if !inFront {
		itr = -1
	}
	if inFront && existBall.WayPoint < newBall.WayPoint {
		way_point = int(newBall.WayPoint)
	} else if existBall.WayPoint > newBall.WayPoint {
		way_point = int(newBall.WayPoint)
	}

	for way_point >= 0 {
		if way_point >= len(mgr.WayPoints) {
			break
		}
		way_point2 := &mgr.WayPoints[way_point]
		newBall.X, newBall.Y = way_point2.X, way_point2.Y
		if !newBall.CollidesWithPhysically(existBall, thePad) {
			break
		}
		way_point += itr
	}
	mgr.SetWayPointInt(newBall, way_point)
}

func (mgr *WayPointMgr) GetPriority(theBall *Ball) uint8 {
	radius := float32(DefaultBallRadius)
	prev_priority := mgr.GetPriority2(int(theBall.WayPoint - radius))
	next_priority := mgr.GetPriority2(int(theBall.WayPoint + radius))
	return max(prev_priority, next_priority)
}

func (mgr *WayPointMgr) GetPriority2(thePoint int) uint8 {
	if thePoint < 0 || thePoint >= len(mgr.WayPoints) {
		return 0
	}
	return mgr.WayPoints[thePoint].Priority
}

func (mgr *WayPointMgr) GetPriority3(theBullet *Bullet) uint8 {
	return mgr.GetPriority(&theBullet.Ball)
}

func (mgr *WayPointMgr) GetRotationForPoint(theWayPoint int) float32 {
	way_point := theWayPoint
	if way_point < 0 {
		way_point = 0
	}
	if way_point >= len(mgr.WayPoints)-1 {
		way_point = len(mgr.WayPoints) - 1
	}
	mgr.CalcPerpendicularForPoint(way_point)
	return mgr.WayPoints[way_point].Rotation
}

func (mgr *WayPointMgr) InTunnel1(theWayPoint int) bool {
	if theWayPoint < 0 {
		return true
	}
	if theWayPoint >= len(mgr.WayPoints) {
		return false
	}
	return mgr.WayPoints[theWayPoint].InTunnel
}

func (mgr *WayPointMgr) InTunnel2(theBall *Ball, inFront bool) bool {
	way_point := int(theBall.WayPoint)
	if inFront {
		return mgr.InTunnel1(way_point + int(DefaultBallRadius))
	} else {
		return mgr.InTunnel1(way_point - int(DefaultBallRadius))
	}
}

func (mgr *WayPointMgr) LoadCurve(filePath string) {
	path_points := LoadCurveData(filePath)
	for i := range path_points {
		mgr.WayPoints = append(mgr.WayPoints, WayPoint{
			HasPerpendicular: false,
			Perpendicular:    rl.Vector3Zero(),
			HaveAvgRotation:  false,
			Rotation:         0,
			AvgRotation:      0,
			PathPoint: PathPoint{
				X:        path_points[i].X,
				Y:        path_points[i].Y,
				InTunnel: path_points[i].InTunnel,
				Priority: path_points[i].Priority,
			},
		})
	}
}

func (mgr *WayPointMgr) SetWayPoint(ball *Ball, thePoint float32) {
	if len(mgr.WayPoints) == 0 {
		return
	}

	point := int(thePoint)
	next_point := 0
	if thePoint < 0 {
		point, next_point = 0, 1
	} else if point >= len(mgr.WayPoints) {
		point = len(mgr.WayPoints) - 1
		next_point = point + 1
	} else {
		point = int(thePoint)
		next_point = point + 1
	}

	way_point := &mgr.WayPoints[point]
	next_way_point := way_point
	if next_point < len(mgr.WayPoints) {
		next_way_point = &mgr.WayPoints[next_point]
	}
	mix := thePoint - float32(point)

	mgr.CalcAvgRotationForPoint(point)
	ball.X = way_point.X + mix*(next_way_point.X-way_point.X)
	ball.Y = way_point.Y + mix*(next_way_point.Y-way_point.Y)
	ball.WayPoint = thePoint
	ball.SetRotation(way_point.AvgRotation, false)
}

func (mgr *WayPointMgr) SetWayPointInt(ball *Ball, thePoint int) {
	if len(mgr.WayPoints) == 0 {
		return
	}
	point := 0
	if thePoint >= 0 {
		if thePoint >= len(mgr.WayPoints) {
			point = len(mgr.WayPoints) - 1
		} else {
			point = thePoint
		}
	}

	way_point := &mgr.WayPoints[point]
	mgr.CalcAvgRotationForPoint(point)
	ball.X, ball.Y = way_point.X, way_point.Y
	ball.WayPoint = float32(thePoint)
	ball.SetRotation(way_point.AvgRotation, false)
}
