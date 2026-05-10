# My Rogue

Ebitenで作るブラウザ向けローグライクゲーム（Go + WebAssembly）。

詳細な仕様については [GAME.md](./GAME.md) を参照してください。

## 遊び方

- 矢印キー or WASD: 移動
- 宝（金色タイル）を見つけて、入り口の階段（水色タイル）に戻れば脱出成功
- Rキー: クリア後にリスタート

## 開発

### 必要なもの

- Go 1.26+
- Python 3（開発サーバー用）

### コマンド

```bash
make build            # WASMファイルをビルド
make build-debug      # デバッグ情報付きでビルド
make serve            # ビルドして開発サーバーを起動 (http://localhost:8080)
make clean            # ビルド成果物を削除
make update-wasm-exec # wasm_exec.jsを最新版に更新（Goバージョン変更時）
```
