package main

import rl "github.com/gen2brain/raylib-go/raylib"

var gSounds map[SoundKey]rl.Sound = make(map[SoundKey]rl.Sound)

type SoundKey int32

const (
	Sound_FrogFire SoundKey = iota
	Sound_FrogSwap
	Sound_BallClick1
	Sound_BallClick2
	Sound_ExtraLife
	Sound_GapBonus
	Sound_Chain
	Sound_Combo
	Sound_BallDestroyed1
	Sound_BallDestroyed2
	Sound_BallDestroyed3
	Sound_BallDestroyed4
	Sound_BallDestroyed5
	Sound_LightTrail
	Sound_LightTrailEnd
)

func InitGlobalSounds() {
	gSounds[Sound_FrogFire] = rl.LoadSound("sounds/ballfire.ogg")
	gSounds[Sound_FrogSwap] = rl.LoadSound("sounds/ballswap.ogg")
	gSounds[Sound_BallClick1] = rl.LoadSound("sounds/ballclick1.ogg")
	rl.SetSoundVolume(gSounds[Sound_BallClick1], 0.8)
	gSounds[Sound_BallClick2] = rl.LoadSound("sounds/ballclick2.ogg")
	gSounds[Sound_ExtraLife] = rl.LoadSound("sounds/extralife.ogg")
	gSounds[Sound_GapBonus] = rl.LoadSound("sounds/gapbonus.ogg")
	gSounds[Sound_Chain] = rl.LoadSound("sounds/chain.ogg")
	gSounds[Sound_Combo] = rl.LoadSound("sounds/combo.wav")
	gSounds[Sound_BallDestroyed1] = rl.LoadSound("sounds/ballsdestroyed1.ogg")
	gSounds[Sound_BallDestroyed2] = rl.LoadSound("sounds/ballsdestroyed2.ogg")
	gSounds[Sound_BallDestroyed3] = rl.LoadSound("sounds/ballsdestroyed3.ogg")
	gSounds[Sound_BallDestroyed4] = rl.LoadSound("sounds/ballsdestroyed4.ogg")
	gSounds[Sound_BallDestroyed5] = rl.LoadSound("sounds/ballsdestroyed5.ogg")
	rl.SetSoundVolume(gSounds[Sound_BallDestroyed1], 0.8)
	rl.SetSoundVolume(gSounds[Sound_BallDestroyed2], 0.8)
	rl.SetSoundVolume(gSounds[Sound_BallDestroyed3], 0.85)
	rl.SetSoundVolume(gSounds[Sound_BallDestroyed4], 0.9)
	rl.SetSoundVolume(gSounds[Sound_BallDestroyed5], 0.95)
	gSounds[Sound_LightTrail] = rl.LoadSound("sounds/lighttrail.ogg")
	rl.SetSoundVolume(gSounds[Sound_LightTrail], 0.7)
	gSounds[Sound_LightTrailEnd] = rl.LoadSound("sounds/chant3.ogg")
}

func DestroyGlobalSounds() {
	for i := range gSounds {
		rl.UnloadSound(gSounds[i])
	}
}
