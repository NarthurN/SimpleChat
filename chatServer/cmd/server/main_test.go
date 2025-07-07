package main

import (
	"bufio"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChat_AddAndRemoveClient(t *testing.T) {
	chat := NewChat()
	client := &Client{Name: "alice"}

	// 1. Проверяем успешное добавление
	err := chat.AddClient(client)
	require.NoError(t, err, "Добавление нового клиента не должно вызывать ошибку")
	assert.Contains(t, chat.clients, "alice", "Клиент 'alice' должен быть в чате")

	// 2. Проверяем удаление
	chat.RemoveClient("alice")
	assert.NotContains(t, chat.clients, "alice", "Клиент 'alice' должен быть удален из чата")
}

// Тест на добавление клиента с уже существующим именем
func TestChat_AddExistingClient(t *testing.T) {
	chat := NewChat()
	client1 := Client{Name: "bob"}
	client2 := Client{Name: "bob"}

	// Добавляем первого клиента
	err1 := chat.AddClient(&client1)
	require.NoError(t, err1)

	// Пытаемся добавить второго с тем же именем
	err2 := chat.AddClient(&client2)
	assert.Error(t, err2, "Должна быть ошибка при добавлении клиента с существующим именем")
}

// Тестируем полный цикл: JOIN, MSG, QUIT
func TestHandleClient_FullFlow(t *testing.T) {
	chat := NewChat()

	// net.Pipe() создает два связанных конца соединения:
	// один для "клиента", другой для "сервера"
	clientConn, serverConn := net.Pipe()

	// Запускаем обработчик сервера в отдельной горутине
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		handleClient(serverConn, chat)
	}()

	// --- Эмулируем действия клиента ---
	clientReader := bufio.NewReader(clientConn)

	// 1. Отправляем JOIN
	_, err := clientConn.Write([]byte("JOIN:testuser\n"))
	require.NoError(t, err)

	// Читаем приветственное сообщение от сервера
	response, _ := clientReader.ReadString('\n')
	assert.Contains(t, response, "Вы успешно присоединились", "Сервер должен приветствовать пользователя")

	// 2. Отправляем MSG
	_, err = clientConn.Write([]byte("MSG:hello\n"))
	require.NoError(t, err)

	// 3. Отправляем QUIT
	_, err = clientConn.Write([]byte("QUIT\n"))
	require.NoError(t, err)

	// Ждем завершения горутины handleClient
	wg.Wait()

	// Проверяем, что клиент был удален из чата
	assert.NotContains(t, chat.clients, "testuser", "Клиент должен быть удален после QUIT")
}
