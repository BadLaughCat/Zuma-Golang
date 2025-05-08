package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type SoundMgr struct {
	UpdateCount   int32
	SoundMap      map[int32]SoundDesc
	LoopingSounds [LoopType_Max]*LoopingSound
}

type SoundDesc struct {
	rl.Sound
	Pan, Pitch float32
}

func InitSoundManager() *SoundMgr {
	mgr := &SoundMgr{UpdateCount: 0, SoundMap: make(map[int32]SoundDesc)}
	for i := range LoopType_Max {
		mgr.LoopingSounds[i] = &LoopingSound{Sound: new(SoundDesc), Volume: 0}
	}
	mgr.LoopingSounds[LoopType_RollIn].Sound.Sound = rl.LoadSound("sounds/rolling.ogg")
	mgr.LoopingSounds[LoopType_RollOut].Sound.Sound = rl.LoadSound("sounds/rolling.ogg")
	return mgr
}

func (mgr *SoundMgr) Destroy() {
	rl.UnloadSound(mgr.LoopingSounds[LoopType_RollIn].Sound.Sound)
	rl.UnloadSound(mgr.LoopingSounds[LoopType_RollOut].Sound.Sound)
}

func (mgr *SoundMgr) AddSound(theSound rl.Sound, theDelay, thePan int32, thePitchShift float32) {
	tmp := SoundDesc{theSound, float32(thePan), thePitchShift}
	if theDelay == 0 {
		mgr.PlaySample(tmp)
	} else {
		mgr.SoundMap[mgr.UpdateCount+theDelay] = tmp
	}
}

func (mgr SoundMgr) PlaySample(theDesc SoundDesc) {
	if theDesc.Pan != 0 {
		rl.SetSoundPan(theDesc.Sound, theDesc.Pan)
	}
	if theDesc.Pitch != 0 {
		rl.SetSoundPitch(theDesc.Sound, float32(math.Pow(1.0594630943592952645618252949463, float64(theDesc.Pitch))))
	}
	rl.PlaySound(theDesc.Sound)
}

func (mgr SoundMgr) PlayLoop(theSound LoopType) {
	mgr.LoopingSounds[theSound].Play()
}

func (mgr SoundMgr) StopLoop(theSound LoopType) {
	if mgr.LoopingSounds[theSound].Volume > 0.99 {
		mgr.LoopingSounds[theSound].Volume = 0.98
	}
}

func (mgr *SoundMgr) Update() {
	mgr.UpdateCount++
	for i := range mgr.SoundMap {
		if i <= mgr.UpdateCount {
			mgr.PlaySample(mgr.SoundMap[i])
			delete(mgr.SoundMap, i)
		}
	}
	for i := range LoopType_Max {
		mgr.LoopingSounds[i].Update()
	}
}

type LoopType int32

const (
	LoopType_RollIn LoopType = iota
	LoopType_RollOut
	LoopType_Max
)

type LoopingSound struct {
	Sound     *SoundDesc
	Volume    float32
	IsPlaying bool
}

func (loops *LoopingSound) Play() {
	if loops.Sound == nil {
		return
	}
	loops.IsPlaying = true
	loops.Volume = 1
	rl.SetSoundVolume(loops.Sound.Sound, 1)
	go func() {
		for {
			if !loops.IsPlaying {
				return
			}
			if !rl.IsSoundPlaying(loops.Sound.Sound) {
				rl.PlaySound(loops.Sound.Sound)
			}
		}
	}()
}

func (loops *LoopingSound) Update() {
	if loops.Sound == nil || !loops.IsPlaying {
		return
	}
	if loops.Volume < 1 {
		loops.Volume -= 0.02
		if loops.Volume > 0 {
			rl.SetSoundVolume(loops.Sound.Sound, loops.Volume)
		} else {
			loops.IsPlaying = false
		}
	}
}
