package main

import (
	"context"
	"log"
	"os"

	betweenusBot "github.com/shai38/betweenus/internal/bot"
	"github.com/shai38/betweenus/internal/storage"
)

const dbPath = "betweenus.db"

func main() {
	// Получение токена из переменной окружения
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	rep, err := storage.NewSQLiteRepository(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := rep.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	// Создание бота с заданием поведения по умолчанию
	b, err := betweenusBot.New(token, rep)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Регистрация команд
	if err := betweenusBot.SetBotCommands(ctx, b); err != nil {
		log.Printf("failed to set bot commands: %v", err)
	}

	log.Println("BetweenUs is running...")

	// Запуск
	b.Start(ctx)
}
