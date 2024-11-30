package bot

import (
	"database/sql"
	"log"
	"strings"

	"IT-dep-final_project/internal/commands"
	"IT-dep-final_project/internal/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api *tgbotapi.BotAPI
	db  *sql.DB
}

func NewBot(token string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	database, err := db.Connect()
	if err != nil {
		return nil, err
	}

	return &Bot{api: api, db: database}, nil
}

func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore non-Message Updates
			continue
		}

		command := strings.Split(update.Message.Text, " ")
		switch command[0] {
		case "/add":
			commands.AddTask(b.db, update.Message.Chat.ID, command[1:])
		case "/list":
			commands.ListTasks(b.db, update.Message.Chat.ID, b.api)
		case "/complete":
			commands.CompleteTask(b.db, update.Message.Chat.ID, command[1:], b.api)
		default:
			b.api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command. Use /add, /list, or /complete."))
		}
	}
}
