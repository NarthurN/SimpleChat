package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/NarthurN/SimpleChat/shared/settings"
)

// Горутина для приёма сообщений от сервера и вывода их в консоль
func receiveMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Соединение с сервером закрыто.")
			} else {
				fmt.Printf("Ошибка чтения от сервера: %v\n", err)
			}
			os.Exit(0)
		}
		// decryptedMessage := settings.SimpleDecrypt(strings.TrimSpace(msg))
		// fmt.Println(decryptedMessage)
		trimmedMessage := strings.TrimSpace(message)
		parts := strings.SplitN(trimmedMessage, ":", 2)

		if len(parts) >= 2 {
			prefix := parts[0]
			content := parts[1]

			if strings.HasPrefix(prefix, "[private from") {
				idx := strings.Index(trimmedMessage, "]:")
				if idx != -1 {
					formattedPrefix := trimmedMessage[:idx+2]
					encryptedContent := trimmedMessage[idx+2:]

					decryptedContent := settings.SimpleDecrypt(encryptedContent)
					fmt.Printf("%s%s\n", formattedPrefix, decryptedContent)
				} else {
					// Если формат не соответствует ожидаемому, выводим как есть
					fmt.Println(trimmedMessage)
				}
			} else if prefix == "SERVER" {
				fmt.Println(trimmedMessage)
			} else {
				// Это обычное сообщение от пользователя, например "Alice: Hello"
				// Расшифровываем только часть после двоеточия
				decryptedContent := settings.SimpleDecrypt(content)
				fmt.Printf("%s:%s\n", prefix, decryptedContent)
			}
		} else {
			fmt.Println(trimmedMessage)
		}
	}
}

func main() {
	conn, err := net.Dial(settings.ServerProtocol, settings.ServerAddress)
	if err != nil {
		log.Fatalf("Не удалось подключиться к серверу: %v", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("невозможно закрыть соединение: %v", err)
		}
	}()

	fmt.Println("Введите команду JOIN:<username> для входа в чат:")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		if !scanner.Scan() {
			log.Println("Завершение работы клиента.")
			return
		}
		line := scanner.Text()
		if strings.HasPrefix(line, "JOIN:") && len(strings.TrimSpace(line)) > 5 {
			_, err := fmt.Fprintln(conn, line)
			if err != nil {
				log.Fatalf("Ошибка отправки команды JOIN: %v", err)
			}
			break
		} else {
			fmt.Println("Сначала введите команду JOIN:<username>")
		}
	}

	// Запуск приёма сообщений от сервера в отдельной горутине
	go receiveMessages(conn)

	fmt.Println("Вы в чате! Используйте MSG:<сообщение>, P_MSG:<кому>:<сообщение> или QUIT.")

	// Основной цикл: чтение команд и сообщений с консоли и отправка на сервер
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		// QUIT — завершение работы
		if strings.ToUpper(text) == "QUIT" {
			fmt.Fprintln(conn, "QUIT")
			break
		}

		if strings.HasPrefix(text, "MSG:") {
			parts := strings.SplitN(text, ":", 2)
			if len(parts) == 2 {
				encryptedContent := settings.SimpleEncrypt(parts[1])
				fmt.Fprintf(conn, "MSG:%s\n", encryptedContent)
			} else {
				fmt.Println("Неверный формат MSG. Используйте MSG:<сообщение>")
			}
		} else if strings.HasPrefix(text, "P_MSG:") {
			parts := strings.SplitN(text, ":", 3)
			if len(parts) == 3 {
				recipient := parts[1]
				messageContent := parts[2]
				encryptedMessage := settings.SimpleEncrypt(messageContent)
				fmt.Fprintf(conn, "P_MSG:%s:%s\n", recipient, encryptedMessage)
			} else {
				fmt.Println("Неверный формат P_MSG. Используйте P_MSG:<кому>:<сообщение>")
			}
		} else {
			fmt.Println("Используйте MSG:<сообщение>, P_MSG:<кому>:<сообщение> или QUIT.")
		}
	}

	log.Println("Клиент завершил работу.")
}
