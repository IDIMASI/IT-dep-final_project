package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	bot, err := tgbotapi.NewBotAPI("7202323899:AAHFa2FBQyo2-gbP82dwxyWfPKvWZ3bSeJ8")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		if update.Message.IsCommand() { // если это команда
			switch update.Message.Command() {
			case "start":
				msg.Text = "Привет ещё раз,этот телеграмм бот создан для планирования и напоминания про твои задачи.\nБот имеет такие команды:\n*Создает новую задачу(/newtask).\n*Вывод всех задач(/seealltask).\n*Задач на день(/seedaytask).\n*Задач на неделю(/seeweektask).\n*Удаление задачи(/killtask).\n"
				//обработка ввода и добавление в БД
			case "newtask":
				msg.Text = "Чтобы создать новую введи сообщения в специальном формате:\n“ Название задачи” “день” “время в который задача должна быть выполнена” “день” “время напоминания про задачу”.\nПример: ДЗ_топоры 20.08.2024 21:00 19.08.2024 21:00.\nМожно не указывать напоминаие про задачу.\n"
				//обработка ввода и добавление в БД
			case "kill task":
				msg.Text = "I'm ok."
				//удаление по id
			case "seealltask":
				//вывод данных из БД
			case "seedaytask":
				//вывод данных из БД
			case "seeweektask":
				//вывод данных из БД
			default:
				msg.Text = "Я не знаю такой команды поробуй /start"
			}
		} else { // если это не команда
			msg.Text = "Привет-привет это TO-DO-Teltgrambot для более подробной информации напиши /start"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}
