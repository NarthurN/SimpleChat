package settings

import "strings"

const (
	ServerAddress  = "localhost:8080"
	ServerProtocol = "tcp"
	EncryptionKey  = 3
)

// simpleEncrypt сдвигает каждый символ в строке.
func SimpleEncrypt(text string) string {
	var result strings.Builder
	for _, char := range text {
		encryptedChar := char + EncryptionKey
		result.WriteRune(encryptedChar)
	}
	return result.String()
}

// simpleDecrypt возвращает символы на место.
func SimpleDecrypt(encryptedText string) string {
	var result strings.Builder
	for _, char := range encryptedText {
		decryptedChar := char - EncryptionKey
		result.WriteRune(decryptedChar)
	}
	return result.String()
}
