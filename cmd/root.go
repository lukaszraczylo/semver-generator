/*
Copyright © 2021 LUKASZ RACZYLO <lukasz$raczylo,com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "semver-gen generate [flags]",
	Short: "An effortless semantic version generator",
	Long: `semver-gen // Lukasz Raczylo, raczylo.com

Effortless semantic version generator with git commit keywords matching, allowing you to focus on the development.
Visit https://github.com/lukaszraczylo/semver-generator for more information, documentation and examples.`,
	Run: func(cmd *cobra.Command, args []string) {
		main()
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func (r *Setup) setupCobra() {
	r.RepositoryName, err = rootCmd.Flags().GetString("repository")
	if err != nil {
		panic(err)
	}
	r.RepositoryBranch, err = rootCmd.Flags().GetString("branch")
	if err != nil {
		panic(err)
	}
	r.LocalConfigFile, err = rootCmd.Flags().GetString("config")
	if err != nil {
		panic(err)
	}
	r.UseLocal = params.varUseLocal
	if err != nil {
		panic(err)
	}
}

type myParams struct {
	varRepoName       string
	varRepoBranch     string
	varLocalCfg       string
	varUseLocal       bool
	varShowVersion    bool
	varDebug          bool
	varUpdate         bool
	varStrict         bool
	varGenerateInTest bool
	varExisting       bool
}

var params myParams

func init() {
	repo = &Setup{}
	cobra.OnInitialize(repo.setupCobra)
	rootCmd.PersistentFlags().StringVarP(&params.varRepoName, "repository", "r", "https://github.com/lukaszraczylo/simple-gql-client", "Remote repository URL.")
	rootCmd.PersistentFlags().StringVarP(&params.varRepoBranch, "branch", "b", "main", "Remote repository URL Branch.")
	rootCmd.PersistentFlags().StringVarP(&params.varLocalCfg, "config", "c", "semver.yaml", "Path to config file")
	rootCmd.PersistentFlags().BoolVarP(&params.varUseLocal, "local", "l", false, "Use local repository")
	rootCmd.PersistentFlags().BoolVarP(&params.varShowVersion, "version", "v", false, "Display version")
	rootCmd.PersistentFlags().BoolVarP(&params.varDebug, "debug", "d", false, "Enable debug mode")
	rootCmd.PersistentFlags().BoolVarP(&params.varUpdate, "update", "u", false, "Update binary with latest")
	rootCmd.PersistentFlags().BoolVarP(&params.varStrict, "strict", "s", false, "Strict matching")
	rootCmd.PersistentFlags().BoolVarP(&params.varExisting, "existing", "e", false, "Respect existing tags")
}
