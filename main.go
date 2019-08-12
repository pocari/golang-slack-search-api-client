package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"
)

type searchResult struct {
	channel   string
	username  string
	text      string
	datetime  time.Time
	permalink string
}

func slackTimestampToGolangTime(timestamp string) (time.Time, error) {
	if strings.Contains(timestamp, ".") {
		components := strings.Split(timestamp, ".")
		intValue, err := strconv.ParseInt(components[0], 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(intValue, 0), nil
	}
	intValue, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(intValue, 0), nil
}

func matchesSliceToSearchResultArray(result [][]slack.SearchMessage) ([]searchResult, error) {
	ret := []searchResult{}

	for _, searchMessages := range result {
		for _, sm := range searchMessages {
			t, err := slackTimestampToGolangTime(sm.Timestamp)
			if err != nil {
				return nil, err
			}
			ret = append(ret, searchResult{
				channel:   sm.Channel.Name,
				username:  sm.Username,
				text:      sm.Text,
				datetime:  t,
				permalink: sm.Permalink,
			})
		}
	}

	return ret, nil
}

func search() ([]searchResult, error) {
	// query := "疲れた from:@kuruma after:2019-07-01 before:2019-08-01"
	query := "疲れた from:@kuruma"
	api := slack.New(os.Getenv("SLACK_TOKEN"))
	result := [][]slack.SearchMessage{}
	sp := slack.SearchParameters{
		Count: 20,
	}

	messages, _, err := api.Search(query, sp)
	if err != nil {
		return nil, err
	}
	result = append(result, messages.Matches)

	// fmt.Printf("messages.Paging.Pages: %v\n", messages.Paging.Pages)
	if messages.Paging.Pages >= 2 {
		for i := 2; i <= messages.Paging.Pages; i++ {
			// fmt.Printf("page %v get\n", i)

			sp.Page = i
			messages, _, err := api.Search(query, sp)
			if err != nil {
				return nil, err
			}
			result = append(result, messages.Matches)
		}
	}
	return matchesSliceToSearchResultArray(result)
}

func main() {
	result, err := search()
	if err != nil {
		panic(err)
	}

	for i, sr := range result {
		fmt.Println("--------------------------------------")
		fmt.Printf("No: %d %v in %s\n", i, sr.datetime, sr.channel)
		fmt.Printf("link: %s\n", sr.permalink)
		fmt.Printf("%s\n", strings.Replace(sr.text, "```", "---", -1))
	}
}
