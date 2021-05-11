package cmd

import (
	"fmt"
	"os"

	gql "github.com/lukaszraczylo/simple-gql-client"
)

func checkLatestRelease() (string, bool) {
	ghToken, ghTokenSet := os.LookupEnv("GHCR_TOKEN")
	if ghTokenSet {
		gql.GraphQLUrl = "https://api.github.com/graphql"
		headers := map[string]interface{}{
			"Authorization": fmt.Sprintf("bearer %s", ghToken),
		}
		variables := map[string]interface{}{}
		var query = `query { viewer { login }}`
		result, err := gql.Query(query, variables, headers)
		if err != nil {
			fmt.Println("Query error", err)
			return "", false
		}
		fmt.Println(result)
		return result, true
	} else {
		return "[no GITHUB_TOKEN set]", false
	}
}
