//go:build !debug

package debug

// Enabled はデバッグモードが有効かどうかを示します。
// ビルドタグ 'debug' が指定されていない場合に false になります。
const Enabled = false
