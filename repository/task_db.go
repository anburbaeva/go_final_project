package repository

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/anburbaeva/go_final_project/model"

	"github.com/jmoiron/sqlx"
)

const (
	FirstDay                  = 1
	MinusOneDay               = -1
	AddingOneMounth           = 1
	MinRepeatIntervalDays     = 1
	MaxRepeatIntervalDay      = 31
	MinMinusRepeatIntervalDay = -2
	MaxRepeatIntervalDays     = 400
	MinMonths                 = 1
	MaxMonths                 = 12
	MinWeek                   = 1
	MAX_WEEK                  = 7
	OneWeek
	LimitTasks = 25
)

type TodoTaskSqlite struct {
	db *sqlx.DB
}

func NewTodoTaskSqlite(db *sqlx.DB) *TodoTaskSqlite {
	return &TodoTaskSqlite{db: db}
}

func (t *TodoTaskSqlite) NextDate(nd model.NextDate) (string, error) {
	if nd.Repeat == "" {
		return "", fmt.Errorf("неправильный формат повтора: %v", nd.Repeat)
	}

	if !regexp.MustCompile(`^([wdm]\s.*|y)?$`).MatchString(nd.Repeat) {
		return "", fmt.Errorf("неправильный формат повтора: %v", nd.Repeat)
	}

	switch nd.Repeat[0] {
	case 'd':
		repeatIntervalDays, err := findRepeatIntervalDays(nd)
		return repeatIntervalDays, err
	case 'y':
		repeatIntervalYears, err := findRepeatIntervalYears(nd)
		return repeatIntervalYears, err
	case 'w':
		repeatIntervalWeeks, err := findRepeatIntervalWeeks(nd)
		return repeatIntervalWeeks, err
	case 'm':
		repeatIntervalMonths, err := findRepeatIntervalMonths(nd)
		return repeatIntervalMonths, err
	default:
		return "", nil
	}
}
func (t *TodoTaskSqlite) CreateTask(task model.Task) (int64, error) {
	err := t.checkTask(&task)
	if err != nil {
		return 0, err
	}

	query := fmt.Sprintf("INSERT INTO %s (title, comment, date, repeat) VALUES ($1, $2, $3, $4) RETURNING id", "scheduler")
	row := t.db.QueryRow(query, task.Title, task.Comment, task.Date, task.Repeat)

	var id int64
	if err = row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}
func (t *TodoTaskSqlite) GetTasks(search string) (model.ListTasks, error) {
	var tasks []model.Task
	var query string

	switch typeSearch(search) {
	case 0:
		query = fmt.Sprintf("SELECT * FROM %s ORDER BY date LIMIT ?", "scheduler")
		err := t.db.Select(&tasks, query, LimitTasks)
		if err != nil {
			return model.ListTasks{}, err
		}
	case 1:
		s, _ := time.Parse(`20060102`, search)
		st := s.Format(`20060102`)
		query = fmt.Sprintf("SELECT * FROM %s WHERE date = ? ORDER BY date LIMIT ?", "scheduler")
		err := t.db.Select(&tasks, query, st, LimitTasks)
		if err != nil {
			return model.ListTasks{}, err
		}
	case 2:
		searchQuery := fmt.Sprintf("%%%s%%", search)
		query := `SELECT * FROM scheduler WHERE LOWER(title) LIKE ? OR LOWER(comment) LIKE ? ORDER BY date LIMIT ?`
		rows, err := t.db.Queryx(query, searchQuery, searchQuery, LimitTasks)
		if err != nil {
			return model.ListTasks{}, err
		}
		for rows.Next() {
			var task model.Task
			err := rows.StructScan(&task)
			if err != nil {
				return model.ListTasks{}, err
			}
			tasks = append(tasks, task)
		}
	}

	if len(tasks) == 0 {
		return model.ListTasks{Tasks: []model.Task{}}, nil
	}
	return model.ListTasks{Tasks: tasks}, nil
}
func (t *TodoTaskSqlite) GetTaskById(id string) (model.Task, error) {
	if id == "" {
		return model.Task{}, fmt.Errorf("нет идентификатора")
	}
	if _, err := strconv.Atoi(id); err != nil {
		return model.Task{}, fmt.Errorf("неправильный идентификатор")
	}
	var task model.Task
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", "scheduler")
	err := t.db.Get(&task, query, id)
	if err != nil {
		return model.Task{}, fmt.Errorf("задача не найдена")
	}
	return task, err
}
func (t *TodoTaskSqlite) UpdateTask(task model.Task) error {
	err := t.checkTask(&task)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("UPDATE %s SET title = ?, comment = ?, date = ?, repeat = ? WHERE id = ?", "scheduler")
	_, err = t.db.Exec(query, task.Title, task.Comment, task.Date, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("задача не найдена")
	}
	return nil
}
func (t *TodoTaskSqlite) DeleteTask(id string) error {
	_, err := t.GetTaskById(id)
	if err != nil {
		return err
	}
	queryDelete := fmt.Sprintf("DELETE FROM %s WHERE id = ?", "scheduler")
	_, err = t.db.Exec(queryDelete, id)
	if err != nil {
		return err
	}
	return nil
}
func (t *TodoTaskSqlite) TaskDone(id string) error {
	task, err := t.GetTaskById(id)
	if err != nil {
		return err
	}

	if task.Repeat == "" {
		queryDeleteTask := fmt.Sprintf("DELETE FROM %s WHERE id = ?", "scheduler")
		logrus.Println(queryDeleteTask)
		t.db.Exec(queryDeleteTask, id)
		return nil
	}

	nd := model.NextDate{
		Date:   task.Date,
		Now:    time.Now().Format(`20060102`),
		Repeat: task.Repeat,
	}

	newDate, err := t.NextDate(nd)
	if err != nil {
		return err
	}

	task.Date = newDate
	queryUpdateTask := fmt.Sprintf("UPDATE %s SET date = ? WHERE id = ?", "scheduler")
	logrus.Println(queryUpdateTask)
	_, err = t.db.Exec(queryUpdateTask, task.Date, id)
	if err != nil {
		return err
	}
	return nil

}
func (t *TodoTaskSqlite) checkTask(task *model.Task) error {
	if task.Title == "" {
		return fmt.Errorf("название не может быть пустым")
	}

	if !regexp.MustCompile(`^([wdm]\s.*|y)?$`).MatchString(task.Repeat) {
		return fmt.Errorf("неправильный формат повтора: %v", task.Repeat)
	}

	now := time.Now().Format(`20060102`)

	if task.Date == "" {
		task.Date = now
	}

	_, err := time.Parse(`20060102`, task.Date)
	if err != nil {
		return fmt.Errorf("неправильная дата")
	}

	if task.Date < now {
		if task.Repeat == "" {
			task.Date = now
		}
		if task.Repeat != "" {
			nd := model.NextDate{
				Date:   task.Date,
				Now:    now,
				Repeat: task.Repeat,
			}
			task.Date, err = t.NextDate(nd)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func findRepeatIntervalDays(nd model.NextDate) (string, error) {
	now, err := timeNow(nd)
	if err != nil {
		return "", err
	}
	stringRepeatIntervalDays := strings.TrimPrefix(nd.Repeat, "d ")
	repeatIntervalDays, err := strconv.Atoi(stringRepeatIntervalDays)

	searchDate := nd.Date

	for searchDate <= now.Format(`20060102`) || searchDate <= nd.Date {
		d, err := time.Parse(`20060102`, searchDate)
		if err != nil {
			return "неправильная дата", fmt.Errorf("непонятно: %v", nd.Date)
		}
		searchDate = d.AddDate(0, 0, repeatIntervalDays).Format(`20060102`)
	}
	return searchDate, nil
}
func findRepeatIntervalMonths(nd model.NextDate) (string, error) {
	now, err := timeNow(nd)
	if err != nil {
		return "", err
	}
	repeatSrt := strings.TrimPrefix(nd.Repeat, "m ")
	isConstainsNumMonth := strings.Contains(repeatSrt, " ")
	monthsSlice := make([]string, 0)
	months := make([]int, 0)
	if isConstainsNumMonth {
		IndexSep := strings.Index(repeatSrt, " ")
		repeatSrtMounth := repeatSrt[IndexSep+1:]
		repeatSrt = repeatSrt[:IndexSep]
		monthsSlice = strings.Split(repeatSrtMounth, ",")

		for _, v := range monthsSlice {
			vi, err := strconv.Atoi(strings.TrimSpace(v))
			if err != nil {
				return "неправильное значение повтора.", err
			}
			if vi < MinMonths || vi > MaxMonths {
				return "неправильное значение повтора." + "нужно m <через запятую от 1 до 12>.",
					fmt.Errorf("ошибка здесь: %v", nd.Repeat)
			}
			months = append(months, vi)
		}
	}

	monthDaysSlice := strings.Split(repeatSrt, ",")

	monthDays := make([]int, 0)

	for _, v := range monthDaysSlice {
		vi, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return "ошибка конвертации", err
		}
		if vi < MinMinusRepeatIntervalDay || vi > MaxRepeatIntervalDay {
			return "нерпавильное значение повтора." + "m <через запятую от 1 до 31,-1,-2> [через запятую от 1 до 12]", fmt.Errorf("ошибка здесь: %v", nd.Repeat)
		}
		monthDays = append(monthDays, vi)
	}
	nextDates := make([]time.Time, 0)

	if len(months) > 0 {
		for i := 0; i < len(months); i++ {
			m := months[i]
			for j := 0; j < len(monthDays); j++ {
				d := monthDays[j]
				nd := findDayOfMonth(now, nd.Date, m, d)
				nextDates = append(nextDates, nd)
			}
		}
	} else if len(monthDays) > 0 {
		for _, d := range monthDays {
			nextDates = append(nextDates, findDayOfMonth(now, nd.Date, 0, d))
		}
	}

	findNearestDate := findNearestDate(now, nd.Date, nextDates)
	return findNearestDate.Format(`20060102`), nil
}
func findRepeatIntervalWeeks(nd model.NextDate) (string, error) {
	now, err := timeNow(nd)
	if err != nil {
		return "", err
	}

	weekdayNumber := strings.TrimPrefix(nd.Repeat, "w ")
	weekDaysSlice := strings.Split(weekdayNumber, ",")

	for _, v := range weekDaysSlice {
		vi, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return "ошибка конвертации", err
		}
		if vi < MinWeek || vi > MAX_WEEK {
			return "нерпавильное значение повтора." + "необходимо w <через запятую от 1 до 7>.", fmt.Errorf("ошибка здесь: %v", nd.Repeat)
		}
		findWeekday, err := findNextWeekDay(now, nd.Date, vi)
		if err != nil {
			return "нерпавильное значение повтора.", fmt.Errorf("ошибка здесь: %v", nd.Repeat)
		}
		if findWeekday > nd.Date && findWeekday > now.Format(`20060102`) {
			return findWeekday, nil
		}
	}
	return "нерпавильное значение повтора." + "необходимо w <через запятую от 1 до 7>.", fmt.Errorf("ошибка здесь: %v", nd.Repeat)
}
func findRepeatIntervalYears(nd model.NextDate) (string, error) {
	now, err := timeNow(nd)
	if err != nil {
		return "", err
	}

	formattedNow := now.Format(`20060102`)
	searchDate := nd.Date

	for searchDate <= formattedNow || searchDate <= nd.Date {
		d, err := time.Parse(`20060102`, searchDate)
		if err != nil {
			return "неправильная дата", fmt.Errorf("ошибка здесь: %v", nd.Date)
		}
		searchDate = d.AddDate(1, 0, 0).Format(`20060102`)
	}
	return searchDate, nil
}
func findNextWeekDay(now time.Time, date string, weekday int) (string, error) {
	eventDate, err := time.Parse(`20060102`, date)
	if err != nil {
		return "", fmt.Errorf("неправильная дата события: %v", err)
	}
	currentWeekday := int(now.Weekday())
	daysUntilWeekday := (weekday - currentWeekday + OneWeek) % OneWeek
	nextWeekday := now.AddDate(0, 0, daysUntilWeekday)

	if nextWeekday.Before(eventDate) {
		nextWeekday = eventDate.AddDate(0, 0, (OneWeek-currentWeekday+weekday)%OneWeek)
	}

	return nextWeekday.Format(`20060102`), nil
}
func findDayOfMonth(now time.Time, date string, month, repeat int) time.Time {
	var searchDay time.Time
	maxDate, err := time.Parse(`20060102`, date)
	if err != nil {
		log.Fatal(err)
	}

	if maxDate.Before(now) {
		maxDate = now
	}

	if month == 0 {
		month = int(maxDate.Month())
	}

	lastDayOfMonth := lastDayMonth(maxDate.Year(), time.Month(month))
	if repeat > lastDayOfMonth {
		searchDay = time.Date(maxDate.Year(), time.Month(month+AddingOneMounth), repeat, 0, 0, 0, 0, time.UTC)
	} else if repeat < lastDayOfMonth && repeat > 0 {
		searchDay = time.Date(maxDate.Year(), time.Month(month), repeat, 0, 0, 0, 0, time.UTC)
	} else if repeat < 0 {
		searchDay = time.Date(maxDate.Year(), time.Month(month), lastDayOfMonth+AddingOneMounth, 0, 0, 0, 0, time.UTC)
		searchDay = searchDay.AddDate(0, 0, repeat)
	}

	if searchDay.Before(maxDate) {
		searchDay = searchDay.AddDate(0, AddingOneMounth, 0)
	}

	return searchDay
}
func lastDayMonth(year int, month time.Month) int {
	nextMonth := time.Date(year, month+AddingOneMounth, FirstDay, 0, 0, 0, 0, time.UTC)
	lastDayOfMonth := nextMonth.AddDate(0, 0, MinusOneDay)
	return lastDayOfMonth.Day()
}
func findNearestDate(now time.Time, date string, dates []time.Time) time.Time {
	if len(dates) == 1 {
		return dates[0]
	}

	var nearestDate time.Time
	dateToDate, err := time.Parse(`20060102`, date)
	if err != nil {
		fmt.Println(err)
	}
	minDifference := math.MaxInt64

	for _, d := range dates {
		if d.After(now) && d.After(dateToDate) {
			difference := int(d.Sub(now).Seconds())
			if difference < minDifference {
				minDifference = difference
				nearestDate = d
			}
		}
	}
	return nearestDate
}
func timeNow(nd model.NextDate) (time.Time, error) {
	var now time.Time
	if nd.Now == "" {
		now = time.Now()
	}
	now, err := time.Parse(`20060102`, nd.Now)
	if err != nil {
		return time.Time{}, fmt.Errorf("непонятно: %v", nd.Now)
	}
	return now, nil
}
func typeSearch(str string) int {
	if str == "" {
		return 0
	}
	_, err := time.Parse(`20060102`, str)
	if err == nil {
		return 1
	}
	return 2
}
