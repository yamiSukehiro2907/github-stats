package helpers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/shurcooL/githubv4"
)

func FetchStats(client *githubv4.Client, username string) (*UserStats, error) {
	var query UserStats
	variables := map[string]interface{}{
		"username": githubv4.String(username),
	}

	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		return nil, err
	}

	return &query, nil
}

func ProcessLanguageStats(stats *UserStats) []LanguageStat {
	languageMap := make(map[string]*LanguageStat)
	totalSize := 0

	for _, repo := range stats.User.Repositories.Nodes {
		for _, edge := range repo.Languages.Edges {
			name := edge.Node.Name
			if _, exists := languageMap[name]; !exists {
				languageMap[name] = &LanguageStat{
					Name:  name,
					Color: edge.Node.Color,
				}
			}
			languageMap[name].Size += edge.Size
			totalSize += edge.Size
		}
	}

	var languages []LanguageStat
	for _, lang := range languageMap {
		if lang.Name == "HTML" || lang.Name == "CSS" {
			continue
		}
		lang.Percentage = float64(lang.Size) / float64(totalSize) * 100
		languages = append(languages, *lang)
	}

	sort.Slice(languages, func(i, j int) bool {
		return languages[i].Size > languages[j].Size
	})

	if len(languages) > 6 {
		languages = languages[:6]
	}

	return languages
}

func GenerateSVG(stats *UserStats, languages []LanguageStat) string {
	cc := stats.User.ContributionsCollection
	repos := stats.User.Repositories

	totalStars := 0
	totalForks := 0
	for _, repo := range repos.Nodes {
		totalStars += repo.StargazerCount
		totalForks += repo.ForkCount
	}

	var sb strings.Builder

	sb.WriteString(`<svg width="500" height="300" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="grad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#58a6ff;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#1f6feb;stop-opacity:1" />
    </linearGradient>
  </defs>
  
  <!-- Background -->
  <rect width="500" height="300" fill="#0d1117" rx="10"/>
  
  <!-- Title -->
  <text x="250" y="30" fill="url(#grad)" font-size="20" font-weight="bold" text-anchor="middle" font-family="Arial, sans-serif">
    📊 GitHub Statistics
  </text>
  
  <!-- Stats -->`)

	sb.WriteString(fmt.Sprintf(`
  <text x="20" y="65" fill="#c9d1d9" font-size="14" font-family="Arial, sans-serif">
    🎯 Total Contributions: <tspan fill="#58a6ff" font-weight="bold">%d</tspan>
  </text>
  <text x="20" y="90" fill="#c9d1d9" font-size="14" font-family="Arial, sans-serif">
    💻 Commits: <tspan fill="#58a6ff" font-weight="bold">%d</tspan>
  </text>
  <text x="20" y="115" fill="#c9d1d9" font-size="14" font-family="Arial, sans-serif">
    🔀 Pull Requests: <tspan fill="#58a6ff" font-weight="bold">%d</tspan>
  </text>
  <text x="260" y="65" fill="#c9d1d9" font-size="14" font-family="Arial, sans-serif">
    📦 Repositories: <tspan fill="#58a6ff" font-weight="bold">%d</tspan>
  </text>
  <text x="260" y="90" fill="#c9d1d9" font-size="14" font-family="Arial, sans-serif">
    ⭐ Total Stars: <tspan fill="#58a6ff" font-weight="bold">%d</tspan>
  </text>
  <text x="260" y="115" fill="#c9d1d9" font-size="14" font-family="Arial, sans-serif">
    🍴 Total Forks: <tspan fill="#58a6ff" font-weight="bold">%d</tspan>
  </text>`,
		cc.ContributionCalendar.TotalContributions,
		cc.TotalCommitContributions,
		cc.TotalPullRequestContributions,
		repos.TotalCount,
		totalStars,
		totalForks,
	))

	sb.WriteString(`
  
  <!-- Languages Section -->
  <text x="20" y="150" fill="#8b949e" font-size="12" font-weight="bold" font-family="Arial, sans-serif">
    TOP LANGUAGES
  </text>`)

	yPos := 170
	for i, lang := range languages {
		if i >= 6 {
			break
		}
		barWidth := lang.Percentage * 4.5
		color := lang.Color
		if color == "" {
			color = "#858585"
		}

		sb.WriteString(fmt.Sprintf(`
  <text x="20" y="%d" fill="#c9d1d9" font-size="11" font-family="monospace">%s</text>
  <rect x="130" y="%d" width="%.1f" height="12" fill="%s" rx="2"/>
  <text x="%.1f" y="%d" fill="#c9d1d9" font-size="10" font-family="monospace">%.1f%%</text>`,
			yPos, lang.Name,
			yPos-10, barWidth, color,
			135+barWidth, yPos, lang.Percentage,
		))
		yPos += 20
	}

	sb.WriteString("\n</svg>")

	return sb.String()
}
