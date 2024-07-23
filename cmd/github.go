package cmd

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/lukaszraczylo/ask"
	graphql "github.com/lukaszraczylo/go-simple-graphql"
	libpack_logger "github.com/lukaszraczylo/graphql-monitoring-proxy/logging"
	"github.com/melbahja/got"
)

func updatePackage() bool {
	ghToken, ghTokenSet := os.LookupEnv("GITHUB_TOKEN")
	if ghTokenSet {
		binaryName := fmt.Sprintf("semver-gen-%s-%s", runtime.GOOS, runtime.GOARCH)
		logger.Info(&libpack_logger.LogMessage{Message: "Checking for updates", Pairs: map[string]interface{}{"binaryName": binaryName}})
		gql := graphql.NewConnection()

		gql.SetEndpoint("https://api.github.com/graphql")
		gql.SetOutput("mapstring")

		headers := map[string]interface{}{
			"Authorization": fmt.Sprintf("Bearer %s", ghToken),
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
			logger.Error(&libpack_logger.LogMessage{Message: "Unable to query GitHub API", Pairs: map[string]interface{}{"error": err.Error()}})
			return false
		}

		output, ok := ask.For(result, "repository.latestRelease.releaseAssets.edges[0].node.downloadUrl").String("")
		if !ok {
			logger.Error(&libpack_logger.LogMessage{Message: "Unable to obtain download url for the binary", Pairs: map[string]interface{}{"binary": binaryName, "output": output}})
			return false
		}
		if flag.Lookup("test.v") == nil && os.Getenv("CI") == "" {
			downloadedBinaryPath := fmt.Sprintf("/tmp/%s", binaryName)
			g := got.New()
			err = g.Download(output, downloadedBinaryPath)
			if err != nil {
				logger.Error(&libpack_logger.LogMessage{Message: "Unable to download binary", Pairs: map[string]interface{}{"error": err.Error(), "binaryPath": downloadedBinaryPath}})
				return false
			}
			currentBinary, err := os.Executable()
			if err != nil {
				logger.Error(&libpack_logger.LogMessage{Message: "Unable to obtain current binary path", Pairs: map[string]interface{}{"error": err.Error()}})
				return false
			}
			err = os.Rename(downloadedBinaryPath, currentBinary)
			if err != nil {
				logger.Error(&libpack_logger.LogMessage{Message: "Unable to overwrite current binary", Pairs: map[string]interface{}{"error": err.Error()}})
				return false
			}
			err = os.Chmod(currentBinary, 0777)
			if err != nil {
				logger.Error(&libpack_logger.LogMessage{Message: "Unable to make binary executable", Pairs: map[string]interface{}{"error": err.Error()}})
				return false
			}
		}
	}
	return true
}

func checkLatestRelease() (string, bool) {
	ghToken, ghTokenSet := os.LookupEnv("GITHUB_TOKEN")
	if ghTokenSet {
		gql := graphql.NewConnection()
		gql.SetEndpoint("https://api.github.com/graphql")
		headers := map[string]interface{}{
			"Authorization": fmt.Sprintf("bearer %s", ghToken),
		}
		variables := map[string]interface{}{}
		var query = `query {
			repository(name: "semver-generator", owner: "lukaszraczylo", followRenames: true) {
				releases(last: 2) {
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
			logger.Error(&libpack_logger.LogMessage{Message: "Unable to query GitHub API", Pairs: map[string]interface{}{"error": err.Error()}})
			return "", false
		}
		output, _ := ask.For(result, "repository.releases.nodes[0].tag.name").String("")
		if output == "v1" {
			output, _ = ask.For(result, "repository.releases.nodes[1].tag.name").String("")
		}
		return output, true
	} else {
		return "[no GITHUB_TOKEN set]", false
	}
}
