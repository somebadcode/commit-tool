package commitlinter

/*import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func isRevisionEqual(repo *git.Repository, x, y plumbing.Revision) (bool, error) {
	hx, err := repo.ResolveRevision(x)
	if err != nil {
		return false, fmt.Errorf("failed to resolve revision %q: %w", x, err)
	}

	var hy *plumbing.Hash
	hy, err = repo.ResolveRevision(y)
	if err != nil {
		return false, fmt.Errorf("failed to resolve revision %q: %w", x, err)
	}

	return *hx == *hy, nil
}
*/
