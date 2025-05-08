package main

import rl "github.com/gen2brain/raylib-go/raylib"

var gTextures map[TextureKey]rl.Texture2D = make(map[TextureKey]rl.Texture2D)

type TextureKey int32

const (
	Texture_BlueBall TextureKey = iota
	Texture_YellowBall
	Texture_RedBall
	Texture_GreenBall
	Texture_PurpleBall
	Texture_WhiteBall

	Texture_BallDots
	Texture_BallExplosion
	Texture_BallShadow
	Texture_FrogBase
	Texture_FrogImageMask
	Texture_FrogEye
	Texture_FrogTongue
	Texture_Sparkle
	Texture_Explosion

	Texture_AccuracyLight
	Texture_BackwardsLight
	Texture_SlowLight

	Texture_BlueLight
	Texture_YellowLight
	Texture_RedLight
	Texture_GreenLight
	Texture_PurpleLight
	Texture_WhiteLight

	Texture_BlueAccuracy
	Texture_YellowAccuracy
	Texture_RedAccuracy
	Texture_GreenAccuracy
	Texture_PurpleAccuracy
	Texture_WhiteAccuracy

	Texture_BlueBackwards
	Texture_YellowBackwards
	Texture_RedBackwards
	Texture_GreenBackwards
	Texture_PurpleBackwards
	Texture_WhiteBackwards

	Texture_BlueBomb
	Texture_YellowBomb
	Texture_RedBomb
	Texture_GreenBomb
	Texture_PurpleBomb
	Texture_WhiteBomb

	Texture_BlueSlow
	Texture_YellowSlow
	Texture_RedSlow
	Texture_GreenSlow
	Texture_PurpleSlow
	Texture_WhiteSlow

	Texture_Hole
	Texture_HoleCover
	Texture_Life
)

func InitGlobalTextures() {
	gTextures[Texture_BlueBall] = rl.LoadTexture("images/baBallBlue.png")
	gTextures[Texture_YellowBall] = rl.LoadTexture("images/baBallYellow.png")
	gTextures[Texture_RedBall] = rl.LoadTexture("images/baBallRed.png")
	gTextures[Texture_GreenBall] = rl.LoadTexture("images/baBallGreen.png")
	gTextures[Texture_PurpleBall] = rl.LoadTexture("images/baBallPurple.png")
	gTextures[Texture_WhiteBall] = rl.LoadTexture("images/baBallWhite.png")

	gTextures[Texture_BallDots] = rl.LoadTexture("images/baDotz.png")
	gTextures[Texture_BallExplosion] = rl.LoadTexture("images/grayplosion.png")
	gTextures[Texture_BallShadow] = rl.LoadTexture("images/ballshadow.png")
	gTextures[Texture_FrogBase] = rl.LoadTexture("images/SMALLFROGonPAD.png")
	gTextures[Texture_FrogImageMask] = rl.LoadTexture("images/mask.png")
	gTextures[Texture_FrogEye] = rl.LoadTexture("images/EYEBLINK.png")
	gTextures[Texture_FrogTongue] = rl.LoadTexture("images/Tongue.png")
	gTextures[Texture_Sparkle] = rl.LoadTexture("images/sparkle.png")
	gTextures[Texture_Explosion] = rl.LoadTexture("images/Explosion.png")

	gTextures[Texture_AccuracyLight] = rl.LoadTexture("images/baAccuracyLight.png")
	gTextures[Texture_BackwardsLight] = rl.LoadTexture("images/baBackwardsLight.png")
	gTextures[Texture_SlowLight] = rl.LoadTexture("images/baSlowLight.png")

	gTextures[Texture_BlueLight] = rl.LoadTexture("images/baLightBlue.png")
	gTextures[Texture_YellowLight] = rl.LoadTexture("images/baLightYellow.png")
	gTextures[Texture_RedLight] = rl.LoadTexture("images/baLightRed.png")
	gTextures[Texture_GreenLight] = rl.LoadTexture("images/baLightGreen.png")
	gTextures[Texture_PurpleLight] = rl.LoadTexture("images/baLightPurple.png")
	gTextures[Texture_WhiteLight] = rl.LoadTexture("images/baLightWhite.png")

	gTextures[Texture_BlueAccuracy] = rl.LoadTexture("images/baAccuracyBlue.png")
	gTextures[Texture_YellowAccuracy] = rl.LoadTexture("images/baAccuracyYellow.png")
	gTextures[Texture_RedAccuracy] = rl.LoadTexture("images/baAccuracyRed.png")
	gTextures[Texture_GreenAccuracy] = rl.LoadTexture("images/baAccuracyGreen.png")
	gTextures[Texture_PurpleAccuracy] = rl.LoadTexture("images/baAccuracyPurple.png")
	gTextures[Texture_WhiteAccuracy] = rl.LoadTexture("images/baAccuracyWhite.png")

	gTextures[Texture_BlueBackwards] = rl.LoadTexture("images/baBackwardsBlue.png")
	gTextures[Texture_YellowBackwards] = rl.LoadTexture("images/baBackwardsYellow.png")
	gTextures[Texture_RedBackwards] = rl.LoadTexture("images/baBackwardsRed.png")
	gTextures[Texture_GreenBackwards] = rl.LoadTexture("images/baBackwardsGreen.png")
	gTextures[Texture_PurpleBackwards] = rl.LoadTexture("images/baBackwardsPurple.png")
	gTextures[Texture_WhiteBackwards] = rl.LoadTexture("images/baBackwardsWhite.png")

	gTextures[Texture_BlueBomb] = rl.LoadTexture("images/baBombBlue.png")
	gTextures[Texture_YellowBomb] = rl.LoadTexture("images/baBombYellow.png")
	gTextures[Texture_RedBomb] = rl.LoadTexture("images/baBombRed.png")
	gTextures[Texture_GreenBomb] = rl.LoadTexture("images/baBombGreen.png")
	gTextures[Texture_PurpleBomb] = rl.LoadTexture("images/baBombPurple.png")
	gTextures[Texture_WhiteBomb] = rl.LoadTexture("images/baBombWhite.png")

	gTextures[Texture_BlueSlow] = rl.LoadTexture("images/baSlowBlue.png")
	gTextures[Texture_YellowSlow] = rl.LoadTexture("images/baSlowYellow.png")
	gTextures[Texture_RedSlow] = rl.LoadTexture("images/baSlowRed.png")
	gTextures[Texture_GreenSlow] = rl.LoadTexture("images/baSlowGreen.png")
	gTextures[Texture_PurpleSlow] = rl.LoadTexture("images/baSlowPurple.png")
	gTextures[Texture_WhiteSlow] = rl.LoadTexture("images/baSlowWhite.png")

	gTextures[Texture_Hole] = rl.LoadTexture("images/Hole.png")
	gTextures[Texture_HoleCover] = rl.LoadTexture("images/pitcover.png")
	gTextures[Texture_Life] = rl.LoadTexture("images/Life.png")
	for i := range gTextures {
		rl.SetTextureFilter(gTextures[i], rl.FilterTrilinear)
	}
}

func DestroyGlobalTextures() {
	for i := range gTextures {
		rl.UnloadTexture(gTextures[i])
	}
}
