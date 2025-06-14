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
