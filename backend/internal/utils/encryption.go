package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

// EncryptionKey 環境変数から取得する暗号化キー
var encryptionKey []byte

// ErrEncryptionNotInitialized 暗号化キーが初期化されていない場合のエラー
var ErrEncryptionNotInitialized = errors.New("暗号化キーが初期化されていません。ENCRYPTION_KEY環境変数を設定してください")

// IsEncryptionInitialized 暗号化キーが初期化されているかどうかを返す
func IsEncryptionInitialized() bool {
	return len(encryptionKey) > 0
}

// InitEncryption 暗号化キーを初期化する
func InitEncryption() error {
	keyStr := os.Getenv("ENCRYPTION_KEY")
	if keyStr == "" {
		return errors.New("ENCRYPTION_KEY環境変数が設定されていません")
	}

	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return fmt.Errorf("ENCRYPTION_KEYのBase64デコードに失敗しました: %w", err)
	}

	if len(key) != 32 {
		return fmt.Errorf("ENCRYPTION_KEYは32バイトである必要があります（現在: %dバイト）", len(key))
	}

	encryptionKey = key
	return nil
}

// SetEncryptionKey テスト用に暗号化キーを直接設定する
func SetEncryptionKey(key []byte) error {
	if len(key) != 32 {
		return fmt.Errorf("暗号化キーは32バイトである必要があります")
	}
	encryptionKey = key
	return nil
}

// Encrypt 文字列をAES-256-GCMで暗号化する
// 戻り値: Base64エンコードされた（nonce + 暗号文）
func Encrypt(plaintext string) (string, error) {
	if len(encryptionKey) == 0 {
		return "", ErrEncryptionNotInitialized
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("AES暗号の初期化に失敗しました: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("GCMモードの初期化に失敗しました: %w", err)
	}

	// 12バイトのランダムなnonceを生成
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("nonceの生成に失敗しました: %w", err)
	}

	// 暗号化（nonceを先頭に付加）
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Base64エンコードして返す
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt AES-256-GCMで暗号化された文字列を復号する
// 入力: Base64エンコードされた（nonce + 暗号文）
func Decrypt(ciphertext string) (string, error) {
	if len(encryptionKey) == 0 {
		return "", ErrEncryptionNotInitialized
	}

	// Base64デコード
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("Base64デコードに失敗しました: %w", err)
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("AES暗号の初期化に失敗しました: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("GCMモードの初期化に失敗しました: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("暗号文が短すぎます")
	}

	// nonceと暗号文を分離
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// 復号
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("復号に失敗しました: %w", err)
	}

	return string(plaintext), nil
}

// GenerateEncryptionKey 新しい32バイトの暗号化キーを生成してBase64で返す
// ヘルパー関数: 初期セットアップ時に使用
func GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("ランダムキーの生成に失敗しました: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
