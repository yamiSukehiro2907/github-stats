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
		lang.Percentage = float64(lang.Size) / float64(totalSize) * 100
		languages = append(languages, *lang)
	}

	sort.Slice(languages, func(i, j int) bool {
		return languages[i].Size > languages[j].Size
	})

	if len(languages) > 12 {
		languages = languages[:12]
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

	sb.WriteString(`<svg width="1000" height="320" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="grad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#58a6ff;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#1f6feb;stop-opacity:1" />
    </linearGradient>
  </defs>
  
  <rect width="1000" height="320" fill="#0d1117" rx="10"/>
  
  <text x="500" y="35" fill="url(#grad)" font-size="24" font-weight="bold" text-anchor="middle" font-family="Arial, sans-serif">
    📊 GitHub Statistics
  </text>
  `)

	sb.WriteString(fmt.Sprintf(`
  <text x="50" y="80" fill="#c9d1d9" font-size="15" font-family="Arial, sans-serif">
    🎯 Contributions: <tspan fill="#58a6ff" font-weight="bold">%d</tspan>
  </text>
  <text x="50" y="105" fill="#c9d1d9" font-size="15" font-family="Arial, sans-serif">
    💻 Commits: <tspan fill="#58a6ff" font-weight="bold">%d</tspan>
  </text>
  <text x="50" y="130" fill="#c9d1d9" font-size="15" font-family="Arial, sans-serif">
    🔀 PRs: <tspan fill="#58a6ff" font-weight="bold">%d</tspan>
  </text>
  <text x="530" y="80" fill="#c9d1d9" font-size="15" font-family="Arial, sans-serif">
    📦 Repos: <tspan fill="#58a6ff" font-weight="bold">%d</tspan>
  </text>
  <text x="530" y="105" fill="#c9d1d9" font-size="15" font-family="Arial, sans-serif">
    ⭐ Stars: <tspan fill="#58a6ff" font-weight="bold">%d</tspan>
  </text>
  <text x="530" y="130" fill="#c9d1d9" font-size="15" font-family="Arial, sans-serif">
    🍴 Forks: <tspan fill="#58a6ff" font-weight="bold">%d</tspan>
  </text>`,
		cc.ContributionCalendar.TotalContributions,
		cc.TotalCommitContributions,
		cc.TotalPullRequestContributions,
		repos.TotalCount,
		totalStars,
		totalForks,
	))

	sb.WriteString(`
  
  <line x1="50" y1="160" x2="950" y2="160" stroke="#30363d" stroke-width="1"/>
  
  <text x="50" y="185" fill="#8b949e" font-size="14" font-weight="bold" font-family="Arial, sans-serif">
    TOP LANGUAGES
  </text>`)

	leftColumnX := 50
	rightColumnX := 530
	startY := 205
	rowHeight := 18

	for i, lang := range languages {
		if i >= 12 {
			break
		}

		var xPos int
		var yPos int
		if i < 6 {
			xPos = leftColumnX
			yPos = startY + (i * rowHeight)
		} else {
			xPos = rightColumnX
			yPos = startY + ((i - 6) * rowHeight)
		}

		color := lang.Color
		if color == "" {
			color = "#858585"
		}

		barWidth := lang.Percentage * 1.8
		if barWidth < 5 {
			barWidth = 5
		}

		sb.WriteString(fmt.Sprintf(`
  <text x="%d" y="%d" fill="#c9d1d9" font-size="12" font-family="monospace">%s</text>
  <rect x="%d" y="%d" width="%.1f" height="12" fill="%s" rx="2"/>
  <text x="%d" y="%d" fill="#8b949e" font-size="11" font-family="monospace">%.1f%%</text>`,
			xPos, yPos, truncateName(lang.Name, 12),
			xPos+115, yPos-10, barWidth, color,
			xPos+305, yPos, lang.Percentage,
		))
	}

	sb.WriteString("\n</svg>")

	return sb.String()
}

func truncateName(name string, maxLen int) string {
	if len(name) <= maxLen {
		return name
	}
	return name[:maxLen-1] + "…"
}
