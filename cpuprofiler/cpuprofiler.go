/*
 * This file is part of commit-tool which is released under EUPL 1.2.
 * See the file LICENSE in the repository root for full license details.
 *
 * SPDX-License-Identifier: EUPL-1.2
 */

package cpuprofiler

import (
	"fmt"
	"os"
	"runtime/pprof"
)

type Profiler struct {
	Filename string
}

type UnsupportedOutputError struct {
	Name string
}

func (e UnsupportedOutputError) Error() string {
	return fmt.Sprintf("unsupported output: %s", e.Name)
}

func (p Profiler) Start() (func(), error) {
	if p.Filename == "" {
		return func() {}, nil
	}

	if p.Filename == "-" {
		return nil, &UnsupportedOutputError{
			Name: p.Filename,
		}
	}

	f, err := os.OpenFile(p.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, fmt.Errorf("cpuprofiler.Start: %w", err)
	}

	err = pprof.StartCPUProfile(f)
	if err != nil {
		return nil, fmt.Errorf("cpuprofiler.Start: %w", err)
	}

	return func() {
		defer func() {
			_ = f.Close()
		}()

		pprof.StopCPUProfile()
	}, nil
}
