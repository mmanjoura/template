package user

import (
	"fmt"

	"github.com/mmanjoura/template"
)

// Authentication providers.
// Currently we only support GitHub but any OAuth provider could be supported.
const (
	AuthSourceGitHub = "github"
)

// Validate returns an error if any fields are invalid on the Auth object.
// This can be called by the SQLite implementation to do some basic checks.
func (a *Auth) Validate() error {
	if a.UserID == 0 {
		return template.Errorf(template.EINVALID, "User required.")
	} else if a.Source == "" {
		return template.Errorf(template.EINVALID, "Source required.")
	} else if a.SourceID == "" {
		return template.Errorf(template.EINVALID, "Source ID required.")
	} else if a.AccessToken == "" {
		return template.Errorf(template.EINVALID, "Access token required.")
	}
	return nil
}

// AvatarURL returns a URL to the avatar image hosted by the authentication source.
// Returns an empty string if the authentication source is invalid.
func (a *Auth) AvatarURL(size int) string {
	switch a.Source {
	case AuthSourceGitHub:
		return fmt.Sprintf("https://avatars1.githubusercontent.com/u/%s?s=%d", a.SourceID, size)
	default:
		return ""
	}
}
