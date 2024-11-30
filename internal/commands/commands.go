package commands

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func AddTask(db *sql.DB, chatID int64, args []string) {
	if len(args) < 1 {
		return
	}
	task := strings.Join(args, " ")
	_, err := db.Exec("INSERT INTO tasks (chat_id, task) VALUES ($1, $2)", chatID, task)
	if err != nil {
		log.Println("Error adding task:", err)
		return
	}
	msg := fmt.Sprintf("Task added: %s", task)
	// Здесь вы можете отправить сообщение пользователю, если у вас есть доступ к API
}

func ListTasks(db *sql.DB, chatID int64, api *tgbotapi.BotAPI) {
	rows, err := db.Query("SELECT id, task, completed FROM tasks WHERE chat_id = $1", chatID)
	if err != nil {
		log.Println("Error fetching tasks:", err)
		api.Send(tgbotapi.NewMessage(chatID, "Error fetching tasks."))
		return
	}
	defer rows.Close()

	var tasks []string
	for rows.Next() {
		var id int
		var task string
		var completed bool
		if err := rows.Scan(&id, &task, &completed); err != nil {
			log.Println("Error scanning task:", err)
			continue
		}
		status := " "
		if completed {
			status = "[✓] "
		}
		tasks = append(tasks, fmt.Sprintf("%d: %s%s", id, status, task))
	}

	if len(tasks) == 0 {
		api.Send(tgbotapi.NewMessage(chatID, "No tasks found."))
	} else {
		api.Send(tgbotapi.NewMessage(chatID, "Tasks:\n"+strings.Join(tasks, "\n")))
	}
}

func CompleteTask(db *sql.DB, chatID int64, args []string, api *tgbotapi.BotAPI) {
	if len(args) < 1 {
		api.Send(tgbotapi.NewMessage(chatID, "Usage: /complete <task_id>"))
		return
	}
	taskID := args[0]
	_, err := db.Exec("UPDATE tasks SET completed = TRUE WHERE id = $1 AND chat_id = $2", taskID, chatID)
	if err != nil {
		log.Println("Error completing task:", err)
		api.Send(tgbotapi.NewMessage(chatID, "Error completing task."))
	} else {
		api.Send(tgbotapi.NewMessage(chatID, "Task completed: "+taskID))
	}
}
