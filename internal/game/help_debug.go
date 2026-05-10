//go:build debug

package game

import "image/color"

func (s *HelpScene) addDebugHelp() {
	s.content = append(s.content, []helpLine{
		{"【デバッグ操作】", color.RGBA{255, 100, 100, 255}, 1},
		{"F1: 詳細デバッグ画面", color.White, 1},
		{"[: アニメーション速度 -", color.White, 1},
		{"]: アニメーション速度 +", color.White, 1},
		{"P: アニメーション速度リセット", color.White, 1},
	}...)
}
