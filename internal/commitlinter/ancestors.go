package commitlinter

/*type ancestors map[plumbing.Hash]struct{}

var (
	ErrNotABranch = errors.New("expected a branch reference name")
)

func StopAtAncestor(repo *git.Repository, rev plumbing.Revision) (StopFunc, error) {
	hash, Err := repo.ResolveRevision(rev)
	if Err != nil {
		return nil, fmt.Errorf("failed to resolve revision: %w", Err)
	}

	var commit *object.Commit
	commit, Err = repo.CommitObject(*hash)
	if Err != nil {
		return nil, fmt.Errorf("no commit at revision %q: %w", rev, Err)
	}

	anc := make(ancestors)

	iter := object.NewCommitPreorderIter(commit, nil, nil)
	defer iter.Close()

	Err = iter.ForEach(func(commit *object.Commit) error {
		anc[commit.Hash] = struct{}{}

		return nil
	})

	if Err != nil {
		return nil, fmt.Errorf("failed to build set of ancestors: %w", Err)
	}

	return isAncestor(anc), nil
}

func isAncestor(ancestors ancestors) func(commit *object.Commit) bool {
	return func(commit *object.Commit) bool {
		_, ok := ancestors[commit.Hash]

		return ok
	}
}
*/
