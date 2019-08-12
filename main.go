package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"
)

type config struct {
	searchWord string
	reaction   string
	username   string
	fromSet    bool
	from       time.Time
	toSet      bool
	to         time.Time
	debug      bool
}

var conf config

type searchCondition struct {
	searchWord string
	reaction   string
	username   string
	from       string
	to         string
	debug      bool
}

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

func search(sc searchCondition) ([]searchResult, error) {
	query := ""
	if sc.searchWord != "" {
		query = sc.searchWord
	}
	if sc.reaction != "" {
		query = fmt.Sprintf("%s has:%s", query, sc.reaction)
	}
	if sc.username != "" {
		query = fmt.Sprintf("%s from:%s", query, sc.username)
	}
	if sc.from != "" {
		query = fmt.Sprintf("%s after:%s", query, sc.from)
	}
	if sc.to != "" {
		query = fmt.Sprintf("%s before:%s", query, sc.to)
	}
	if conf.debug {
		fmt.Fprintf(os.Stderr, "query: %s\n", query)
	}
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

	if messages.Paging.Pages >= 2 {
		for i := 2; i <= messages.Paging.Pages; i++ {
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

func parseOptions() error {
	var (
		from string
		to   string
	)
	flag.StringVar(&conf.searchWord, "search-word", "", "search word")
	flag.StringVar(&conf.reaction, "reaction", "", "emoji string. eg) :ok_woman:")
	flag.StringVar(&conf.username, "username", "", "slack username")
	flag.StringVar(&from, "from", "", "message timestamp from")
	flag.StringVar(&to, "to", "", "message timestamp to")
	flag.BoolVar(&conf.debug, "debug", false, "debug mode")

	flag.Parse()

	format := "2006-01-02"
	loc, _ := time.LoadLocation("Asia/Tokyo")
	if from != "" {
		f, err := time.ParseInLocation(format, from, loc)
		if err != nil {
			return err
		}
		conf.fromSet = true
		conf.from = f
	}
	if to != "" {
		t, err := time.ParseInLocation(format, to, loc)
		if err != nil {
			return err
		}
		conf.toSet = true
		conf.to = t
	}

	return nil
}

func generateSearchCondition() searchCondition {
	fromValue := ""
	if conf.fromSet {
		fromValue = conf.from.Format("2006-01-02")
	}
	toValue := ""
	if conf.toSet {
		toValue = conf.to.Format("2006-01-02")
	}

	return searchCondition{
		searchWord: conf.searchWord,
		reaction:   conf.reaction,
		username:   conf.username,
		from:       fromValue,
		to:         toValue,
	}
}

func main() {
	parseOptions()

	sc := generateSearchCondition()
	if conf.debug {
		fmt.Fprintf(os.Stderr, "searchCondition: %v\n", sc)
	}
	result, err := search(sc)
	if err != nil {
		panic(err)
	}

	if sc.searchWord != "" && sc.reaction != "" {
		fmt.Printf("%s 〜 %s の期間に\n", sc.from, sc.to)
		fmt.Printf("%s さん が %s と言い、 %s を押された発言は %d 回です。\n", sc.username, sc.searchWord, sc.reaction, len(result))
	} else if sc.searchWord != "" {
		fmt.Printf("%s 〜 %s の期間に\n", sc.from, sc.to)
		fmt.Printf("%s さん が %s と言った回数は %d 回です。\n", sc.username, sc.searchWord, len(result))
	} else if sc.reaction != "" {
		fmt.Printf("%s 〜 %s の期間に\n", sc.from, sc.to)
		fmt.Printf("%s さん が %s を押された発言は %d 個です。\n", sc.username, sc.reaction, len(result))
	}

	for i, sr := range result {
		fmt.Println("--------------------------------------")
		fmt.Printf("No: %d %v in %s\n", i+1, sr.datetime, sr.channel)
		fmt.Printf("link: %s\n", sr.permalink)
		fmt.Printf("%s\n", strings.Replace(sr.text, "```", "---", -1))
		fmt.Println()
	}
}
