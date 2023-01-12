package main

import (
	"context"
	"log"
	"os"
	"strings"
	"text/template"

	graphql "github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
)

type readmeData struct {
	BlogPostTitles []string
	RecentStars    []struct {
		Description   string `json:"description"`
		NameWithOwner string `json:"nameWithOwner"`
	}
	RecentActivity []struct {
		Description string `json:"description"`
		Name        string `json:"name"`
	}
}

var readme = strings.ReplaceAll(`### Hi there 👋

---

<a href="https://github.com/alexanderjophus"><img src="https://img.shields.io/github/followers/alexanderjophus.svg?label=GitHub&style=social" alt="GitHub"></a>
<a href="https://twitter.com/AlexanderJophus"><img src="https://img.shields.io/twitter/follow/AlexanderJophus?label=Twitter&style=social" alt="Twitter"></a>
<a href="https://twitch.tv/dejophus"><img src="https://img.shields.io/twitch/status/dejophus?style=social" alt="Twitch"></a>

§§§yaml
name: Alexander Jophus
located_in: Bristol, UK
job: Senior Software Engineer (Go)

fields_of_interests: ["Developer Experience", "DevOps", "Making Microservices Go Zoom"]
familiarity: ["Go", "Python", "Kubernetes"]
currently_learning: ["Game Design", "Elixir"]
hobbies: ["Gaming", "Music"]
§§§

<a href="https://github.com/alexanderjophus/alexanderjophus">
  <img align="center" src="https://github-readme-stats-git-masterrstaa-rickstaa.vercel.app/api/top-langs?username=alexanderjophus&hide=java,html,tex&langs_count=3&theme=vision-friendly-dark" />
</a>
<a href="https://github.com/alexanderjophus/alexanderjophus">
  <img align="center" src="https://github-readme-stats-git-masterrstaa-rickstaa.vercel.app/api?username=alexanderjophus&show_icons=true&line_height=27&count_private=true&theme=vision-friendly-dark" alt="Alexanders GitHub Stats" />
</a>

## Recent Stars
| Repository | Description |
|---|---|{{ range .RecentStars }}
| [{{ .NameWithOwner }}](https://www.github.com/{{ .NameWithOwner }}) | {{ .Description }} |{{ end }}

## Actively Working On (publicly)
| Repository | Description |
|---|---|{{ range .RecentActivity }}
| [{{ .Name }}](https://www.github.com/alexanderjophus/{{ .Name }}) | {{ .Description }} |{{ end }}`, "§", "`")

var query struct {
	User struct {
		StarredRepositories struct {
			Nodes []struct {
				Description   string `json:"description"`
				NameWithOwner string `json:"nameWithOwner"`
			} `json:"nodes"`
		} `graphql:"starredRepositories(last:5)"`
		Repositories struct {
			Nodes []struct {
				Description string `json:"description"`
				Name        string `json:"name"`
			} `json:"nodes"`
		} `graphql:"repositories(privacy:PUBLIC, last:5, orderBy:{field:PUSHED_AT, direction:ASC})"`
	} `graphql:"user(login: \"alexanderjophus\")"`
}

func main() {
	tmpl := template.Must(template.New("README").Parse(readme))

	out, err := os.OpenFile("README.md", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_GRAPHQL_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := graphql.NewClient("https://api.github.com/graphql", httpClient)
	err = client.Query(context.Background(), &query, nil)
	if err != nil {
		log.Fatal(err)
	}

	tmpl.Execute(out, readmeData{
		BlogPostTitles: []string{},
		RecentStars:    query.User.StarredRepositories.Nodes,
		RecentActivity: query.User.Repositories.Nodes,
	})
}
