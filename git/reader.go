package git

import (
	"strings"

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

	err = cIter.ForEach(func(c *object.Commit) error {
		if limit > 0 && count >= limit {
			return nil
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