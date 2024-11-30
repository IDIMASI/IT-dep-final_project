package commands

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AddTask добавляет новую задачу с заданным временем уведомления
// AddTask добавляет новую задачу с заданным временем уведомления
func AddTask(db *sql.DB, chatID int64, args []string, scheduler *gocron.Scheduler, api *tgbotapi.BotAPI) {
	if len(args) < 2 {
		api.Send(tgbotapi.NewMessage(chatID, "Usage: /add <task> <due_time (HH:MM:SS)>"))
		return
	}
	task := strings.Join(args[:len(args)-1], " ")
	dueTime := args[len(args)-1] // Время в формате HH:MM:SS

	// Сохраняем задачу в базе данных
	_, err := db.Exec("INSERT INTO tasks (chat_id, task, due_time) VALUES ($1, $2, $3)", chatID, task, dueTime)
	if err != nil {
		log.Println("Error adding task:", err)
		api.Send(tgbotapi.NewMessage(chatID, "Error adding task."))
		return
	}

	// Парсим время
	dueTimeParsed, err := time.Parse("15:04:05", dueTime)
	if err != nil {
		log.Println("Error parsing due time:", err)
		api.Send(tgbotapi.NewMessage(chatID, "Error parsing due time. Please use HH:MM:SS format."))
		return
	}

	// Получаем текущее время
	now := time.Now()
	dueTimeParsed = time.Date(now.Year(), now.Month(), now.Day(), dueTimeParsed.Hour(), dueTimeParsed.Minute(), dueTimeParsed.Second(), 0, now.Location())

	// Если время уже прошло, добавляем один день
	if dueTimeParsed.Before(now) {
		dueTimeParsed = dueTimeParsed.Add(24 * time.Hour)
	}

	// Запланируем задачу
	scheduler.At(dueTimeParsed.Format("15:04")).Do(func() {
		// Отправляем уведомление
		message := fmt.Sprintf("Reminder: %s", task)
		api.Send(tgbotapi.NewMessage(chatID, message))
		log.Printf("Sent notification to chat %d: %s", chatID, task)
	})

	api.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Task added: %s (Due: %s)", task, dueTime)))
}

// ListTasks lists all tasks for a given chat ID
func ListTasks(db *sql.DB, chatID int64, api *tgbotapi.BotAPI) {
	rows, err := db.Query("SELECT task, due_time FROM tasks WHERE chat_id = $1", chatID)
	if err != nil {
		log.Println("Error retrieving tasks:", err)
		api.Send(tgbotapi.NewMessage(chatID, "Error retrieving tasks."))
		return
	}
	defer rows.Close()

	var tasks []string
	for rows.Next() {
		var task, dueTime string
		if err := rows.Scan(&task, &dueTime); err != nil {
			log.Println("Error scanning task:", err)
			continue
		}
		tasks = append(tasks, fmt.Sprintf("Task: %s (Due: %s)", task, dueTime))
	}

	if len(tasks) == 0 {
		api.Send(tgbotapi.NewMessage(chatID, "No tasks found."))
		return
	}

	api.Send(tgbotapi.NewMessage(chatID, strings.Join(tasks, "\n")))
}

// CompleteTask marks a task as complete
// CompleteTask marks a task as complete
func CompleteTask(db *sql.DB, chatID int64, args []string, api *tgbotapi.BotAPI) {
	if len(args) < 1 {
		api.Send(tgbotapi.NewMessage(chatID, "Usage: /complete <task_id>"))
		return
	}
	taskID := args[0] // Предполагаем, что task_id передан как первый аргумент

	_, err := db.Exec("DELETE FROM tasks WHERE chat_id = $1 AND id = $2", chatID, taskID)
	if err != nil {
		log.Println("Error completing task:", err)
		api.Send(tgbotapi.NewMessage(chatID, "Error completing task."))
		return
	}

	api.Send(tgbotapi.NewMessage(chatID, "Task completed successfully."))
}
