package main

import (
	"context"
	"log"
	"os"

	betweenusBot "github.com/shai38/betweenus/internal/bot"
)

func main() {
	// Получение токена из переменной окружения
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	// Создание бота с заданием поведения по умолчанию
	b, err := betweenusBot.New(token)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("BetweenUs is running...")

	// Запуск
	b.Start(context.Background())
}
