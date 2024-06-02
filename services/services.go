package services

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func NextDate(now time.Time, dateStr string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("повтор пуст")
	}

	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		return "", errors.New("неверный формат даты")
	}

	switch {
	case repeat == "y":
		return calcNextYearDate(now, date)

	case strings.HasPrefix(repeat, "d "):
		days, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(repeat, "d ")))
		if err != nil || days <= 0 || days > 400 {
			return "", errors.New("неверное значение дней")
		}
		return calcNextDayDate(now, date, days)

	case strings.HasPrefix(repeat, "w "):
		daysOfWeek, err := parseDaysOfWeek(strings.TrimSpace(strings.TrimPrefix(repeat, "w ")))
		if err != nil {
			return "", err
		}
		return getNextWeekdayDate(now, date, daysOfWeek)

	case strings.HasPrefix(repeat, "m "):
		daysOfMonth, months, err := parseDaysOfMonthAndMonths(strings.TrimSpace(strings.TrimPrefix(repeat, "m ")))
		if err != nil {
			return "", err
		}
		return getNextMonthDate(now, date, daysOfMonth, months)

	default:
		return "", errors.New("неподдерживаемый формат повторения задач")
	}
}

func calcNextYearDate(now, date time.Time) (string, error) {
	// Повторение каждый год
	date = date.AddDate(1, 0, 0)
	for date.Before(now) {
		date = date.AddDate(1, 0, 0)
	}
	return date.Format("20060102"), nil
}

func calcNextDayDate(now, date time.Time, days int) (string, error) {
	// Повторение каждый N день
	date = date.AddDate(0, 0, days)
	for date.Before(now) {
		date = date.AddDate(0, 0, days)
	}
	return date.Format("20060102"), nil
}

func parseDaysOfWeek(days string) ([]time.Weekday, error) {
	parts := strings.Split(days, ",")
	var daysOfWeek []time.Weekday
	for _, part := range parts {
		day, err := strconv.Atoi(part)
		if err != nil || day < 1 || day > 7 {
			return nil, errors.New("неверное значение дня недели")
		}
		daysOfWeek = append(daysOfWeek, time.Weekday((day % 7)))
	}
	return daysOfWeek, nil
}

func getNextWeekdayDate(now, startDate time.Time, daysOfWeek []time.Weekday) (string, error) {
	date := startDate
	for {
		for _, day := range daysOfWeek {
			if date.Weekday() == day {
				if date.After(now) {
					return date.Format("20060102"), nil
				}
			}
		}
		date = date.AddDate(0, 0, 1)
	}
}

func parseDaysOfMonthAndMonths(input string) ([]int, []int, error) {
	parts := strings.Split(input, " ")
	if len(parts) < 1 || len(parts) > 2 {
		return nil, nil, errors.New("неверный формат дня месяца")
	}

	daysOfMonth, err := parseDaysOfMonth(parts[0])
	if err != nil {
		return nil, nil, err
	}

	var months []int
	if len(parts) == 2 {
		months, err = parseMonths(parts[1])
		if err != nil {
			return nil, nil, err
		}
	}
	return daysOfMonth, months, nil
}

func parseDaysOfMonth(days string) ([]int, error) {
	parts := strings.Split(days, ",")
	var daysOfMonth []int
	for _, part := range parts {
		day, err := strconv.Atoi(part)
		if err != nil || day < -31 || day > 31 || day == 0 {
			return nil, errors.New("неверное значение дня месяца")
		}
		daysOfMonth = append(daysOfMonth, day)
	}
	return daysOfMonth, nil
}

func parseMonths(months string) ([]int, error) {
	parts := strings.Split(months, ",")
	var monthList []int
	for _, part := range parts {
		month, err := strconv.Atoi(part)
		if err != nil || month < 1 || month > 12 {
			return nil, errors.New("неверное значение месяца")
		}
		monthList = append(monthList, month)
	}
	return monthList, nil
}

func getNextMonthDate(now, startDate time.Time, daysOfMonth, months []int) (string, error) {
	date := startDate.AddDate(0, 1, 0) // прибавляем месяц сразу
	for {
		if len(months) == 0 || contains(months, int(date.Month())) {
			for _, day := range daysOfMonth {
				nextDate := getDateForDay(date, day)
				if nextDate.After(now) {
					return nextDate.Format("20060102"), nil
				}
			}
		}
		date = date.AddDate(0, 1, 0)
	}
}

func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func getDateForDay(date time.Time, day int) time.Time {
	year, month, _ := date.Date()
	if day > 0 {
		return time.Date(year, month, day, 0, 0, 0, 0, date.Location())
	}
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, date.Location()).Day()
	return time.Date(year, month, lastDay+day+1, 0, 0, 0, 0, date.Location())
}
