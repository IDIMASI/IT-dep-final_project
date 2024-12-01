package bot

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"IT-dep-final_project/internal/commands"
	"IT-dep-final_project/internal/db"

	"github.com/go-co-op/gocron"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api       *tgbotapi.BotAPI
	db        *sql.DB
	scheduler *gocron.Scheduler
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

	scheduler := gocron.NewScheduler(time.UTC)

	return &Bot{api: api, db: database, scheduler: scheduler}, nil
}

func (b *Bot) Start() {
	log.Printf("Authorized on account %s", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	go b.scheduler.StartAsync()

	for update := range updates {
		if update.Message == nil {
			continue
		}

		command := strings.Split(update.Message.Text, " ")
		switch command[0] {
		case "/add":
			commands.AddTask(b.db, update.Message.Chat.ID, command[1:], b.scheduler, b.api)
		case "/list":
			commands.ListTasks(b.db, update.Message.Chat.ID, b.api)
		case "/complete":
			commands.CompleteTask(b.db, update.Message.Chat.ID, command[1:], b.api)
		case "/completeall":
			commands.CompleteAllTasks(b.db, update.Message.Chat.ID, b.api)
		case "/week":
			commands.WeekTasks(b.db, update.Message.Chat.ID, b.api)
		case "/month":
			commands.MonthTasks(b.db, update.Message.Chat.ID, b.api)
		default:
			b.api.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Используйте /add, /list, /complete, /completeall, /week или /month."))
		}
	}
}
