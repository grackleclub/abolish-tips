package main

import "runtime/debug"

// dev is the version reported when no VCS build info is available.
const dev = "dev"

// buildVersion returns the module version, or the VCS revision (with a
// "-dirty" suffix for uncommitted builds) when run from source, falling
// back to dev when no build info is available. The git tag is not in
// build info, so the release tag is injected separately via ldflags.
func buildVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return dev
	}
	if v := bi.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	var rev, modified string
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			modified = s.Value
		}
	}
	if rev == "" {
		return dev
	}
	if len(rev) > 12 {
		rev = rev[:12]
	}
	if modified == "true" {
		rev += "-dirty"
	}
	return rev
}
