package utils

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// CommitDetails represents a git commit
type CommitDetails struct {
	Timestamp time.Time
	Hash      string
	Author    string
	Message   string
}

// TagDetails represents a git tag
type TagDetails struct {
	Name string
	Hash string
}

// GitRepository represents a git repository
type GitRepository struct {
	Handler     *git.Repository
	Name        string
	Branch      string
	LocalPath   string
	UseLocal    bool
	Commits     []CommitDetails
	Tags        []TagDetails
	StartCommit string
}

// PrepareRepository prepares the git repository for use
func PrepareRepository(repo *GitRepository) error {
	var err error

	if !repo.UseLocal {
		u, err := url.Parse(repo.Name)
		if err != nil {
			Error("Unable to parse repository URL", map[string]interface{}{
				"error": err.Error(),
				"url":   repo.Name,
			})
			return err
		}

		repo.LocalPath = fmt.Sprintf("/tmp/semver/%s/%s", u.Path, repo.Branch)
		os.RemoveAll(repo.LocalPath)

		repo.Handler, err = git.PlainClone(repo.LocalPath, false, &git.CloneOptions{
			URL:           repo.Name,
			ReferenceName: plumbing.NewBranchReferenceName(repo.Branch),
			SingleBranch:  true,
			Auth: &http.BasicAuth{
				Username: os.Getenv("GITHUB_USERNAME"),
				Password: os.Getenv("GITHUB_TOKEN"),
			},
			Tags: git.AllTags,
		})

		if err != nil {
			Error("Unable to clone repository", map[string]interface{}{
				"error": err.Error(),
				"url":   repo.Name,
			})
			return err
		}
	} else {
		repo.LocalPath = "./"
		repo.Handler, err = git.PlainOpen(repo.LocalPath)
		if err != nil {
			Error("Unable to open local repository", map[string]interface{}{
				"error": err.Error(),
				"path":  repo.LocalPath,
			})
			return err
		}
	}

	os.Chdir(repo.LocalPath)
	return nil
}

// ListCommits lists all commits in the repository
func ListCommits(repo *GitRepository) ([]CommitDetails, error) {
	var ref *plumbing.Reference
	var err error

	// Check if Handler is nil to avoid panic
	if repo.Handler == nil {
		Debug("Repository handler is nil, skipping commit listing", nil)
		return repo.Commits, nil
	}

	ref, err = repo.Handler.Head()
	if err != nil {
		return []CommitDetails{}, err
	}

	commitsList, err := repo.Handler.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return []CommitDetails{}, err
	}

	var tmpResults []CommitDetails
	commitsList.ForEach(func(c *object.Commit) error {
		tmpResults = append(tmpResults, CommitDetails{
			Hash:      c.Hash.String(),
			Author:    c.Author.String(),
			Message:   c.Message,
			Timestamp: c.Author.When,
		})
		sort.Slice(tmpResults, func(i, j int) bool {
			return tmpResults[i].Timestamp.Unix() < tmpResults[j].Timestamp.Unix()
		})
		return nil
	})

	Debug("Listing commits", map[string]interface{}{"commits": tmpResults})

	// Filter commits starting from the specified commit if provided
	if repo.StartCommit != "" {
		for commitId, cmt := range tmpResults {
			if cmt.Hash == repo.StartCommit {
				Debug("Found commit match", map[string]interface{}{
					"commit": cmt.Hash,
					"index":  commitId,
				})
				repo.Commits = tmpResults[commitId:]
				break
			}
		}
	} else {
		repo.Commits = tmpResults
	}

	Debug("Commits after filtering", map[string]interface{}{"commits": repo.Commits})
	return repo.Commits, err
}

// ListExistingTags lists all tags in the repository
func ListExistingTags(repo *GitRepository) {
	Debug("Listing existing tags", nil)

	// Check if Handler is nil to avoid panic
	if repo.Handler == nil {
		Debug("Repository handler is nil, skipping tag listing", nil)
		return
	}

	refs, err := repo.Handler.Tags()
	if err != nil {
		Error("Unable to list tags", map[string]interface{}{"error": err.Error()})
		return
	}

	if err := refs.ForEach(func(ref *plumbing.Reference) error {
		repo.Tags = append(repo.Tags, TagDetails{
			Name: ref.Name().Short(),
			Hash: ref.Hash().String(),
		})

		Debug("Found tag", map[string]interface{}{
			"tag":  ref.Name().Short(),
			"hash": ref.Hash().String(),
		})

		return nil
	}); err != nil {
		Error("Error iterating tags", map[string]interface{}{"error": err.Error()})
	}
}
