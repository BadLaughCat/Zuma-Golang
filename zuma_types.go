package main

type LevelDesc struct {
	Id                                         int32
	Name, DisplayName, ImagePath               string
	FireSpeed                                  float32
	ReloadDelay                                int32
	FrogX, FrogY                               int32
	BGColor, Difficulty, TreasureFreq, ParTime int32
	IsInSpace                                  bool
	Stage, Level                               int32
	CurveDescs                                 []CurveDesc
}

func NewLevelDesc() *LevelDesc {
	return &LevelDesc{
		FrogX:        320,
		FrogY:        240,
		FireSpeed:    6.0,
		BGColor:      -1,
		TreasureFreq: 300,
	}
}

type CurveDesc struct {
	DangerDistance     int32
	StartDistance      int32
	Speed              float32
	MergeSpeed         float32
	MaxSpeed           float32
	NumBalls           int32
	BallRepeat         int32
	MaxSingle          int32
	NumColors          int32
	SlowFactor         float32
	SlowSpeed          float32
	SlowDistance       int32
	AccelerationRate   float32
	CurAcceleration    float32
	ScoreTarget        int32
	SkullRotation      int32
	ZumaBack, ZumaSlow int32
	PowerUpFreq        [PowerType_Max]int32
}

func NewCurveDesc() CurveDesc {
	var powerup_freq [PowerType_Max]int32
	for i := range PowerType_Max {
		powerup_freq[i] = 3000
	}
	return CurveDesc{
		DangerDistance:   600,
		StartDistance:    40,
		Speed:            0.5,
		MergeSpeed:       0.05,
		MaxSpeed:         100,
		NumBalls:         0,
		BallRepeat:       40,
		MaxSingle:        10,
		NumColors:        4,
		SlowFactor:       4,
		SlowSpeed:        0,
		SlowDistance:     500,
		AccelerationRate: 0,
		CurAcceleration:  0,
		ScoreTarget:      1000,
		SkullRotation:    -1,
		ZumaBack:         300,
		ZumaSlow:         1100,
		PowerUpFreq:      powerup_freq,
	}
}
