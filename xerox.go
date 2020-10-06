package main

import (
	"context"
	"io/ioutil"
	"log"
	"time"

	"github.com/albertpaulp/xerox/sheetsclient"

	"github.com/albertpaulp/xerox/gitclient"

	"gopkg.in/yaml.v2"

	"github.com/google/go-github/github"
	"google.golang.org/api/sheets/v4"
)

// YamlConfig Configuration options from config.yml
type YamlConfig struct {
	Owner         string `yaml:"owner"`
	Repository    string `yaml:"repository"`
	Branch        string `yaml:"branch"`
	SpreadsheetID string `yaml:"spreadsheetID"`
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

func main() {
	log.Printf("Warming up Xerox machine...")

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

	for index, commit := range commits {
		values[index] = []interface{}{
			"",
			"",
			commit.Commit.Author.Date.Format("02.01.2006"),
			trimCommitMessage(commit.Commit.GetMessage()),
		}
	}

	rb := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         values,
	}

	_, err := spreadsheetService.
		Spreadsheets.
		Values.
		Append(config.SpreadsheetID, "C:F", rb).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Do()
	handleError(err)

	log.Printf("Updated spreadsheet.")
}
