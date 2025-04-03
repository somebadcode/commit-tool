/*
 * Copyright 2025 Tobias Dahlberg
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package linter

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
)

type LintError struct {
	Err  error
	Hash plumbing.Hash
	Pos  int
}

func (err LintError) Unwrap() error {
	return err.Err
}

func (err LintError) Error() string {
	return fmt.Sprintf("bad commit message at %s on line %d: %s", err.Hash, err.Pos, err.Err)
}

type Error struct {
	errs []error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d violations found", len(e.errs))
}

func (e *Error) Unwrap() []error {
	return e.errs
}
