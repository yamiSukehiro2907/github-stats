package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github-stats/helpers"
	_ "github-stats/helpers"
	"log"
	"os"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func main() {
	token := os.Getenv("ACCESS_TOKEN")
	if token == "" {
		log.Fatal("ACCESS_TOKEN environment variable is required")
	}

	username := os.Getenv("GITHUB_USERNAME")
	if username == "" {
		log.Fatal("GITHUB_USERNAME environment variable is required")
	}

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	stats, err := helpers.FetchStats(client, username)
	if err != nil {
		log.Fatalf("Error fetching stats: %v", err)
	}

	languageStats := helpers.ProcessLanguageStats(stats)

	svg := helpers.GenerateSVG(stats, languageStats)

	if err := os.WriteFile("stats.svg", []byte(svg), 0644); err != nil {
		log.Fatalf("Error writing SVG: %v", err)
	}

	jsonData, _ := json.MarshalIndent(stats, "", "  ")
	if err := os.WriteFile("stats.json", jsonData, 0644); err != nil {
		log.Fatalf("Error writing JSON: %v", err)
	}

	fmt.Println("✅ Stats generated successfully!")
	fmt.Printf("Total Contributions: %d\n", stats.User.ContributionsCollection.ContributionCalendar.TotalContributions)
	fmt.Printf("Public Repositories: %d\n", stats.User.Repositories.TotalCount)
}
