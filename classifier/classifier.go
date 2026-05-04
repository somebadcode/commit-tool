/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package classifier

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"

	"codeberg.org/somebadcode/commit-tool/commitparser"
)

type Classification int

const (
	None Classification = iota
	Patch
	Minor
	Major
)

var reCommitType = regexp.MustCompile(`^[a-z]+$`)

type Classifier struct {
	typeMapping map[string]Classification
}

type jsonMapping struct {
	None  []string `json:"none,omitempty"`
	Patch []string `json:"patch,omitempty"`
	Minor []string `json:"minor,omitempty"`
	Major []string `json:"major,omitempty"`
}

var (
	ErrDuplicateCommitType    = errors.New("duplicate commit type")
	ErrNonUnique              = errors.New("non-unique item")
	ErrNonCanonicalCommitType = errors.New("non-canonical commit type")
)

// New will parse the optional [io.Reader] r. When r is nil then the default commit type classification is used.
// For the default classification, "chore", "docs", "refactor", "style" and "test" are [None], "feat" is [Minor],
// "fix" and all types others are [Patch]. When providing a custom classification mapping, make sure to include at
// least "feat", "fix" and "chore" ith the same mapping.
func New(r io.Reader) (*Classifier, error) {
	if r != nil {
		c := &Classifier{typeMapping: make(map[string]Classification)}
		err := c.parse(r)
		if err != nil {
			return nil, fmt.Errorf("read classification mapping: %v", err)
		}
	}

	return &Classifier{
		typeMapping: map[string]Classification{
			"chore":    None,
			"docs":     None,
			"refactor": None,
			"style":    None,
			"test":     None,
			"feat":     Minor,
			"fix":      Patch,
		},
	}, nil
}

func (c *Classifier) Classify(msg *commitparser.CommitMessage) Classification {
	if msg.Breaking {
		return Major
	}

	class, ok := c.typeMapping[msg.Type]
	if !ok {
		return Patch
	}

	return class
}

func (c *Classifier) parse(r io.Reader) error {
	var jw jsonMapping

	err := json.NewDecoder(r).Decode(&jw)
	if err != nil {
		return fmt.Errorf("parse commit type mapping: %w", err)
	}

	err = validateUnique(jw.None, jw.Patch, jw.Minor, jw.Major)
	if err != nil {
		return fmt.Errorf("parse commit type mapping: %w", err)
	}

	return nil
}

func validateUnique(ss ...[]string) error {
	controlSet := make(map[string]struct{})

	for _, s := range ss {
		for _, elem := range s {
			if !reCommitType.MatchString(elem) {
				return fmt.Errorf("type %q: %w", elem, ErrDuplicateCommitType)
			}

			if _, exists := controlSet[elem]; exists {
				return fmt.Errorf("item %q: %w", elem, ErrDuplicateCommitType)
			}

			controlSet[elem] = struct{}{}
		}
	}

	return nil
}
