package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func HandleNextDate(w http.ResponseWriter, r *http.Request) {
	currentDateStr := r.URL.Query().Get("now")
	initialDateStr := r.URL.Query().Get("date")
	repeatInterval := r.URL.Query().Get("repeat")

	currentDate, err := time.Parse(formatDate, currentDateStr)
	if err != nil {
		http.Error(w, "Invalid 'now' date format", http.StatusBadRequest)
		return
	}

	nextCalculatedDate, err := CalculateNextDate(currentDate, initialDateStr, repeatInterval)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "%s", nextCalculatedDate)
}

func CalculateNextDate(current time.Time, initial string, interval string) (string, error) {
	if interval == "" {
		return "", fmt.Errorf("repeat is empty")
	}

	validInitialDate, err := time.Parse(formatDate, initial)
	if err != nil {
		return "", fmt.Errorf("incorrect date format")
	}

	rules := strings.Split(interval, " ")
	ruleType := rules[0]

	switch ruleType {
	case "d":
		if len(rules) < 2 {
			return "", fmt.Errorf("no days")
		}
		days, err := strconv.Atoi(rules[1])
		if err != nil || days < 0 || days > 400 {
			return "", fmt.Errorf("Days must be a positive integer between 0 and 400")
		}
		return incrementDays(current, validInitialDate, days), nil

	case "y":
		return incrementYear(current, validInitialDate), nil

	case "w":
		if len(rules) < 2 {
			return "", fmt.Errorf("no days")
		}
		return incrementWeek(current, validInitialDate, rules[1])

	case "m":
		if len(rules) < 2 {
			return "", fmt.Errorf("no days")
		}
		return incrementMonth(current, validInitialDate, rules)

	default:
		return "", fmt.Errorf("Invalid rule type")
	}
}

func incrementDays(current, validInitial time.Time, days int) string {
	if validInitial.Equal(current) {
		return current.Format(formatDate)
	}
	validInitial = validInitial.AddDate(0, 0, days)
	for validInitial.Before(current) {
		validInitial = validInitial.AddDate(0, 0, days)
	}
	return validInitial.Format(formatDate)
}

func incrementYear(current, validInitial time.Time) string {
	validInitial = validInitial.AddDate(1, 0, 0)
	for validInitial.Before(current) {
		validInitial = validInitial.AddDate(1, 0, 0)
	}
	return validInitial.Format(formatDate)
}

func incrementWeek(current time.Time, validInitial time.Time, days string) (string, error) {
	daysOfWeek := strings.Split(days, ",")
	weekDaysMap := make(map[int]bool)

	for _, day := range daysOfWeek {
		dayInt, err := strconv.Atoi(day)
		if err != nil || dayInt < 1 || dayInt > 7 {
			return "", fmt.Errorf("Invalid day of the week")
		}
		if dayInt == 7 {
			dayInt = 0
		}
		weekDaysMap[dayInt] = true
	}

	for {
		if weekDaysMap[int(validInitial.Weekday())] && current.Before(validInitial) {
			break
		}
		validInitial = validInitial.AddDate(0, 0, 1)
	}

	return validInitial.Format(formatDate), nil
}

func incrementMonth(current time.Time, validInitial time.Time, repeat []string) (string, error) {
	daysArray := strings.Split(repeat[1], ",")
	var monthsArray []string
	if len(repeat) > 2 {
		monthsArray = strings.Split(repeat[2], ",")
	}

	daysMap := map[int]bool{}
	for _, day := range daysArray {
		dayInt, err := strconv.Atoi(day)
		if err != nil || dayInt < -2 || dayInt > 31 || dayInt == 0 {
			return "", fmt.Errorf("Invalid day of the month")
		}
		daysMap[dayInt] = true
	}

	monthsMap := map[int]bool{}
	for _, month := range monthsArray {
		if month == "" {
			continue
		}
		monthInt, err := strconv.Atoi(month)
		if err != nil || monthInt < 1 || monthInt > 12 {
			return "", fmt.Errorf("Invalid month")
		}
		monthsMap[monthInt] = true
	}

	for {
		if len(monthsMap) == 0 || monthsMap[int(validInitial.Month())] {
			lastDayOfMonth := time.Date(validInitial.Year(), validInitial.Month()+1, 0, 0, 0, 0, 0, validInitial.Location()).Day()
			secondLastDay := lastDayOfMonth - 1
			if daysMap[secondLastDay] || daysMap[lastDayOfMonth] {
				break
			}
			for day := range daysMap {
				if day <= lastDayOfMonth && day > 0 {
					break
				}
			}
			validInitial = validInitial.AddDate(0, 1, 0)
			continue
		}

		validInitial = validInitial.AddDate(0, 1, 0)
	}

	return validInitial.Format(formatDate), nil
}
