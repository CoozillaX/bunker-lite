package enhance

import "bunker-core/protocol/mpay/android"

// UseGamePackage selects a package by Bedrock version. An empty version selects
// the current official published package.
func UseGamePackage(user *android.AndroidMpayUser, bedrockVersion string) error {
	if bedrockVersion == "" {
		return user.UsePublishedPackage()
	}
	return user.UpdateGameInfoByBedrockVersion(bedrockVersion)
}
