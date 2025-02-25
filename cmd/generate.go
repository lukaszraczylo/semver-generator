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

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate [flags]",
	Short: "Generates semantic version",
	Long: `Semantic version generation using your configuration file and fuzzy matching of git commit messages.
	Please refer to documentation on https://github.com/lukaszraczylo/semver-generator for more information.`,
	Run: func(cmd *cobra.Command, args []string) {
		repo.Generate = true
		repo.setupCobra()
		main()
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
