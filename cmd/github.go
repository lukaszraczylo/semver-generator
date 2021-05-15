package cmd

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	gql "github.com/lukaszraczylo/simple-gql-client"
	"github.com/melbahja/got"
	"github.com/tidwall/gjson"
)

func updatePackage() bool {
	ghToken, ghTokenSet := os.LookupEnv("GITHUB_TOKEN")
	if ghTokenSet {
		binaryName := fmt.Sprintf("semver-gen-%s-%s", runtime.GOOS, runtime.GOARCH)
		fmt.Println("Downloading", binaryName)
		gql.GraphQLUrl = "https://api.github.com/graphql"
		headers := map[string]interface{}{
			"Authorization": fmt.Sprintf("bearer %s", ghToken),
		}
		variables := map[string]interface{}{
			"binaryName": binaryName,
		}
		var query = `query ($binaryName: String) {
			repository(name: "semver-generator", owner: "lukaszraczylo") {
				latestRelease {
					releaseAssets(first: 10, name: $binaryName) {
						edges {
							node {
								name
								downloadUrl
							}
						}
					}
				}
			}
		}`
		result, err := gql.Query(query, variables, headers)
		if err != nil {
			fmt.Println("Query error", err)
			return false
		}
		result = gjson.Get(result, "repository.latestRelease.releaseAssets.edges.0.node.downloadUrl").String()
		if result == "" {
			fmt.Println("Unable to obtain download url for the binary", binaryName)
			return false
		}
		if flag.Lookup("test.v") == nil && os.Getenv("CI") == "" {
			downloadedBinaryPath := fmt.Sprintf("/tmp/%s", binaryName)
			g := got.New()
			err = g.Download(result, downloadedBinaryPath)
			if err != nil {
				fmt.Println("Unable to download binary", err.Error())
				return false
			}
			currentBinary, err := os.Executable()
			if err != nil {
				fmt.Println("Unable to obtain current binary path", err.Error())
				return false
			}
			err = os.Rename(downloadedBinaryPath, currentBinary)
			if err != nil {
				fmt.Println("Unable to overwrite current binary", err.Error())
				return false
			}
			err = os.Chmod(currentBinary, 0777)
			if err != nil {
				fmt.Println("Unable to make binary executable", err.Error())
				return false
			}
		}
	}
	return true
}

func checkLatestRelease() (string, bool) {
	ghToken, ghTokenSet := os.LookupEnv("GITHUB_TOKEN")
	if ghTokenSet {
		gql.GraphQLUrl = "https://api.github.com/graphql"
		headers := map[string]interface{}{
			"Authorization": fmt.Sprintf("bearer %s", ghToken),
		}
		variables := map[string]interface{}{}
		var query = `query {
			repository(name: "semver-generator", owner: "lukaszraczylo") {
				releases(last: 1) {
					nodes {
						tag {
							name
						}
					}
				}
			}
		}`
		result, err := gql.Query(query, variables, headers)
		if err != nil {
			fmt.Println("Query error", err)
			return "", false
		}
		result = gjson.Get(result, "repository.releases.nodes.0.tag.name").String()
		return result, true
	} else {
		return "[no GITHUB_TOKEN set]", false
	}
}
