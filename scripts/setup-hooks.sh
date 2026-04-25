#!/usr/bin/env bash
# git の hooksPath を .githooks に切り替える。
# クローン後に一度だけ実行すれば、以降のコミットで pre-commit が自動的に走る。
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

git config core.hooksPath .githooks
chmod +x .githooks/* 2>/dev/null || true

echo "git hooksPath を .githooks に設定しました。"
echo "今後は git commit 前にフロントエンド / バックエンドの品質検証が走り、失敗するとコミットが中断されます。"
