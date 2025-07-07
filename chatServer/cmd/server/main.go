package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/NarthurN/SimpleChat/shared/settings"
)

// Client теперь хранит и соединение, и имя пользователя
type Client struct {
	Conn net.Conn
	Name string
}

// Chat управляет всеми клиентами и их взаимодействием
type Chat struct {
	clients   map[string]*Client
	clientsMu sync.RWMutex
}

func NewChat() *Chat {
	return &Chat{
		clients: make(map[string]*Client),
	}
}

// Добавить клиента в чат
func (c *Chat) AddClient(client *Client) error {
	c.clientsMu.Lock()
	defer c.clientsMu.Unlock()

	if _, exists := c.clients[client.Name]; exists {
		return fmt.Errorf("клиент с именем '%s' уже существует", client.Name)
	}

	c.clients[client.Name] = client
	return nil
}

// Убрать клиента из чата
func (c *Chat) RemoveClient(name string) {
	c.clientsMu.Lock()
	defer c.clientsMu.Unlock()

	delete(c.clients, name)
}

// Рассылка сообщения
func (c *Chat) Broadcast(senderName, formattedMessage string) {
	c.clientsMu.RLock()
	defer c.clientsMu.RUnlock()

	for name, client := range c.clients {
		if name != senderName {
			_, err := fmt.Fprintln(client.Conn, formattedMessage)
			if err != nil {
				log.Printf("Не удалось отправить сообщение клиенту %s: %v\n", name, err)
			}
		}
	}

}

func main() {
	// создаем чат
	chat := NewChat()

	// Создаем соединение TCP-сервера
	serverListener, err := net.Listen(settings.ServerProtocol, settings.ServerAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}

	// Откладываем закрытие соединения
	defer func() {
		if err := serverListener.Close(); err != nil {
			log.Fatalf("failed to close listener: %v\n", err)
		}
	}()

	log.Println("server is listening on", settings.ServerAddress)

	// Обрабатываем входящие содинения
	for {
		conn, err := serverListener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v\n", err)
			continue
		}

		go handleClient(conn, chat)
	}
}

// Обработчик клиентского соединения
func handleClient(conn net.Conn, chat *Chat) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}()

	log.Printf("Новый клиент подключился: %s\n", conn.RemoteAddr())
	reader := bufio.NewReader(conn)

	// --- Фаза 1: Аутентификация (JOIN) ---
	joinMsg, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Не удалось прочитать JOIN сообщение от %s: %v", conn.RemoteAddr(), err)
		return
	}

	// Проверка валидности данных
	parts := strings.SplitN(strings.TrimSpace(joinMsg), ":", 2)
	if len(parts) != 2 || parts[0] != "JOIN" || len(parts[1]) == 0 {
		fmt.Fprintf(conn, "ERROR: Первой командой должна быть JOIN:<username>\n")
		return
	}
	username := parts[1]

	// Добавляем пользователя
	client := &Client{
		Conn: conn,
		Name: username,
	}

	if err := chat.AddClient(client); err != nil {
		fmt.Fprintf(conn, "ERROR: %v\n", err)
		return
	}
	fmt.Fprintf(conn, "Вы успешно присоединились к чату как %s!\n", username)

	// Отключаем пользователя при выходе
	defer func() {
		chat.RemoveClient(username)
		log.Printf("Пользователь '%s' покинул чат.", username)
		chat.Broadcast("", fmt.Sprintf("SERVER: %s покинул чат.", username))
	}()

	chat.Broadcast(username, fmt.Sprintf("SERVER: %s присоединился к чату.", username))

	// --- Фаза 2: Обмен сообщениями ---
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Клиент '%s' отключился.", username)
			} else {
				log.Printf("Ошибка чтения от '%s': %v", username, err)
			}
			return
		}

		// Игнорируем пустые сообщения
		trimmedMessage := strings.TrimSpace(message)
		if trimmedMessage == "" {
			continue
		}

		// Парсим команду
		parts := strings.SplitN(trimmedMessage, ":", 2)
		command := parts[0]

		switch command {
		case "MSG":
			if len(parts) < 2 || parts[1] == "" {
				fmt.Fprintln(conn, "ERROR: Неверный формат. Используйте MSG:<message>")
				continue
			}
			msgContent := parts[1]
			fullMessage := fmt.Sprintf("%s: %s", username, msgContent)
			chat.Broadcast(username, fullMessage)
		case "QUIT":
			return
		default:
			fmt.Fprintf(conn, "ERROR: Неизвестная команда '%s'\n", command)
		}
	}
}
