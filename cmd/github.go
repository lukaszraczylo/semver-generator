package cmd

import (
	"github.com/lukaszraczylo/semver-generator/cmd/utils"
)

// These functions are now in the utils package
// They are kept here as stubs for backward compatibility

func updatePackage() bool {
	return utils.UpdatePackage()
}

func checkLatestRelease() (string, bool) {
	return utils.CheckLatestRelease()
}
