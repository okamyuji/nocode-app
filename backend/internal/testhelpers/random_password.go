package testhelpers

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// randomTestPassword テストコンテナ用のパスワードを実行時に生成する。
//
// リテラルのパスワードをソースに置くとシークレットスキャナ（GitGuardian）が
// ハードコードされた資格情報として検出するため、値を固定しない。
//
// 文字集合はDSNに素のまま埋め込んでも構造を壊さないものだけに絞る。
// 具体的には "@:/?&%+'\"\\ #" と空白を除外する（"#" はURL形式DSNでフラグメント区切りになる）。
// SQL Serverのパスワード複雑性要件を満たすため、大文字・小文字・数字・記号を1文字ずつ必ず含める。
func randomTestPassword() string {
	const (
		upper   = "ABCDEFGHJKLMNPQRSTUVWXYZ"
		lower   = "abcdefghijkmnpqrstuvwxyz"
		digits  = "23456789"
		symbols = "!-_"
		length  = 16
	)
	all := upper + lower + digits + symbols

	chars := []byte{
		pickChar(upper),
		pickChar(lower),
		pickChar(digits),
		pickChar(symbols),
	}
	for len(chars) < length {
		chars = append(chars, pickChar(all))
	}

	// Fisher-Yatesで並びをシャッフルし、先頭4文字が常に同じ文字種になるのを避ける。
	for i := len(chars) - 1; i > 0; i-- {
		j := randomInt(i + 1)
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}

// pickChar 与えられた文字集合から暗号論的乱数で1文字選ぶ
func pickChar(set string) byte {
	return set[randomInt(len(set))]
}

// randomInt [0, n) の乱数を返す。crypto/randが失敗する環境は想定しないためpanicする。
func randomInt(n int) int {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic(fmt.Sprintf("テスト用パスワードの生成に失敗しました: %v", err))
	}
	return int(v.Int64())
}
