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

var ctx = context.Background()
var config = loadConfig()

// YamlConfig Configuration options from config.yml
type YamlConfig struct {
	Owner            string `yaml:"owner"`
	Repository       string `yaml:"repository"`
	Branch           string `yaml:"branch"`
	SpreadsheetID    string `yaml:"spreadsheetID"`
	ColumnRange      string `yaml:"columnRange"`
	GoTimeFormat     string `yaml:"goTimeFormat"`
	CommitTrimLength int    `yaml:"commitTrimLength"`
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
	if len(message) > config.CommitTrimLength {
		message = message[0:config.CommitTrimLength]
	}
	return message
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

func getCommitStatus(githubClient *github.Client, ref string) int {
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

func getCommits(githubClient *github.Client, fromDate time.Time) []*github.RepositoryCommit {
	fromDateBOD := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, time.Local)
	fromDateEOD := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 23, 59, 0, 0, time.Local)

	options := github.CommitsListOptions{
		SHA:   config.Branch,
		Since: fromDateBOD,
		Until: fromDateEOD,
	}
	commits, _, er := githubClient.Repositories.ListCommits(ctx, config.Owner, config.Repository, &options)
	handleError(er)
	return commits
}

func formatCommits(githubClient *github.Client, commits []*github.RepositoryCommit, dateToPull time.Time) [][]interface{} {
	orderedCommits := make([][]interface{}, 100)

	for i, j := len(commits), 0; i >= 1; i, j = i-1, j+1 {
		orderedCommits[i-1] = []interface{}{
			"",
			"",
			dateToPull.Format(config.GoTimeFormat),
			trimCommitMessage(commits[j].Commit.GetMessage()),
			getCommitStatus(githubClient, commits[j].GetSHA()),
		}
	}
	return orderedCommits
}

func main() {
	log.Printf("Warming Xerox machine...")

	githubClient := gitclient.Client(ctx)
	spreadsheetService := sheetsclient.Client()
	dateToPull := time.Now().AddDate(0, 0, -1)

	commits := getCommits(githubClient, dateToPull)
	formattedCommits := formatCommits(githubClient, commits, dateToPull)

	valueRange := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         formattedCommits,
	}

	_, err := spreadsheetService.
		Spreadsheets.
		Values.
		Append(config.SpreadsheetID, config.ColumnRange, valueRange).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Do()
	handleError(err)

	log.Printf("Updated spreadsheet.")
}
