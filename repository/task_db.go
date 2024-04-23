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
	ErrInvalidRepeatValue              = "Некорректное значение повтора."
	ErrRepeat                          = "Обрати внимание вот сюда: %v."
	AllowedDaysRange                   = "Допускается от 1 до 400 дней."
	AllowedWeekdaysRange               = "Допускается w <через запятую от 1 до 7>."
	AllowedDaysAndMonthRange           = "m <через запятую от 1 до 31,-1,-2> [через запятую от 1 до 12]"
	InvalidDateMessage                 = "Некорректная дата."
	UnableToConvertRepeatToNumberError = "Не удалось сконвертировать повтор в число."
	ErrorParseDate                     = "Не смог разобрать вот это: %s."
	MonthNumberRange                   = "Допускается m <через запятую от 1 до 12>."
	WrongDate                          = "некорректная дата создания события: %v"
	WrongRepeat                        = "Проблемы с форматом повтора, погляди сюда повтор: %v"
	TaskNameRequiredErrorMessage       = "Название задачи не может быть пустым."
	TaskNotFoundErrorMessage           = "Задача не найдена"
	TaskIdRequiredErrorMessage         = "Не указан идентификатор"
	InvalidTaskIdErrorMessage          = "Некорректный идентификатор"
	PrefixRepeatM                      = "m "
	PrefixRepeatD                      = "d "
	PrefixRepeatW                      = "w "
	PrefixDay                          = 'd'
	PrefixWeek                         = 'w'
	PrefixMonth                        = 'm'
	PrefixYear                         = 'y'
	SeparatorSpace                     = " "
	SeparatorComma                     = ","
	Format_yyyymmdd                    = `20060102`
	Format_dd_mm_yyyy                  = `02.01.2006`
	FirstDay                           = 1
	MinusOneDay                        = -1
	AddingOneMounth                    = 1
	MinRepeatIntervalDays              = 1
	MaxRepeatIntervalDay               = 31
	MinMinusRepeatIntervalDay          = -2
	MaxRepeatIntervalDays              = 400
	MinMonths                          = 1
	MaxMonths                          = 12
	MinWeek                            = 1
	MAX_WEEK                           = 7
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
		return "", fmt.Errorf(WrongRepeat, nd.Repeat)
	}

	if !regexp.MustCompile(`^([wdm]\s.*|y)?$`).MatchString(nd.Repeat) {
		return "", fmt.Errorf(WrongRepeat, nd.Repeat)
	}

	switch nd.Repeat[0] {
	case PrefixDay:
		repeatIntervalDays, err := findRepeatIntervalDays(nd)
		return repeatIntervalDays, err
	case PrefixYear:
		repeatIntervalYears, err := findRepeatIntervalYears(nd)
		return repeatIntervalYears, err
	case PrefixWeek:
		repeatIntervalWeeks, err := findRepeatIntervalWeeks(nd)
		return repeatIntervalWeeks, err
	case PrefixMonth:
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

	query := fmt.Sprintf("INSERT INTO %s (title, comment, date, repeat) VALUES ($1, $2, $3, $4) RETURNING id", TaskTable)
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
		query = fmt.Sprintf("SELECT * FROM %s ORDER BY date LIMIT ?", TaskTable)
		err := t.db.Select(&tasks, query, LimitTasks)
		if err != nil {
			return model.ListTasks{}, err
		}
	case 1:
		s, _ := time.Parse(Format_dd_mm_yyyy, search)
		st := s.Format(Format_yyyymmdd)
		query = fmt.Sprintf("SELECT * FROM %s WHERE date = ? ORDER BY date LIMIT ?", TaskTable)
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
		return model.Task{}, fmt.Errorf(TaskIdRequiredErrorMessage)
	}
	if _, err := strconv.Atoi(id); err != nil {
		return model.Task{}, fmt.Errorf(InvalidTaskIdErrorMessage)
	}
	var task model.Task
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", TaskTable)
	err := t.db.Get(&task, query, id)
	if err != nil {
		return model.Task{}, fmt.Errorf(TaskNotFoundErrorMessage)
	}
	return task, err
}
func (t *TodoTaskSqlite) UpdateTask(task model.Task) error {
	err := t.checkTask(&task)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("UPDATE %s SET title = ?, comment = ?, date = ?, repeat = ? WHERE id = ?", TaskTable)
	_, err = t.db.Exec(query, task.Title, task.Comment, task.Date, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf(TaskNotFoundErrorMessage)
	}
	return nil
}
func (t *TodoTaskSqlite) DeleteTask(id string) error {
	_, err := t.GetTaskById(id)
	if err != nil {
		return err
	}
	queryDelete := fmt.Sprintf("DELETE FROM %s WHERE id = ?", TaskTable)
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
		queryDeleteTask := fmt.Sprintf("DELETE FROM %s WHERE id = ?", TaskTable)
		logrus.Println(queryDeleteTask)
		t.db.Exec(queryDeleteTask, id)
		return nil
	}

	nd := model.NextDate{
		Date:   task.Date,
		Now:    time.Now().Format(Format_yyyymmdd),
		Repeat: task.Repeat,
	}

	newDate, err := t.NextDate(nd)
	if err != nil {
		return err
	}

	task.Date = newDate
	queryUpdateTask := fmt.Sprintf("UPDATE %s SET date = ? WHERE id = ?", TaskTable)
	logrus.Println(queryUpdateTask)
	_, err = t.db.Exec(queryUpdateTask, task.Date, id)
	if err != nil {
		return err
	}
	return nil

}
func (t *TodoTaskSqlite) checkTask(task *model.Task) error {
	if task.Title == "" {
		return fmt.Errorf(TaskNameRequiredErrorMessage)
	}

	if !regexp.MustCompile(`^([wdm]\s.*|y)?$`).MatchString(task.Repeat) {
		return fmt.Errorf(WrongRepeat, task.Repeat)
	}

	now := time.Now().Format(Format_yyyymmdd)

	if task.Date == "" {
		task.Date = now
	}

	_, err := time.Parse(Format_yyyymmdd, task.Date)
	if err != nil {
		return fmt.Errorf(InvalidDateMessage)
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
	stringRepeatIntervalDays := strings.TrimPrefix(nd.Repeat, PrefixRepeatD)
	repeatIntervalDays, err := strconv.Atoi(stringRepeatIntervalDays)
	if err != nil {
		return ErrInvalidRepeatValue, fmt.Errorf(ErrRepeat, nd.Repeat)
	}
	if repeatIntervalDays < MinRepeatIntervalDays || repeatIntervalDays > MaxRepeatIntervalDays {
		return ErrInvalidRepeatValue + AllowedDaysRange, fmt.Errorf(ErrRepeat, nd.Repeat)
	}
	searchDate := nd.Date

	for searchDate <= now.Format(Format_yyyymmdd) || searchDate <= nd.Date {
		d, err := time.Parse(Format_yyyymmdd, searchDate)
		if err != nil {
			return InvalidDateMessage, fmt.Errorf(ErrorParseDate, nd.Date)
		}
		searchDate = d.AddDate(0, 0, repeatIntervalDays).Format(Format_yyyymmdd)
	}
	return searchDate, nil
}
func findRepeatIntervalMonths(nd model.NextDate) (string, error) {
	now, err := timeNow(nd)
	if err != nil {
		return "", err
	}
	repeatSrt := strings.TrimPrefix(nd.Repeat, PrefixRepeatM)
	isConstainsNumMonth := strings.Contains(repeatSrt, SeparatorSpace)
	monthsSlice := make([]string, 0)
	months := make([]int, 0)
	if isConstainsNumMonth {
		IndexSep := strings.Index(repeatSrt, SeparatorSpace)
		repeatSrtMounth := repeatSrt[IndexSep+1:]
		repeatSrt = repeatSrt[:IndexSep]
		monthsSlice = strings.Split(repeatSrtMounth, SeparatorComma)

		for _, v := range monthsSlice {
			vi, err := strconv.Atoi(strings.TrimSpace(v))
			if err != nil {
				return ErrInvalidRepeatValue, err
			}
			if vi < MinMonths || vi > MaxMonths {
				return ErrInvalidRepeatValue + MonthNumberRange,
					fmt.Errorf(ErrRepeat, nd.Repeat)
			}
			months = append(months, vi)
		}
	}

	monthDaysSlice := strings.Split(repeatSrt, SeparatorComma)

	monthDays := make([]int, 0)

	for _, v := range monthDaysSlice {
		vi, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return UnableToConvertRepeatToNumberError, err
		}
		if vi < MinMinusRepeatIntervalDay || vi > MaxRepeatIntervalDay {
			return ErrInvalidRepeatValue + AllowedDaysAndMonthRange, fmt.Errorf(ErrRepeat, nd.Repeat)
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
	return findNearestDate.Format(Format_yyyymmdd), nil
}
func findRepeatIntervalWeeks(nd model.NextDate) (string, error) {
	now, err := timeNow(nd)
	if err != nil {
		return "", err
	}

	weekdayNumber := strings.TrimPrefix(nd.Repeat, PrefixRepeatW)
	weekDaysSlice := strings.Split(weekdayNumber, SeparatorComma)

	for _, v := range weekDaysSlice {
		vi, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return UnableToConvertRepeatToNumberError, err
		}
		if vi < MinWeek || vi > MAX_WEEK {
			return ErrInvalidRepeatValue + AllowedWeekdaysRange, fmt.Errorf(ErrRepeat, nd.Repeat)
		}
		findWeekday, err := findNextWeekDay(now, nd.Date, vi)
		if err != nil {
			return ErrInvalidRepeatValue, fmt.Errorf(ErrRepeat, nd.Repeat)
		}
		if findWeekday > nd.Date && findWeekday > now.Format(Format_yyyymmdd) {
			return findWeekday, nil
		}
	}
	return ErrInvalidRepeatValue + AllowedWeekdaysRange, fmt.Errorf(ErrRepeat, nd.Repeat)
}
func findRepeatIntervalYears(nd model.NextDate) (string, error) {
	now, err := timeNow(nd)
	if err != nil {
		return "", err
	}

	formattedNow := now.Format(Format_yyyymmdd)
	searchDate := nd.Date

	for searchDate <= formattedNow || searchDate <= nd.Date {
		d, err := time.Parse(Format_yyyymmdd, searchDate)
		if err != nil {
			return InvalidDateMessage, fmt.Errorf(ErrRepeat, nd.Date)
		}
		searchDate = d.AddDate(1, 0, 0).Format(Format_yyyymmdd)
	}
	return searchDate, nil
}
func findNextWeekDay(now time.Time, date string, weekday int) (string, error) {
	eventDate, err := time.Parse(Format_yyyymmdd, date)
	if err != nil {
		return "", fmt.Errorf(WrongDate, err)
	}
	currentWeekday := int(now.Weekday())
	daysUntilWeekday := (weekday - currentWeekday + OneWeek) % OneWeek
	nextWeekday := now.AddDate(0, 0, daysUntilWeekday)

	if nextWeekday.Before(eventDate) {
		nextWeekday = eventDate.AddDate(0, 0, (OneWeek-currentWeekday+weekday)%OneWeek)
	}

	return nextWeekday.Format(Format_yyyymmdd), nil
}
func findDayOfMonth(now time.Time, date string, month, repeat int) time.Time {
	var searchDay time.Time
	maxDate, err := time.Parse(Format_yyyymmdd, date)
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
	dateToDate, err := time.Parse(Format_yyyymmdd, date)
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
	now, err := time.Parse(Format_yyyymmdd, nd.Now)
	if err != nil {
		return time.Time{}, fmt.Errorf(ErrorParseDate, nd.Now)
	}
	return now, nil
}
func typeSearch(str string) int {
	if str == "" {
		return 0
	}
	_, err := time.Parse(Format_dd_mm_yyyy, str)
	if err == nil {
		return 1
	}
	return 2
}
