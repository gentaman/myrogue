# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

Ebitenを使ったブラウザ向けローグライクゲーム。Go + WebAssemblyで動作する。

## ビルド・実行

- WASMビルド: `make build` (GOOS=js GOARCH=wasm)
- 開発サーバー起動: `make serve` (ビルド後 http://localhost:8080)
- クリーン: `make clean`
- wasm_exec.js更新: `make update-wasm-exec`

## アーキテクチャ

- `main.go`: ゲーム全体（マップ生成、入力処理、描画）を単一ファイルで実装
- `index.html` + `wasm_exec.js`: WASMをブラウザで実行するためのホスト
- ゲームエンジン: [Ebiten v2](https://github.com/hajimehoshi/ebiten)
- ビルドターゲット: WebAssembly（`game.wasm`）
