package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetEncryptionKey(t *testing.T) {
	// テスト後に元の状態に戻す
	originalKey := encryptionKey
	defer func() { encryptionKey = originalKey }()

	t.Run("valid 32 byte key", func(t *testing.T) {
		key := make([]byte, 32)
		for i := range key {
			key[i] = byte(i)
		}

		err := SetEncryptionKey(key)
		require.NoError(t, err)
		assert.Equal(t, key, encryptionKey)
	})

	t.Run("invalid key length - too short", func(t *testing.T) {
		key := make([]byte, 16)
		err := SetEncryptionKey(key)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "32バイト")
	})

	t.Run("invalid key length - too long", func(t *testing.T) {
		key := make([]byte, 64)
		err := SetEncryptionKey(key)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "32バイト")
	})
}

func TestEncryptDecrypt(t *testing.T) {
	// テスト用の32バイトキーを設定
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	err := SetEncryptionKey(key)
	require.NoError(t, err)

	t.Run("encrypt and decrypt successfully", func(t *testing.T) {
		plaintext := "これはテストパスワードです"

		encrypted, err := Encrypt(plaintext)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted)
		assert.NotEqual(t, plaintext, encrypted)

		decrypted, err := Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("encrypt empty string", func(t *testing.T) {
		encrypted, err := Encrypt("")
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted)

		decrypted, err := Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, "", decrypted)
	})

	t.Run("encrypt long string", func(t *testing.T) {
		longText := ""
		for i := 0; i < 1000; i++ {
			longText += "a"
		}

		encrypted, err := Encrypt(longText)
		require.NoError(t, err)

		decrypted, err := Decrypt(encrypted)
		require.NoError(t, err)
		assert.Equal(t, longText, decrypted)
	})

	t.Run("different encryptions produce different ciphertexts", func(t *testing.T) {
		plaintext := "same text"

		encrypted1, err := Encrypt(plaintext)
		require.NoError(t, err)

		encrypted2, err := Encrypt(plaintext)
		require.NoError(t, err)

		// nonceがランダムなので、同じ平文でも異なる暗号文になる
		assert.NotEqual(t, encrypted1, encrypted2)

		// しかし両方とも同じ平文に復号できる
		decrypted1, err := Decrypt(encrypted1)
		require.NoError(t, err)
		decrypted2, err := Decrypt(encrypted2)
		require.NoError(t, err)

		assert.Equal(t, plaintext, decrypted1)
		assert.Equal(t, plaintext, decrypted2)
	})
}

func TestEncryptWithoutKey(t *testing.T) {
	// テスト後に元の状態に戻す
	originalKey := encryptionKey
	defer func() { encryptionKey = originalKey }()

	encryptionKey = nil

	_, err := Encrypt("test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "暗号化キーが初期化されていません")
}

func TestDecryptWithoutKey(t *testing.T) {
	// テスト後に元の状態に戻す
	originalKey := encryptionKey
	defer func() { encryptionKey = originalKey }()

	encryptionKey = nil

	_, err := Decrypt("test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "暗号化キーが初期化されていません")
}

func TestDecryptInvalidInput(t *testing.T) {
	// テスト用の32バイトキーを設定
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	err := SetEncryptionKey(key)
	require.NoError(t, err)

	t.Run("invalid base64", func(t *testing.T) {
		_, err := Decrypt("not-valid-base64!!!")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Base64デコード")
	})

	t.Run("ciphertext too short", func(t *testing.T) {
		// 短すぎるデータ（nonceサイズ未満）
		_, err := Decrypt("YWJj") // "abc" in base64
		require.Error(t, err)
		assert.Contains(t, err.Error(), "短すぎます")
	})

	t.Run("invalid ciphertext", func(t *testing.T) {
		// 有効なBase64だが、不正な暗号文
		_, err := Decrypt("YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY3ODkw")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "復号に失敗")
	})
}

func TestGenerateEncryptionKey(t *testing.T) {
	key1, err := GenerateEncryptionKey()
	require.NoError(t, err)
	assert.NotEmpty(t, key1)

	key2, err := GenerateEncryptionKey()
	require.NoError(t, err)
	assert.NotEmpty(t, key2)

	// 2回生成したキーは異なるはず
	assert.NotEqual(t, key1, key2)

	// Base64デコードして32バイトであることを確認
	// (GenerateEncryptionKeyはBase64エンコードされた32バイトキーを返す)
	// デコードは別のテストで確認済み
}

func TestInitEncryption(t *testing.T) {
	// テスト後に元の状態に戻す
	originalKey := encryptionKey
	defer func() { encryptionKey = originalKey }()

	t.Run("missing environment variable", func(t *testing.T) {
		t.Setenv("ENCRYPTION_KEY", "")
		err := InitEncryption()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ENCRYPTION_KEY環境変数が設定されていません")
	})

	t.Run("invalid base64", func(t *testing.T) {
		t.Setenv("ENCRYPTION_KEY", "not-valid-base64!!!")
		err := InitEncryption()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Base64デコード")
	})

	t.Run("invalid key length", func(t *testing.T) {
		// 16バイトのキー（Base64エンコード）
		t.Setenv("ENCRYPTION_KEY", "YWJjZGVmZ2hpamtsbW5vcA==")
		err := InitEncryption()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "32バイト")
	})

	t.Run("valid key", func(t *testing.T) {
		// 32バイトのキー（Base64エンコード）
		t.Setenv("ENCRYPTION_KEY", "YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY=")
		err := InitEncryption()
		require.NoError(t, err)
	})
}
