package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const DefaultBallRadius int32 = 16
const GameWidth int32 = 640
const GameHeight int32 = 480
const MaxGapSize int32 = 300
const MaxPriority int32 = 5

var globalBallBlink bool = false

var globalBoard *Board = nil

func main() {
	rl.InitAudioDevice()
	rl.InitWindow(GameWidth, GameHeight, "Zuma not Deluxe")
	defer rl.CloseWindow()
	rl.SetTargetFPS(100)

	InitGlobalTextures()
	InitGlobalSounds()
	InitFonts()
	globalBoard = NewBoard()

	globalBoard.SpriteMgr.BackgroundImage = rl.LoadTexture("levels/spiral/Spiral.jpg")
	rl.SetTextureFilter(globalBoard.SpriteMgr.BackgroundImage, rl.FilterTrilinear)
	defer rl.UnloadTexture(globalBoard.SpriteMgr.BackgroundImage)
	globalBoard.ScoreTarget = 1000
	globalBoard.TargetBarSize = 256
	globalBoard.LevelDesc.CurveDescs = make([]CurveDesc, 1)
	globalBoard.LevelDesc.CurveDescs[0] = NewCurveDesc()
	globalBoard.CurveList = make([]Curve, 1)
	globalBoard.CurveList[0] = Curve{Board: globalBoard, WayPointMgr: new(WayPointMgr), CurveIndex: 0}
	globalBoard.CurveList[0].WayPointMgr.LoadWayPoints(LoadCurveData("levels/spiral/spiral.dat"))
	globalBoard.CurveList[0].SetupLevel(globalBoard.LevelDesc, globalBoard.SpriteMgr, 0)

	globalBoard.StartLevel()
	for i := range globalBoard.CurveList {
		globalBoard.CurveList[i].StartLevel()
		globalBoard.CurveList[0].DangerPoint = globalBoard.CurveList[0].WayPointMgr.GetNumPoints() - globalBoard.CurveList[0].CurveDesc.DangerDistance
	}

	for !rl.WindowShouldClose() {
		mouse_x, mouse_y := rl.GetMouseX(), rl.GetMouseY()
		from_center_x := mouse_x - globalBoard.Frog.CenterX
		var angle float32
		if from_center_x != 0 {
			tmp := float32(math.Atan(float64(globalBoard.Frog.CenterY-mouse_y) / float64(from_center_x)))
			if from_center_x < 0 {
				tmp += math.Pi
			}
			angle = tmp + math.Pi/2.0
		} else {
			angle = 0.0
			if mouse_y < globalBoard.Frog.CenterY {
				angle = math.Pi
			}
		}
		globalBoard.Frog.SetAngle(angle)

		// Mouse Events
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			can_fire := true
			for i := range globalBoard.CurveList {
				if !globalBoard.CurveList[i].CanFire() {
					can_fire = false
					break
				}
			}
			if can_fire && globalBoard.Frog.StartFire(true) {
				rl.PlaySound(gSounds[Sound_FrogFire])
			}
		} else if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
			globalBoard.Frog.SwapBullets(true)
		}

		globalBoard.Update()

		// Update Foreground
		rl.BeginDrawing()
		globalBoard.Draw()
		rl.EndDrawing()
	}
	globalBoard.SoundMgr.Destroy()
	DestroyFontTextures()
	DestroyGlobalSounds()
	DestroyGlobalTextures()
	rl.CloseAudioDevice()
}
