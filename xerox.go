package main

import (
	"context"
	"io/ioutil"
	"log"
	"time"

	"github.com/albertpaulp/xerox/gitclient"
	"github.com/albertpaulp/xerox/sheetsclient"
	"github.com/google/go-github/github"
	"google.golang.org/api/sheets/v4"
	"gopkg.in/yaml.v2"
)

const buildState = "success"

// YamlConfig Configuration options from config.yml
type YamlConfig struct {
	Owner         string `yaml:"owner"`
	Repository    string `yaml:"repository"`
	Branch        string `yaml:"branch"`
	SpreadsheetID string `yaml:"spreadsheetID"`
	ColumnRange   string `yaml:"columnRange"`
	GoTimeFormat  string `yaml:"goTimeFormat"`
}

func loadConfig() *YamlConfig {
	yamlFile, er := ioutil.ReadFile("config.yml")
	if er != nil {
		panic("Could not find config.yml file")
	}
	config := YamlConfig{}
	err := yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic("Could not parse configuration.yml file")
	}
	return &config
}

func trimCommitMessage(message string) string {
	if len(message) > 100 {
		message = message[0:100]
	}
	return message
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

func getCommitStatus(ctx context.Context, githubClient *github.Client, ref string) int {
	config := loadConfig()
	options := github.ListOptions{
		Page:    1,
		PerPage: 1,
	}
	statuses, _, err := githubClient.Repositories.ListStatuses(ctx, config.Owner, config.Repository, ref, &options)
	handleError(err)
	state := 0
	if statuses[0].GetState() != buildState {
		state = 1
	}
	return state
}

func main() {
	log.Printf("Warming Xerox machine...")

	context := context.Background()
	githubClient := gitclient.Client()
	spreadsheetService := sheetsclient.Client()
	config := loadConfig()

	yesterday := time.Now().AddDate(0, 0, -1)
	yesterdayBOD := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.Local)
	yesterdayEOD := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 0, 0, time.Local)

	options := github.CommitsListOptions{
		SHA:   config.Branch,
		Since: yesterdayBOD,
		Until: yesterdayEOD,
	}
	commits, _, er := githubClient.Repositories.ListCommits(context, config.Owner, config.Repository, &options)
	handleError(er)
	values := make([][]interface{}, 100)

	for i, j := len(commits), 0; i >= 1; i, j = i-1, j+1 {
		values[i-1] = []interface{}{
			"",
			"",
			commits[j].Commit.Author.Date.Format(config.GoTimeFormat),
			trimCommitMessage(commits[j].Commit.GetMessage()),
			getCommitStatus(context, githubClient, commits[j].GetSHA()),
		}
	}

	rb := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         values,
	}

	_, err := spreadsheetService.
		Spreadsheets.
		Values.
		Append(config.SpreadsheetID, config.ColumnRange, rb).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Do()
	handleError(err)

	log.Printf("Updated spreadsheet.")
}
