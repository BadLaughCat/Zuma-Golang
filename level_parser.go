package main

import (
	"os"
	"strconv"

	"github.com/tidwall/gjson"
)

type LevelParser struct {
	GraphicsMap map[string]LevelDesc
	SettingsMap map[string]LevelDescModify
}

func NewLevelParser() LevelParser {
	return LevelParser{
		GraphicsMap: make(map[string]LevelDesc),
		SettingsMap: make(map[string]LevelDescModify),
	}
}

func (parser *LevelParser) ParseLevels(filePath string) {
	raw, _ := os.ReadFile(filePath)
	json := string(raw)

	{
		graphic_list := gjson.Get(json, "Graphics").Array()
		for i := range graphic_list {
			obj := graphic_list[i].Map()
			id := obj["id"].String()
			desc := NewLevelDesc()
			desc.FrogX = int32(obj["frogx"].Int())
			desc.FrogY = int32(obj["frogy"].Int())
			TryGetAndSet(obj, "space", func(r gjson.Result) { desc.IsInSpace = r.Bool() })
			TryGetAndSet(obj, "dispname", func(r gjson.Result) { desc.DisplayName = r.String() })

			curve_ids := obj["curves"].Array()
			desc.CurveDescs = make([]CurveDesc, len(curve_ids))
			for k := range curve_ids {
				desc.CurveDescs[k] = NewCurveDesc()
				desc.CurveDescs[k].FilePath = "./levels/" + id + "/" + curve_ids[k].String() + ".dat"
			}
			TryGetAndSet(obj, "skullrot", func(r gjson.Result) {
				for k := range desc.CurveDescs {
					desc.CurveDescs[k].SkullRotation = int32(r.Int())
				}
			})

			TryGetAndSet(obj, "image", func(r gjson.Result) { desc.ImagePath = r.String() })
			TryGetAndSet(obj, "TreasurePoints", func(r gjson.Result) {
				treasure_points := r.Array()
				desc.TreasurePoints = make([]TreasurePoint, len(treasure_points))
				for k := range treasure_points {
					iter_item := treasure_points[k].Map()
					the_point := &desc.TreasurePoints[k]
					the_point.X = int32(iter_item["x"].Int())
					the_point.Y = int32(iter_item["y"].Int())
					delete(iter_item, "x")
					delete(iter_item, "y")
					// dist1 dist2
					for m := range iter_item {
						index, _ := strconv.Atoi(m[4:])
						widen_for_index(&the_point.CurveDist, index)
						the_point.CurveDist[index-1] = int32(iter_item[m].Int())
					}
				}
			})
			TryGetAndSet(obj, "Cutouts", func(r gjson.Result) {
				cutouts := r.Array()
				for k := range cutouts {
					iter_item := cutouts[k].Map()
					sprite := SpriteDesc{IsCutout: true}
					sprite.ImagePath = "./levels/" + id + "/" + iter_item["image"].String() + ".png"
					sprite.X = int32(iter_item["x"].Int())
					sprite.Y = int32(iter_item["y"].Int())
					sprite.Priority = int32(iter_item["pri"].Int())
					desc.Sprites = append(desc.Sprites, sprite)
				}
			})
			TryGetAndSet(obj, "BackgroundAlphas", func(r gjson.Result) {
				bg_alphas := r.Array()
				for k := range bg_alphas {
					iter_item := bg_alphas[k].Map()
					sprite := SpriteDesc{}
					sprite.ImagePath = iter_item["image"].String()
					sprite.X = int32(iter_item["x"].Int())
					sprite.Y = int32(iter_item["y"].Int())
					TryGetAndSet(obj, "vx", func(r gjson.Result) { sprite.VX = float32(r.Float()) })
					TryGetAndSet(obj, "vy", func(r gjson.Result) { sprite.VY = float32(r.Float()) })
					desc.BackgroundAlphas = append(desc.BackgroundAlphas, sprite)
				}
			})
			parser.GraphicsMap[id] = *desc
		}
	}
	{
		settings_list := gjson.Get(json, "Settings").Array()
		for i := range settings_list {
			obj := settings_list[i].Map()
			desc := &LevelDescModify{}
			desc.CurveDesc = NewCurveDesc()
			curve_desc := &desc.CurveDesc

			TryGetAndSet(obj, "colors", func(r gjson.Result) { curve_desc.NumColors = int32(r.Int()) })
			TryGetAndSet(obj, "firespeed", func(r gjson.Result) { desc.FireSpeed = new(float32); *desc.FireSpeed = float32(r.Float()) })
			TryGetAndSet(obj, "mergespeed", func(r gjson.Result) { curve_desc.MergeSpeed = float32(r.Float()) })
			TryGetAndSet(obj, "partime", func(r gjson.Result) { desc.ParTime = int32(r.Int()) })
			TryGetAndSet(obj, "powerfreq", func(r gjson.Result) {
				for k := range PowerType_Max {
					if curve_desc.PowerUpFreq[k] > 0 {
						curve_desc.PowerUpFreq[k] = int32(r.Int())
					}
				}
			})
			TryGetAndSet(obj, "reloaddelay", func(r gjson.Result) { desc.ReloadDelay = new(int32); *desc.ReloadDelay = int32(r.Int()) })
			TryGetAndSet(obj, "repeat", func(r gjson.Result) { curve_desc.BallRepeat = int32(r.Int()) })
			TryGetAndSet(obj, "score", func(r gjson.Result) { curve_desc.ScoreTarget = int32(r.Int()) })
			TryGetAndSet(obj, "single", func(r gjson.Result) { curve_desc.MaxSingle = int32(r.Int()) })
			TryGetAndSet(obj, "slowfactor", func(r gjson.Result) { curve_desc.SlowFactor = float32(r.Float()) })
			TryGetAndSet(obj, "speed", func(r gjson.Result) { curve_desc.Speed = float32(r.Float()) })
			TryGetAndSet(obj, "start", func(r gjson.Result) { curve_desc.StartDistance = int32(r.Int()) })
			TryGetAndSet(obj, "treasurefreq", func(r gjson.Result) { desc.TreasureFreq = new(int32); *desc.TreasureFreq = int32(r.Int()) })
			TryGetAndSet(obj, "zumaback", func(r gjson.Result) { curve_desc.ZumaBack = int32(r.Int()) })
			TryGetAndSet(obj, "zumaslow", func(r gjson.Result) { curve_desc.ZumaSlow = int32(r.Int()) })

			if desc.ParTime == 0 {
				desc.ParTime = 35 * curve_desc.ScoreTarget / 1000
				desc.ParTime = 5 * ((desc.ParTime + 4) / 5)
			}

			parser.SettingsMap[obj["id"].String()] = *desc
		}
	}
}

func TryGetAndSet(jobject map[string]gjson.Result, key string, invoke func(gjson.Result)) {
	if result, found := jobject[key]; found {
		invoke(result)
	}
}
