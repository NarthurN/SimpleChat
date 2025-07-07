package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Тестируем шифрование и дешифрование
func TestEncryptionDecryption(t *testing.T) {
	originalText := "Hello, World! 123"

	// Шифруем
	encrypted := SimpleEncrypt(originalText)
	assert.NotEqual(t, originalText, encrypted, "Зашифрованный текст не должен совпадать с оригиналом")

	// Дешифруем
	decrypted := SimpleDecrypt(encrypted)
	assert.Equal(t, originalText, decrypted, "Расшифрованный текст должен совпадать с оригиналом")
}

// Тест на пустую строку
func TestEncrypt_EmptyString(t *testing.T) {
	assert.Equal(t, "", SimpleEncrypt(""), "Шифрование пустой строки должно возвращать пустую строку")
	assert.Equal(t, "", SimpleDecrypt(""), "Дешифрование пустой строки должно возвращать пустую строку")
}
