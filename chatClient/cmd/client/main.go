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
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Соединение с сервером закрыто.")
			} else {
				fmt.Printf("Ошибка чтения от сервера: %v\n", err)
			}
			os.Exit(0)
		}
		fmt.Print(msg)
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

	fmt.Println("Теперь вы можете отправлять сообщения (MSG:<текст>) или команду QUIT для выхода.")

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
		// MSG:<message> — отправка сообщения
		if strings.HasPrefix(text, "MSG:") {
			fmt.Fprintln(conn, text)
		} else if strings.HasPrefix(text, "P_MSG:") {
			fmt.Fprintln(conn, text)
		} else {
			fmt.Println("Используйте MSG:<сообщение>, P_MSG:<кому>:<сообщение> или QUIT.")
		}
	}

	log.Println("Клиент завершил работу.")
}
