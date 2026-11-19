/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"testing"

	"codeberg.org/somebadcode/commit-tool/cli/bindings"
)

type DummyNode struct {
	tb          testing.TB
	priority    int
	doubleBind  bool
	failExecute bool
	Line        string
	Lines       []string
	ExtraText   string
}

type DummyNodeOption func(*DummyNode)

func FailExecute() func(*DummyNode) {
	return func(node *DummyNode) {
		node.failExecute = true
	}
}

func WithPriority(priority int) DummyNodeOption {
	return func(node *DummyNode) {
		node.priority = priority
	}
}

func NewDummyNode(options ...DummyNodeOption) *DummyNode {
	node := &DummyNode{}

	for _, option := range options {
		option(node)
	}

	return node
}

func (node *DummyNode) Bind(b *bindings.Bindings) {
	b.BindString("line", &node.Line)
	b.BindStringSlice("lines", &node.Lines)
	b.BindString("extra", &node.ExtraText)

	if node.doubleBind {
		b.BindString("line", &node.Line)
	}
}

func (node *DummyNode) Execute(_ context.Context, c *CLI) error {
	if node.failExecute {
		return errors.New("internal error")
	}

	scanner := bufio.NewScanner(c.Stdin())

	for scanner.Scan() {
		node.Lines = append(node.Lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	for _, line := range node.Lines {
		if _, err := fmt.Fprint(c.Stdout(), line); err != nil {
			return err
		}
	}

	_, err := fmt.Fprint(c.Stderr(), node.ExtraText)
	if err != nil {
		return err
	}

	return nil
}

func (node *DummyNode) Priority() int {
	return node.priority
}
