package app

import (
	"github.com/gentaman/myrogue/internal/core/component"
	"github.com/gentaman/myrogue/internal/core/content"
)

// InitializePlayer はプレイヤーのステータスをリセットし、指定された種族と属性に基づいて初期化します。
func InitializePlayer(reg *content.Registry, race component.Race, elem component.Element) {
	if len(reg.Players) == 0 {
		return
	}

	// テンプレートからリセット（player.json の値に戻す）
	reg.Players[0] = reg.PlayerTemplate
	p := &reg.Players[0]

	// 種族ボーナスの定義
	raceBonuses := map[component.Race][6]int{
		component.RaceHuman:    {0, 0, 0, 0, 0, 0},
		component.RaceElf:      {-1, 2, 1, -1, 2, 0},
		component.RaceDwarf:    {2, -1, 0, 3, -2, 0},
		component.RaceGnome:    {-1, 3, 2, -1, 0, 0},
		component.RaceHalfling: {-1, 0, 0, -1, 3, 2},
	}

	// ボーナス適用
	bonus := raceBonuses[race]
	p.Race = race
	p.Element = elem
	p.Str += bonus[0]
	p.Wis += bonus[1]
	p.Fai += bonus[2]
	p.Vit += bonus[3]
	p.Agi += bonus[4]
	p.Luk += bonus[5]

	// HP/MP の再計算（テンプレートの値をベースに加算）
	p.HP = p.HP + p.Vit*3
	p.MP = p.MP + p.Wis*2
}
