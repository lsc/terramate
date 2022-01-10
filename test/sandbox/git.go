// Copyright 2021 Mineiros GmbH
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

package sandbox

import (
	"testing"

	"github.com/madlambda/spells/assert"
	"github.com/mineiros-io/terramate/git"
	"github.com/mineiros-io/terramate/test"
)

const baseRemoteName = "origin"

// Git is a git wrapper that makes testing easy by handling
// errors automatically, failing the caller test.
type Git struct {
	t       *testing.T
	g       *git.Git
	basedir string
}

func NewGit(t *testing.T, basedir string) Git {
	return Git{
		t:       t,
		g:       test.NewGitWrapper(t, basedir, []string{}),
		basedir: basedir,
	}
}

// Init will initialize the git repo with a default local remote.
// After calling Init(), the methods Push() and Pull() pushes and pulls changes
// from/to the configured default remote.
func (git Git) Init() {
	t := git.t
	t.Helper()

	git.InitBasic()

	// the main branch only exists after first commit.
	path := test.WriteFile(t, git.basedir, "README.md", "# generated by terramate")
	git.Add(path)
	git.Commit("first commit")

	git.SetupRemote(baseRemoteName, "main")
}

// SetupRemote will do basic setup of a remote with the given branch.
func (git Git) SetupRemote(remote, branch string) {
	t := git.t
	t.Helper()

	baredir := t.TempDir()
	baregit := test.NewGitWrapper(t, baredir, []string{})

	assert.NoError(t, baregit.Init(baredir, true), "Git.Init(%v, true)", baredir)

	git.RemoteAdd(remote, baredir)
	git.PushOn(remote, branch)
}

// InitBasic will do basic git initialization of a repo,
// not providing a remote configuration.
func (git Git) InitBasic() {
	t := git.t
	t.Helper()

	if err := git.g.Init(git.basedir, false); err != nil {
		t.Fatalf("Git.Init(%v) = %v", git.basedir, err)
	}
}

// RevParse parses the reference name and returns the reference hash.
func (git Git) RevParse(ref string) string {
	git.t.Helper()

	val, err := git.g.RevParse(ref)
	if err != nil {
		git.t.Fatalf("Git.RevParse(%v) = %v", ref, err)
	}

	return val
}

// RemoteAdd adds a new remote on the repo
func (git Git) RemoteAdd(name, url string) {
	err := git.g.RemoteAdd(name, url)
	assert.NoError(git.t, err, "Git.RemoteAdd(%v, %v)", name, url)
}

// Add will add files to the commit list
func (git Git) Add(files ...string) {
	git.t.Helper()

	if err := git.g.Add(files...); err != nil {
		git.t.Fatalf("Git.Add(%v) = %v", files, err)
	}
}

// CurrentBranch returns the short branch name that HEAD points to.
func (git *Git) CurrentBranch() string {
	git.t.Helper()

	branch, err := git.g.CurrentBranch()
	if err != nil {
		git.t.Fatalf("Git.CurrentBranch() = %v", err)
	}
	return branch
}

// Commit will commit previously added files
func (git Git) Commit(msg string, args ...string) {
	git.t.Helper()

	if err := git.g.Commit(msg, args...); err != nil {
		git.t.Fatalf("Git.Commit(%q, %v) = %v", msg, args, err)
	}
}

// Push pushes changes from branch onto default remote
func (git Git) Push(branch string) {
	git.t.Helper()
	git.PushOn(baseRemoteName, branch)
}

// PushOn pushes changes from branch onto the given remote
func (git Git) PushOn(remote, branch string) {
	git.t.Helper()

	if err := git.g.Push(remote, branch); err != nil {
		git.t.Fatalf("Git.Push(%v, %v) = %v", baseRemoteName, branch, err)
	}
}

// Pull pulls changes from default remote into branch
func (git Git) Pull(branch string) {
	git.t.Helper()

	if err := git.g.Pull(baseRemoteName, branch); err != nil {
		git.t.Fatalf("Git.Pull(%v, %v) = %v", baseRemoteName, branch, err)
	}
}

// CommitAll will add all changed files and commit all of them
func (git Git) CommitAll(msg string) {
	git.t.Helper()

	git.Add(".")
	git.Commit(msg)
}

// Checkout will checkout a pre-existing revision
func (git Git) Checkout(rev string) {
	git.t.Helper()
	git.checkout(rev, false)
}

// CheckoutNew will checkout a new revision (creating it on the process)
func (git Git) CheckoutNew(rev string) {
	git.t.Helper()
	git.checkout(rev, true)
}

func (git Git) checkout(rev string, create bool) {
	git.t.Helper()

	if err := git.g.Checkout(rev, create); err != nil {
		git.t.Fatalf("Git.Checkout(%s, %v) = %v", rev, create, err)
	}
}

func (git Git) Merge(branch string) {
	git.t.Helper()

	if err := git.g.Merge(branch); err != nil {
		git.t.Fatalf("Git.Merge(%s) = %v", branch, err)
	}
}

// BaseDir the repository base dir
func (git Git) BaseDir() string {
	return git.basedir
}
