.PHONY: build serve clean update-wasm-exec help

# デフォルトターゲット
.DEFAULT_GOAL := help

# WASMビルド
build:
	GOOS=js GOARCH=wasm go build -o game.wasm main.go

# 開発サーバー起動（ビルド後に自動起動）
serve: build
	@echo "開発サーバーを起動します: http://localhost:8080"
	python3 -m http.server 8080

# wasm_exec.jsの更新（Goバージョン変更時）
update-wasm-exec:
	@if [ -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" .; \
		echo "wasm_exec.jsを更新しました (misc/wasm/)"; \
	elif [ -f "$$(go env GOROOT)/lib/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" .; \
		echo "wasm_exec.jsを更新しました (lib/wasm/)"; \
	else \
		echo "エラー: wasm_exec.jsが見つかりません"; \
		exit 1; \
	fi

# クリーンアップ
clean:
	rm -f game.wasm

# ヘルプ
help:
	@echo "使用可能なコマンド:"
	@echo "  make build            - WASMファイルをビルド"
	@echo "  make serve            - ビルドして開発サーバーを起動"
	@echo "  make clean            - ビルド成果物を削除"
	@echo "  make update-wasm-exec - wasm_exec.jsを最新版に更新"
