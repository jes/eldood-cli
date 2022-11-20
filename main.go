package main

// TODO:
// - command line argument to specify custom eldood instance (instead of https://eldood.uk)
// - config file for default eldood instance
// - some way to participate in the poll rather than just print the status
// - a better orange
// - when the table is too wide to fit into the terminal, chunk up the dates and draw several tables?
// - make colour optional
// - separate data fetching & json parsing from output formatting
// - unit tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Response struct {
	name          string
	okDates       []string
	ifneedbeDates []string
}

var AnsiGreen = "\033[32m\033[7m"
var AnsiOrange = "\033[33m\033[7m"
var AnsiReset = "\033[0m"

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: eldood TOKEN\n")
		os.Exit(1)
	}
	token := os.Args[1]

	resp, err := http.Get("https://eldood.uk/" + token + "/json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "http get: %v\n", err)
		os.Exit(1)
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json decode: %v\n", err)
		os.Exit(1)
	}

	if result["status"].(string) != "ok" {
		fmt.Fprintf(os.Stderr, "bad status (is the token '%s' correct?)\n", token)
		os.Exit(1)
	}

	var dates []string
	for _, r := range result["dates"].([]interface{}) {
		dates = append(dates, r.(string))
	}

	var responses []Response
	maxNameLength := 0

	for _, r1 := range result["responses"].([]interface{}) {
		r := r1.(map[string]interface{})
		name := r["name"].(string)
		okDates := r["ok_dates"].([]interface{})
		ifneedbeDates := r["ifneedbe_dates"].([]interface{})

		if len(name) > maxNameLength {
			maxNameLength = len(name)
		}

		responses = append(responses, Response{
			name:          name,
			okDates:       toStringSlice(okDates),
			ifneedbeDates: toStringSlice(ifneedbeDates),
		})
	}

	fmt.Println(result["name"].(string))
	fmt.Println(result["descr"].(string))
	fmt.Println()

	weekdays := spaces(maxNameLength + 1)
	monthdays := spaces(maxNameLength + 1)
	months := spaces(maxNameLength + 1)
	dateLayout := "20060102"
	for _, date := range dates {
		time, err := time.Parse(dateLayout, date)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse date %s: %v\n", date, err)
			os.Exit(1)
		}
		weekdays = weekdays + fmt.Sprintf(" %4s", time.Weekday().String()[0:3])
		monthdays = monthdays + fmt.Sprintf("  %02d ", time.Day())
		months = months + fmt.Sprintf(" %4s", time.Month().String()[0:3])
	}
	fmt.Println(weekdays)
	fmt.Println(monthdays)
	fmt.Println(months)
	fmt.Println()
	okCount := make(map[string]int)
	ifneedbeCount := make(map[string]int)
	for _, r := range responses {
		fmt.Print(spaces(maxNameLength - len(r.name)))
		fmt.Print(r.name)
		fmt.Print("  ")

		for _, date := range dates {
			if r.okDate(date) {
				fmt.Print(AnsiGreen)
				fmt.Print("  \u2713  ")
				fmt.Print(AnsiReset)
				if _, exists := okCount[date]; exists {
					okCount[date]++
				} else {
					okCount[date] = 1
				}
			} else if r.ifneedbeDate(date) {
				fmt.Print(AnsiOrange)
				fmt.Print(" (\u2713) ")
				fmt.Print(AnsiReset)
				if _, exists := ifneedbeCount[date]; exists {
					ifneedbeCount[date]++
				} else {
					ifneedbeCount[date] = 1
				}
			} else {
				fmt.Print(spaces(5))
			}
		}

		fmt.Println()
	}
	fmt.Println()
	oks := spaces(maxNameLength)
	ifneedbes := spaces(maxNameLength)
	for _, date := range dates {
		oks = oks + fmt.Sprintf("   %2d", okCount[date])
		ifneedbeString := ""
		if ifneedbeCount[date] != 0 {
			ifneedbeString = fmt.Sprintf("+%d", ifneedbeCount[date])
		}
		ifneedbes = ifneedbes + fmt.Sprintf("  %3s", ifneedbeString)
	}
	fmt.Println(oks)
	fmt.Println(ifneedbes)
}

func toStringSlice(in []interface{}) []string {
	var out []string
	for _, v := range in {
		out = append(out, v.(string))
	}
	return out
}

func spaces(n int) string {
	var out string
	for i := 0; i < n; i++ {
		out = out + " "
	}
	return out
}

func (r Response) okDate(date string) bool {
	for _, d := range r.okDates {
		if d == date {
			return true
		}
	}
	return false
}

func (r Response) ifneedbeDate(date string) bool {
	for _, d := range r.ifneedbeDates {
		if d == date {
			return true
		}
	}
	return false
}
