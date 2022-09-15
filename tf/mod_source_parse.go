// Copyright 2022 Mineiros GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tf

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/mineiros-io/terramate/errors"
)

// Source represents a module source
type Source struct {
	// URL is the Git URL of the source.
	URL string

	// Path is the path of the source URL. It includes the domain of the URL on it.
	// Eg. github.com/mineiros-io/example
	Path string

	// Subdir is the subdir component of the source path, if any, as defined
	// here: https://www.terraform.io/language/modules/sources#modules-in-package-sub-directories
	Subdir string

	// Ref is the specific reference of this source, if any.
	Ref string

	// Raw source
	Raw string
}

const (
	// ErrUnsupportedModSrc indicates that a module source string is unsupported.
	ErrUnsupportedModSrc errors.Kind = "unsupported module source"

	// ErrInvalidModSrc indicates that a module source string is invalid.
	ErrInvalidModSrc errors.Kind = "invalid module source"
)

// ParseSource parses the given modsource string.
// The modsource must be a valid Terraform Git/Github source reference as documented in:
//
// - https://www.terraform.io/language/modules/sources
//
// Source references that are not Git/Github are not supported.
func ParseSource(modsource string) (Source, error) {
	switch {
	case strings.HasPrefix(modsource, "github.com"):
		u, err := url.Parse(modsource)
		if err != nil {
			return Source{}, errors.E(ErrInvalidModSrc, err,
				"%s is not a URL", modsource)
		}
		ref := u.Query().Get("ref")
		subdir := parseURLSubdir(u)
		u.RawQuery = ""
		u.Scheme = "https"
		u.Path = strings.TrimSuffix(u.Path, ".git")

		path := filepath.Join(u.Host, u.Path)
		return Source{
			Raw:    modsource,
			URL:    u.String() + ".git",
			Path:   path,
			Subdir: subdir,
			Ref:    ref,
		}, nil

	case strings.HasPrefix(modsource, "git@"):
		// In a git scp like url it could be any user@host, but here we are supporting
		// the specific options allowed by Terraform on module.source:
		// - https://www.terraform.io/language/modules/sources#github
		// In this case being Github ssh access, which is always git@.

		rawURL := strings.TrimPrefix(modsource, "git@")

		// This is not a valid URL given the nature of scp strings
		// But it is enough for us to parse the query parameters
		// and form a path that makes sense.
		u, err := url.Parse(rawURL)
		if err != nil {
			return Source{}, errors.E(ErrInvalidModSrc, err,
				"invalid URL inside %s", modsource)
		}

		ref := u.Query().Get("ref")
		u.RawQuery = ""
		path, subdir := parseSubdir(u.Opaque)
		u.Opaque = path
		path = strings.TrimSuffix(filepath.Join(u.Scheme, u.Opaque), ".git")

		return Source{
			Raw:    modsource,
			URL:    "git@" + u.String(),
			Path:   path,
			Subdir: subdir,
			Ref:    ref,
		}, nil

	case strings.HasPrefix(modsource, "git::"):
		// Generic git: https://www.terraform.io/language/modules/sources#generic-git-repository
		rawURL := strings.TrimPrefix(modsource, "git::")
		u, err := url.Parse(rawURL)
		if err != nil {
			return Source{}, errors.E(ErrInvalidModSrc, modsource)
		}

		if u.Path == "" {
			return Source{}, errors.E(
				ErrInvalidModSrc,
				"source %q is missing the path component",
				modsource,
			)
		}

		subdir := parseURLSubdir(u)
		// We don't want : on the path. So we replace the possible :
		// that can exist on the host.
		path := filepath.Join(strings.Replace(u.Host, ":", "-", -1), u.Path)
		path = strings.TrimSuffix(path, ".git")

		if err != nil {
			return Source{}, err
		}
		ref := u.Query().Get("ref")
		u.RawQuery = ""
		return Source{
			Raw:    modsource,
			URL:    u.String(),
			Path:   path,
			Subdir: subdir,
			Ref:    ref,
		}, nil

	default:
		return Source{}, errors.E(ErrUnsupportedModSrc)
	}
}

func parseSubdir(s string) (string, string) {
	if !strings.Contains(s, "//") {
		return s, ""
	}

	// From the specs we should have a single // on the path:
	// https://www.terraform.io/language/modules/sources#modules-in-package-sub-directories
	parsed := strings.Split(s, "//")
	if len(parsed[1]) == 0 {
		return parsed[0], ""
	}
	return parsed[0], "/" + parsed[1]
}

func parseURLSubdir(u *url.URL) string {
	path, subdir := parseSubdir(u.Path)
	u.Path = path
	return subdir
}
