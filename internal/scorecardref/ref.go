// Package scorecardref parses and validates repository references of the form
// platform/owner/repo.
package scorecardref

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/uwu-tools/scorecard-mcp/internal/model"
)

// Supported platforms.
const (
	PlatformGitHub = "github.com"
	PlatformGitLab = "gitlab.com"
)

var (
	supportedPlatforms = map[string]bool{PlatformGitHub: true, PlatformGitLab: true}
	commitSHA          = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
)

// Parse parses a "platform/owner/repo" reference. The platform is optional and
// defaults to github.com; gitlab.com is also supported. A leading scheme and a
// trailing slash are tolerated. Unsupported platforms and malformed references
// return an error.
func Parse(s string) (model.RepoRef, error) {
	orig := s
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.Trim(s, "/")
	if s == "" {
		return model.RepoRef{}, fmt.Errorf("empty repository reference; expected platform/owner/repo")
	}

	parts := strings.Split(s, "/")
	var ref model.RepoRef
	switch len(parts) {
	case 2:
		ref = model.RepoRef{Platform: PlatformGitHub, Org: parts[0], Name: parts[1]}
	case 3:
		ref = model.RepoRef{Platform: strings.ToLower(parts[0]), Org: parts[1], Name: parts[2]}
	default:
		return model.RepoRef{}, fmt.Errorf(
			"invalid repository reference %q; expected platform/owner/repo (platform optional, defaults to %s)",
			orig, PlatformGitHub)
	}

	if ref.Org == "" || ref.Name == "" {
		return model.RepoRef{}, fmt.Errorf(
			"invalid repository reference %q; owner and repository must be non-empty", orig)
	}
	if !supportedPlatforms[ref.Platform] {
		return model.RepoRef{}, fmt.Errorf(
			"unsupported platform %q; supported platforms are %s and %s",
			ref.Platform, PlatformGitHub, PlatformGitLab)
	}
	return ref, nil
}

// ValidateCommit returns an error if commit is non-empty and is not a
// 40-character hexadecimal SHA.
func ValidateCommit(commit string) error {
	if commit == "" {
		return nil
	}
	if !commitSHA.MatchString(commit) {
		return fmt.Errorf("invalid commit %q; expected a 40-character hexadecimal SHA", commit)
	}
	return nil
}
