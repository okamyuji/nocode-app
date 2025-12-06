#!/usr/bin/env bash
#
# Goコード品質チェックスクリプト
# 実行内容: go fmt, goimports, go vet, staticcheck, golangci-lint, go build, go test
#

set -e

# 出力用カラー定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # カラーなし

# スクリプトのディレクトリ
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}  Goコード品質チェック${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# チェック失敗フラグ
FAILED=0

# 1. go fmt
echo -e "${YELLOW}[1/7] go fmt を実行中...${NC}"
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    echo -e "${RED}✗ 以下のファイルにフォーマットが必要です:${NC}"
    echo "$UNFORMATTED"
    echo ""
    echo -e "${YELLOW}'gofmt -w .' を実行して修正してください${NC}"
    FAILED=1
else
    echo -e "${GREEN}✓ 全ファイルが正しくフォーマットされています${NC}"
fi
echo ""

# 2. goimports
echo -e "${YELLOW}[2/7] goimports を実行中...${NC}"
if command -v goimports &> /dev/null; then
    UNIMPORTED=$(goimports -l .)
    if [ -n "$UNIMPORTED" ]; then
        echo -e "${RED}✗ 以下のファイルにインポートの問題があります:${NC}"
        echo "$UNIMPORTED"
        echo ""
        echo -e "${YELLOW}'goimports -w .' を実行して修正してください${NC}"
        FAILED=1
    else
        echo -e "${GREEN}✓ 全インポートが正しく整理されています${NC}"
    fi
else
    echo -e "${YELLOW}⚠ goimports がインストールされていません。以下でインストール:${NC}"
    echo "  go install golang.org/x/tools/cmd/goimports@latest"
fi
echo ""

# 3. go vet
echo -e "${YELLOW}[3/7] go vet を実行中...${NC}"
if go vet ./... 2>&1; then
    echo -e "${GREEN}✓ go vet 合格${NC}"
else
    echo -e "${RED}✗ go vet で問題が見つかりました${NC}"
    FAILED=1
fi
echo ""

# 4. staticcheck（インストール済みの場合）
echo -e "${YELLOW}[4/7] staticcheck を実行中...${NC}"
if command -v staticcheck &> /dev/null; then
    if staticcheck ./... 2>&1; then
        echo -e "${GREEN}✓ staticcheck 合格${NC}"
    else
        echo -e "${RED}✗ staticcheck で問題が見つかりました${NC}"
        FAILED=1
    fi
else
    echo -e "${YELLOW}⚠ staticcheck がインストールされていません。以下でインストール:${NC}"
    echo "  go install honnef.co/go/tools/cmd/staticcheck@latest"
fi
echo ""

# 5. golangci-lint（インストール済みの場合）
echo -e "${YELLOW}[5/7] golangci-lint を実行中...${NC}"
if command -v golangci-lint &> /dev/null; then
    if golangci-lint run ./... 2>&1; then
        echo -e "${GREEN}✓ golangci-lint 合格${NC}"
    else
        echo -e "${RED}✗ golangci-lint で問題が見つかりました${NC}"
        FAILED=1
    fi
else
    echo -e "${YELLOW}⚠ golangci-lint がインストールされていません。以下でインストール:${NC}"
    echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    echo "  または"
    echo "  brew install golangci-lint"
fi
echo ""

# 6. go build
echo -e "${YELLOW}[6/7] go build を実行中...${NC}"
if go build -o /dev/null ./... 2>&1; then
    echo -e "${GREEN}✓ go build 合格${NC}"
else
    echo -e "${RED}✗ go build 失敗${NC}"
    FAILED=1
fi
echo ""

# 7. go test（シャッフルとカバレッジ付き）
echo -e "${YELLOW}[7/7] go test を実行中 (shuffle=on, count=1, カバレッジ計測)...${NC}"
COVERAGE_FILE="coverage.out"
if go test -shuffle=on -count=1 -coverprofile="$COVERAGE_FILE" ./... 2>&1; then
    echo -e "${GREEN}✓ 全テスト合格${NC}"
    echo ""
    
    # カバレッジサマリーを出力
    echo -e "${BLUE}--- カバレッジサマリー ---${NC}"
    go tool cover -func="$COVERAGE_FILE" | grep -E "^total:|coverage:" | while read line; do
        if echo "$line" | grep -q "^total:"; then
            echo -e "${BLUE}$line${NC}"
        else
            echo "$line"
        fi
    done
    
    # パッケージ別カバレッジを表示
    echo ""
    echo -e "${BLUE}--- パッケージ別カバレッジ ---${NC}"
    go tool cover -func="$COVERAGE_FILE" | grep -E "^github.com.*[0-9]+\.[0-9]+%$" | awk -F'/' '{
        pkg = $NF
        gsub(/\t.*/, "", pkg)
        coverage = $NF
        gsub(/.*\t/, "", coverage)
        printf "%-50s %s\n", pkg, coverage
    }' | sort -u
    
    echo ""
else
    echo -e "${RED}✗ 一部のテストが失敗しました${NC}"
    FAILED=1
fi
echo ""

# サマリー
echo -e "${BLUE}================================${NC}"
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}  全チェック合格！${NC}"
    echo -e "${BLUE}================================${NC}"
    exit 0
else
    echo -e "${RED}  一部のチェックが失敗しました${NC}"
    echo -e "${BLUE}================================${NC}"
    exit 1
fi
