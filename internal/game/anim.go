package game

const (
	attackAnimFrames = 8  // 攻撃アニメの総フレーム数
	idlePulsePeriod  = 30 // 待機パルスの半周期（フレーム）
)

// charAnim はキャラクターの描画オフセットとサイズを返す。
// attackAnim: 残り攻撃アニメフレーム数（0なら待機）
func charAnim(frame, attackAnim int, d Dir) (offsetX, offsetY, size int) {
	const normalSize = tileSize - 1 // 15
	const smallSize  = tileSize - 3 // 13

	if attackAnim > 0 {
		// 前半：向き方向に3px前進、後半：1pxに戻る
		shift := 1
		if attackAnim > attackAnimFrames/2 {
			shift = 3
		}
		dx, dy := d.delta()
		return dx * shift, dy * shift, normalSize
	}

	// 待機パルス：idlePulsePeriod フレームごとにサイズ交互
	if (frame/idlePulsePeriod)%2 == 0 {
		return 0, 0, normalSize
	}
	return 1, 1, smallSize
}
