/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package cli

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"codeberg.org/somebadcode/commit-tool/cli/bindings"
)

type UnknownCommandError struct {
	Name string
}

func (e *UnknownCommandError) Error() string {
	return fmt.Sprintf("%s: unknown command", e.Name)
}

type AlreadyExistsError struct {
	Name string
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("%s: command already exists", e.Name)
}

type Executor interface {
	Execute(context.Context, *CLI) error
}

type Prioritizer interface {
	Priority() int
}

type task struct {
	name     string
	node     Executor
	priority int
}

type CLI struct {
	bindings *bindings.Bindings
	tasks    []task
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
}

var ErrStop = errors.New("stop execution")

func New(option ...Option) *CLI {
	cli := &CLI{
		bindings: bindings.New(),
		tasks:    make([]task, 0, 4),
		stdin:    os.Stdin,
		stdout:   os.Stdout,
		stderr:   os.Stderr,
	}

	for _, opt := range option {
		opt(cli)
	}

	return cli
}

func (c *CLI) Stdin() io.Reader {
	return c.stdin
}

func (c *CLI) Stdout() io.Writer {
	return c.stdout
}

func (c *CLI) Stderr() io.Writer {
	return c.stderr
}

func (c *CLI) Register(name string, binder bindings.Binder, options ...bindings.Option) error {
	if ok := c.bindings.Has(name); ok {
		return &AlreadyExistsError{
			Name: name,
		}
	}

	c.bindings.BindNode(name, binder, options...)

	return nil
}

func (c *CLI) Parse(args []string) error {
	c.tasks = c.tasks[:0]

	for len(args) > 0 {
		prev := len(args)

		rem, err := c.bindings.Parse(args, func(keyword string, binder bindings.Binder) {
			if executor, ok := binder.(Executor); ok {
				c.appendTask(keyword, executor)
			}
		})

		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}

		if len(rem) == prev {
			return &UnknownCommandError{
				Name: args[0],
			}
		}

		args = rem
	}

	return nil
}

func (c *CLI) Execute(ctx context.Context) error {
	slices.SortStableFunc(c.tasks, func(a, b task) int {
		return cmp.Compare(b.priority, a.priority)
	})

	for _, t := range c.tasks {
		err := t.node.Execute(ctx, c)
		if err != nil {
			if errors.Is(err, ErrStop) {
				return nil
			}

			return fmt.Errorf("%s: %w", t.name, err)
		}
	}

	return nil
}

func (c *CLI) appendTask(name string, node Executor) {
	var prio int
	if p, ok := node.(Prioritizer); ok {
		prio = p.Priority()
	}

	c.tasks = append(c.tasks, task{
		name:     name,
		node:     node,
		priority: prio,
	})
}
