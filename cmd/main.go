package main

import (
	"context"
	_ "embed"
	"log"
	"os"
	"text/template"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/pkg/encoding/yaml"
	graphql "github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
)

//go:embed README.tpl
var readme string

//go:embed main_go_gen.cue
var cuefile string

//go:embed about_me.cue
var aboutMe string

type readmeData struct {
	AboutMe     string
	RecentStars []struct {
		Description   string `json:"description"`
		NameWithOwner string `json:"nameWithOwner"`
	}
	RecentActivity []struct {
		Description string `json:"description"`
		Name        string `json:"name"`
	}
}

type AboutMe struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Job      string `json:"job"`

	FieldsOfInterest  []string `json:"fieldsOfInterest"`
	Hobbies           []string `json:"hobbies"`
	Familiarities     []string `json:"familiarities"`
	CurrentlyLearning []string `json:"currentlyLearning"`
}

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
	ctx := cuecontext.New()
	v := ctx.CompileString(cuefile)
	if v.Err() != nil {
		log.Fatal(v.Err())
	}

	aboutMe := ctx.CompileString(aboutMe)
	if aboutMe.Err() != nil {
		log.Fatal(aboutMe.Err())
	}

	v = v.Unify(aboutMe).LookupPath(cue.ParsePath("#AboutMe"))
	if v.Err() != nil {
		log.Fatal(v.Err())
	}

	yaml, err := yaml.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}

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
		AboutMe:        yaml,
		RecentStars:    query.User.StarredRepositories.Nodes,
		RecentActivity: query.User.Repositories.Nodes,
	})
}
