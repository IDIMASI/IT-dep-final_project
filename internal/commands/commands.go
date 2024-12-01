package commands

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AddTask добавляет новую задачу с заданным временем уведомления
// AddTask добавляет новую задачу с заданным временем уведомления
func AddTask(db *sql.DB, chatID int64, args []string, scheduler *gocron.Scheduler, api *tgbotapi.BotAPI) {
	if len(args) < 3 {
		api.Send(tgbotapi.NewMessage(chatID, "Usage: /add <task> <due_date (DD-MM-YYYY)> <due_time (HH:MM:SS)>"))
		return
	}
	task := strings.Join(args[:len(args)-2], " ")
	dueDate := args[len(args)-2] // Дата в формате DD-MM-YYYY
	dueTime := args[len(args)-1] // Время в формате HH:MM:SS

	// Разделяем дату на компоненты
	dateParts := strings.Split(dueDate, "-")
	if len(dateParts) != 3 {
		api.Send(tgbotapi.NewMessage(chatID, "Неверный формат даты. Пожалуйста, используйте DD-MM-YYYY."))
		return
	}

	day, err1 := strconv.Atoi(dateParts[0])
	month, err2 := strconv.Atoi(dateParts[1])
	year, err3 := strconv.Atoi(dateParts[2])

	if err1 != nil || err2 != nil || err3 != nil || month < 1 || month > 12 || day < 1 || day > 31 {
		api.Send(tgbotapi.NewMessage(chatID, "Неверная дата. Пожалуйста, введите корректную дату."))
		return
	}

	// Разделяем часы, минуты и секунды
	parts := strings.Split(dueTime, ":")
	if len(parts) != 3 {
		api.Send(tgbotapi.NewMessage(chatID, "Неверный формат времени. Пожалуйста, используйте ЧЧ:ММ:СС."))
		return
	}

	hours, err1 := strconv.Atoi(parts[0])
	minutes, err2 := strconv.Atoi(parts[1])
	seconds, err3 := strconv.Atoi(parts[2])

	if err1 != nil || err2 != nil || err3 != nil || hours < 0 || hours > 23 || minutes < 0 || minutes > 59 || seconds < 0 || seconds > 59 {
		api.Send(tgbotapi.NewMessage(chatID, "Неверное время. Пожалуйста, введите часы от 0 до 23, минуты от 0 до 59 и секунды от 0 до 59."))
		return
	}

	// Сохраняем задачу в базе данных
	_, err := db.Exec("INSERT INTO tasks (chat_id, task, due_time) VALUES ($1, $2, $3)", chatID, task, fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", year, month, day, hours, minutes, seconds))
	if err != nil {
		log.Println("Error adding task:", err)
		api.Send(tgbotapi.NewMessage(chatID, "Ошибка при добавлении задачи."))
		return
	}

	// Создаем временную зону UTC+3
	utcPlus3 := time.FixedZone("UTC+3", 3*60*60)

	// Создаем dueTimeParsed в UTC+3
	dueTimeParsed := time.Date(year, time.Month(month), day, hours, minutes, seconds, 0, utcPlus3)

	// Если время уже прошло, добавляем один день
	now := time.Now().In(utcPlus3)
	if dueTimeParsed.Before(now) {
		dueTimeParsed = dueTimeParsed.Add(24 * time.Hour)
	}

	// Запланируем задачу
	log.Println(dueTimeParsed.Format("2006-01-02 15:04:05"))
	scheduler.Every(1).Day().At(fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)).Do(func() {
		// Отправляем уведомление
		message := fmt.Sprintf("Напоминание: %s", task)
		_, err := api.Send(tgbotapi.NewMessage(chatID, message))
		if err != nil {
			log.Printf("Ошибка при отправке уведомления в чат %d: %s", chatID, err)
		} else {
			log.Printf("Уведомление отправлено в чат %d: %s", chatID, task)
		}
	})

	// Запускаем планировщик в основном потоке
	scheduler.StartAsync()

	api.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Задача добавлена: %s (Срок: %s %s)", task, dueDate, dueTime)))
}

// ListTasks lists all tasks for a given chat ID
func ListTasks(db *sql.DB, chatID int64, api *tgbotapi.BotAPI) {
	rows, err := db.Query("SELECT id, task, due_time FROM tasks WHERE chat_id = $1", chatID)
	if err != nil {
		log.Println("Error retrieving tasks:", err)
		api.Send(tgbotapi.NewMessage(chatID, "Error retrieving tasks."))
		return
	}
	defer rows.Close()

	var tasks []string
	for rows.Next() {
		var id int
		var task, dueTime string
		if err := rows.Scan(&id, &task, &dueTime); err != nil {
			log.Println("Error scanning task:", err)
			continue
		}
		tasks = append(tasks, fmt.Sprintf("Task %d: %s (%s)", id, task, dueTime))
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

func CompleteAllTasks(db *sql.DB, chatID int64, api *tgbotapi.BotAPI) {
	// Delete all tasks for the given chat ID
	_, err := db.Exec("DELETE FROM tasks WHERE chat_id = $1", chatID)
	if err != nil {
		log.Println("Error deleting tasks:", err)
		api.Send(tgbotapi.NewMessage(chatID, "Error deleting tasks."))
		return
	}

	api.Send(tgbotapi.NewMessage(chatID, "All tasks have been completed and removed."))
}

// WeekTasks lists all tasks for a given chat ID within the next week
func WeekTasks(db *sql.DB, chatID int64, api *tgbotapi.BotAPI) {
	now := time.Now()
	nextWeek := now.AddDate(0, 0, 7)

	rows, err := db.Query("SELECT id, task, due_time FROM tasks WHERE chat_id = $1 AND due_time BETWEEN $2 AND $3", chatID, now.Format("2006-01-02 15:04:05"), nextWeek.Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Println("Error retrieving tasks:", err)
		api.Send(tgbotapi.NewMessage(chatID, "Error with recieving tasks"))
		return
	}
	defer rows.Close()

	var tasks []string
	for rows.Next() {
		var id int
		var task, dueTime string
		if err := rows.Scan(&id, &task, &dueTime); err != nil {
			log.Println("Error scanning task:", err)
			continue
		}
		tasks = append(tasks, fmt.Sprintf("Task %d: %s (Period: %s)", id, task, dueTime))
	}

	if len(tasks) == 0 {
		api.Send(tgbotapi.NewMessage(chatID, "tasks for the next week were not found"))
		return
	}

	api.Send(tgbotapi.NewMessage(chatID, strings.Join(tasks, "\n")))
}

// MonthTasks lists all tasks for a given chat ID within the next month
func MonthTasks(db *sql.DB, chatID int64, api *tgbotapi.BotAPI) {
	now := time.Now()
	nextMonth := now.AddDate(0, 1, 0)

	rows, err := db.Query("SELECT id, task, due_time FROM tasks WHERE chat_id = $1 AND due_time BETWEEN $2 AND $3", chatID, now.Format("2006-01-02 15:04:05"), nextMonth.Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Println("Error retrieving tasks:", err)
		api.Send(tgbotapi.NewMessage(chatID, "Error with recieving tasks"))
		return
	}
	defer rows.Close()

	var tasks []string
	for rows.Next() {
		var id int
		var task, dueTime string
		if err := rows.Scan(&id, &task, &dueTime); err != nil {
			log.Println("Error scanning task:", err)
			continue
		}
		tasks = append(tasks, fmt.Sprintf("Task %d: %s (Period: %s)", id, task, dueTime))
	}

	if len(tasks) == 0 {
		api.Send(tgbotapi.NewMessage(chatID, "tasks for the next month were not found"))
		return
	}

	api.Send(tgbotapi.NewMessage(chatID, strings.Join(tasks, "\n")))
}
