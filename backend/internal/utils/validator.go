// Package utils バリデーション、レスポンス処理などのユーティリティ関数を提供
package utils

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator バリデーターインスタンスをラップする構造体
type Validator struct {
	validate *validator.Validate
}

// fieldCodeRegex フィールドコードのバリデーション用正規表現（パフォーマンス向上のため事前コンパイル）
var fieldCodeRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

// NewValidator 新しいValidatorを作成する
func NewValidator() *Validator {
	v := validator.New()

	// フィールドコード用のカスタムバリデーター
	// 英字で始まり、英数字とアンダースコアのみ許可
	err := v.RegisterValidation("fieldcode", func(fl validator.FieldLevel) bool {
		code := fl.Field().String()
		if code == "" {
			return false
		}
		return fieldCodeRegex.MatchString(code)
	})
	if err != nil {
		panic("failed to register fieldcode validation: " + err.Error())
	}

	return &Validator{validate: v}
}

// Validate 構造体をバリデートする
func (v *Validator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}

// ValidateVar 単一の変数をバリデートする
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// ParseAndValidate JSONをパースしてバリデートする
func (v *Validator) ParseAndValidate(r *http.Request, dest interface{}) error {
	if err := ParseJSON(r, dest); err != nil {
		return err
	}
	return v.Validate(dest)
}

// IsValidFieldCode フィールドコードが有効かどうかを確認する
func IsValidFieldCode(code string) bool {
	if code == "" || len(code) > 64 {
		return false
	}
	// 英数字とアンダースコアのみ許可
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, code)
	return matched
}

// SanitizeTableName テーブル名をサニタイズする
func SanitizeTableName(name string) string {
	// 英数字とアンダースコア以外の文字を削除
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return re.ReplaceAllString(name, "")
}

// GetQueryParam デフォルト値付きでクエリパラメータを取得する
func GetQueryParam(r *http.Request, key, defaultValue string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetQueryParamInt デフォルト値付きでクエリパラメータをintとして取得する
func GetQueryParamInt(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	var result int
	for _, c := range value {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			return defaultValue
		}
	}
	return result
}

// ExtractPathParam URLからパスパラメータを抽出する
// 標準のnet/http用のシンプルな実装
func ExtractPathParam(path, pattern string) map[string]string {
	params := make(map[string]string)

	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")

	if len(pathParts) != len(patternParts) {
		return params
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			key := strings.TrimPrefix(part, ":")
			params[key] = pathParts[i]
		}
	}

	return params
}
