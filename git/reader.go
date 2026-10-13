package git

import (
	"strings"
	"fmt"
	"errors"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Repository struct {
	repo *gogit.Repository
}

func Open(path string) (*Repository, error) {
	r, err := gogit.PlainOpenWithOptions(path, &gogit.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, err
	}
	return &Repository{repo: r}, nil
}

func (r *Repository) GetCommits(limit int) ([]Commit, error) {
	refMap, headBranch, err := r.collectRefs()
	if err != nil {
		return nil, err
	}

	head, err := r.repo.Head()
	if err != nil {
		return nil, err
	}

	cIter, err := r.repo.Log(&gogit.LogOptions{
		From: head.Hash(),
	})
	if err != nil {
		return nil, err
	}
	defer cIter.Close()

	var commits []Commit
	count := 0
	stopLimit := errors.New("limit reached")

	err = cIter.ForEach(func(c *object.Commit) error {
		if limit > 0 && count >= limit {
			return stopLimit
		}

		hashStr := c.Hash.String()
		shortHash := hashStr
		if len(shortHash) > 7 {
			shortHash = shortHash[:7]
		}

		subject := strings.Split(c.Message, "\n")[0]

		var parentHashes []string
		for _, ph := range c.ParentHashes {
			parentHashes = append(parentHashes, ph.String())
		}

		refs := refMap[hashStr]
		for i := range refs {
			if refs[i].Name == headBranch {
				refs[i].IsHEAD = true
			}
		}

		commits = append(commits, Commit{
			Hash:         hashStr,
			ShortHash:    shortHash,
			ParentHashes: parentHashes,
			AuthorName:   c.Author.Name,
			AuthorEmail:  c.Author.Email,
			When:         c.Author.When,
			Subject:      subject,
			Refs:         refs,
		})

		count++
		return nil
	})

	if err == stopLimit {
		err = nil
	}

	return commits, err
}

func (r *Repository) collectRefs() (map[string][]Ref, string, error) {
	refMap := make(map[string][]Ref)
	var headBranch string

	if head, err := r.repo.Head(); err == nil && head.Name().IsBranch() {
		headBranch = head.Name().Short()
	}

	iter, err := r.repo.References()
	if err != nil {
		return nil, "", err
	}
	defer iter.Close()

	_ = iter.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() != plumbing.HashReference {
			return nil
		}

		name := ref.Name()
		shortName := name.Short()
		hash := ref.Hash().String()

		refItem := Ref{
			Name:     shortName,
			IsRemote: name.IsRemote(),
			IsTag:    name.IsTag(),
		}

		refMap[hash] = append(refMap[hash], refItem)
		return nil
	})

	return refMap, headBranch, nil
}

func (r *Repository) GetStatus() (*StatusResult, error) {
	w, err := r.repo.Worktree()
	if err != nil {
		return nil, err
	}

	status, err := w.Status()
	if err != nil {
		return nil, err
	}

	res := &StatusResult{}

	head, err := r.repo.Head()
	if err == nil {
		if head.Name().IsBranch() {
			res.Branch = head.Name().Short()
		} else {
			res.Branch = "HEAD (detached)"
		}
	} else {
		res.Branch = "HEAD (empty)"
	}

	if res.Branch != "" && !head.Name().IsBranch() == false {
		r.calculateBranchSync(res, head.Hash())
	}

	for path, fileStatus := range status {
		if fileStatus.Staging != gogit.Unmodified && fileStatus.Staging != gogit.Untracked {
			res.Staged = append(res.Staged, FileChange{
				Path:   path,
				Letter: statusCodeToLetter(fileStatus.Staging),
			})
		}
		if fileStatus.Worktree == gogit.Untracked {
			res.Untracked = append(res.Untracked, FileChange{
				Path:   path,
				Letter: "U",
			})
			continue
		}

		if fileStatus.Worktree != gogit.Unmodified {
			res.Unstaged = append(res.Unstaged, FileChange{
				Path:   path,
				Letter: statusCodeToLetter(fileStatus.Worktree),
			})
		}
	}

	return res, nil
}

func statusCodeToLetter(code gogit.StatusCode) string {
	switch code {
	case gogit.Modified:
		return "M"
	case gogit.Added:
		return "A"
	case gogit.Deleted:
		return "D"
	case gogit.Renamed:
		return "R"
	default:
		return string(code)
	}
}

func (r *Repository) calculateBranchSync(res *StatusResult, localHash plumbing.Hash) {
	config, err := r.repo.Config()
	if err != nil {
		return
	}

	branchCfg := config.Branches[res.Branch]
	if branchCfg == nil || branchCfg.Remote == "" {
		return
	}

	remoteRefName := plumbing.NewRemoteReferenceName(branchCfg.Remote, res.Branch)
	remoteRef, err := r.repo.Reference(remoteRefName, true)
	if err != nil {
		return
	}

	res.HasUpstream = true
	res.UpstreamName = branchCfg.Remote + "/" + res.Branch

	upstreamHash := remoteRef.Hash()
	if localHash == upstreamHash {
		return
	}

	upstreamMap := make(map[string]bool)
	uIter, err := r.repo.Log(&gogit.LogOptions{From: upstreamHash})
	if err == nil {
		count := 0
		_ = uIter.ForEach(func(c *object.Commit) error {
			if count >= 50 {
				return fmt.Errorf("limit")
			}
			upstreamMap[c.Hash.String()] = true
			count++
			return nil
		})
		uIter.Close()
	}

	localMap := make(map[string]bool)
	lIter, err := r.repo.Log(&gogit.LogOptions{From: localHash})
	if err == nil {
		count := 0
		_ = lIter.ForEach(func(c *object.Commit) error {
			if count >= 50 {
				return fmt.Errorf("limit")
			}
			localMap[c.Hash.String()] = true
			if !upstreamMap[c.Hash.String()] {
				res.Ahead++
			}
			count++
			return nil
		})
		lIter.Close()
	}

	for hash := range upstreamMap {
		if !localMap[hash] {
			res.Behind++
		}
	}
}