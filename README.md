# slack search API client

## install

```
go get github.com/pocari/golang-slack-search-api-client
```

## slack api token

```
export SLACK_TOKEN=xxxxxx
```

## run command
```
Usage of bin/golang-slack-search-api-client:
  -debug
        debug mode
  -from string
        message timestamp from
  -reaction string
        emoji string. eg) :ok_woman:
  -search-word string
        search word
  -to string
        message timestamp to
  -username string
        slack username

```

### example

```
bin/golang-slack-search-api-client -search-word test -username uname -from 2019-07-01 -to 2019-08-11
```
