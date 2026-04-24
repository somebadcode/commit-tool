/*
 * SPDX-License-Identifier: EUPL-1.2
 * SPDX-License-Name: European Union Public License 1.2
 *
 * See the NOTICE and LICENSE files in the repository root for attribution and full license details.
 */

package semver

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"unsafe"
)

type Version struct {
	Major, Minor, Patch uint32
	HasVPrefix          bool
	Prerelease          []Segment
	BuildMetadata       []Segment
}

const (
	mask0Through63   uint64 = 0x03ff600000000000 // Character 0-63: '-' and '0-9'
	mask64Through127 uint64 = 0x07fffffe07fffffe // Character 64-127: 'A-Z' and 'a-z'
)

const (
	semverMinLength = 5
)

var (
	ErrDisallowedLeadingZero = errors.New("disallowed leading zero")
	ErrBadVersionFormat      = errors.New("bad version format")
	ErrVersionOutOfRange     = errors.New("version is out of range")
)

// New makes it quick and easy to create a new Version.
func New(major, minor, patch uint32) Version {
	return Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
}

// Parse will parse a string and returns Version.
func Parse(s string) (Version, error) {
	var v Version

	err := v.UnmarshalText([]byte(s))
	if err != nil {
		return v, err
	}

	return v, nil
}

// MustParse is the same as Parse, but it will panic if there's an error. Useful when you're hardcoding the argument.
func MustParse(s string) Version {
	v, err := Parse(s)
	if err != nil {
		panic(err)
	}

	return v
}

// Clone returns an exact clone of Version.
func (v *Version) Clone() Version {
	return Version{
		Prerelease:    slices.Clone(v.Prerelease),
		BuildMetadata: slices.Clone(v.BuildMetadata),
		Major:         v.Major,
		Minor:         v.Minor,
		Patch:         v.Patch,
		HasVPrefix:    v.HasVPrefix,
	}
}

// IsZero returns true if Major, Minor, Patch, Prerelease and BuildMetadata is empty.
func (v *Version) IsZero() bool {
	return v.Major == 0 && v.Minor == 0 && v.Patch == 0 && v.Prerelease == nil && v.BuildMetadata == nil
}

// IncreaseMajor will increase Major, set Minor and Patch to zero and clear Prerelease and BuildMetadata.
func (v *Version) IncreaseMajor() {
	v.Major++
	v.Minor = 0
	v.Patch = 0

	v.Finalize()
}

// IncreaseMinor will increase Minor, set Patch to zero and clear Prerelease and BuildMetadata.
func (v *Version) IncreaseMinor() {
	v.Minor++
	v.Patch = 0

	v.Finalize()
}

// IncreasePatch will increase Patch, and clear Prerelease and BuildMetadata.
func (v *Version) IncreasePatch() {
	v.Patch++

	v.Finalize()
}

// Finalize will clear both Prerelease and BuildMetadata.
func (v *Version) Finalize() {
	v.ClearPrerelease()
	v.ClearBuildMetadata()
}

// ClearPrerelease clears Prerelease
func (v *Version) ClearPrerelease() {
	v.Prerelease = nil
}

// ClearBuildMetadata clears BuildMetadata
func (v *Version) ClearBuildMetadata() {
	v.BuildMetadata = nil
}

// SetPrerelease will set the prerelease information and overwrite the previous prerelease information. The function
// takes one or more arguments where each argument is a Prerelease identifier. To set it to `alpha.1`, pass "alpha" and
// "1" as two separate arguments.
func (v *Version) SetPrerelease(segments ...string) error {
	v.ClearPrerelease()

	for _, segment := range segments {
		err := parseSegment(unsafe.Slice(unsafe.StringData(segment), len(segment)), &v.Prerelease, false)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetBuildMetadata will set the build metadata and overwrite the previous build metadata. The function
// takes one or more arguments where each argument is a build identifier. To set it to `alpha.1`, pass "alpha" and
// "1" as two separate arguments.
func (v *Version) SetBuildMetadata(segments ...string) error {
	v.ClearBuildMetadata()

	for _, segment := range segments {
		err := parseSegment(unsafe.Slice(unsafe.StringData(segment), len(segment)), &v.BuildMetadata, true)
		if err != nil {
			return err

		}
	}

	return nil
}

// String returns a string representation of Version.
func (v *Version) String() string {
	var sb strings.Builder

	if v.HasVPrefix {
		sb.WriteByte('v')
	}

	_, _ = fmt.Fprintf(&sb, "%d.%d.%d", v.Major, v.Minor, v.Patch)

	if v.Prerelease != nil {
		sb.WriteByte('-')
		sb.WriteString(JoinSegments(v.Prerelease, "."))
	}

	if v.BuildMetadata != nil {
		sb.WriteByte('+')
		sb.WriteString(JoinSegments(v.BuildMetadata, "."))
	}

	return sb.String()
}

func (v *Version) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *Version) UnmarshalText(b []byte) error {
	if len(b) < semverMinLength {
		return fmt.Errorf("version string too short: %w", ErrBadVersionFormat)
	}

	var cursor int

	if b[0] == 'v' {
		v.HasVPrefix = true
		cursor++
	}

	var err error

	if v.Major, cursor, err = parseNumericIdentifier(b, cursor); err != nil || cursor >= len(b) || b[cursor] != '.' {
		return fmt.Errorf("major: %w", err)
	}

	// Skip delimiting period.
	cursor++

	if v.Minor, cursor, err = parseNumericIdentifier(b, cursor); err != nil || cursor >= len(b) || b[cursor] != '.' {
		return fmt.Errorf("minor: %w", err)
	}

	// Skip delimiting period.
	cursor++

	if v.Patch, cursor, err = parseNumericIdentifier(b, cursor); err != nil {
		return fmt.Errorf("patch: %w", err)
	}

	// No more information.
	if len(b[cursor:]) == 0 {
		return nil
	}

	if !(b[cursor] == '-' || b[cursor] == '+') {
		return fmt.Errorf("expected '+' or '-', got %c: %w", b[cursor], ErrDisallowedLeadingZero)
	}

	if b[cursor] == '-' {
		cursor++

		end := bytes.IndexByte(b[cursor:], '+')
		if end == -1 {
			end = len(b)
		} else {
			end = cursor + end
		}

		err = parseSegment(b[cursor:end], &v.Prerelease, false)
		if err != nil {
			return fmt.Errorf("prerelease: %w", err)
		}

		cursor = end
	}

	// No more information.
	if len(b[cursor:]) == 0 {
		return nil
	}

	if b[cursor] == '+' {
		cursor++
		err = parseSegment(b[cursor:], &v.BuildMetadata, true)
		if err != nil {
			return fmt.Errorf("metadata: %w", err)
		}
	}

	return nil
}

func parseNumericIdentifier(b []byte, start int) (uint32, int, error) {
	pos := start

	for pos < len(b) && b[pos] >= '0' && b[pos] <= '9' {
		pos++
	}

	if start == pos {
		if b[pos] < '0' || b[pos] > '9' {
			return 0, 0, fmt.Errorf("bad numeric identifier: %w", ErrBadVersionFormat)
		}

		return 0, 0, fmt.Errorf("no numeric identifier: %w", ErrBadVersionFormat)
	}

	if len(b[start:pos]) > 1 && b[start:pos][0] == '0' {
		return 0, 0, fmt.Errorf("bad numeric identifier: %w", ErrDisallowedLeadingZero)
	}

	v, err := strconv.ParseUint(string(b[start:pos]), 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("numeric identifier %q: %w", string(b[start:pos]), err)
	}

	if v > math.MaxUint32 {
		return 0, 0, fmt.Errorf("numeric identifier %d :%w", v, ErrVersionOutOfRange)
	}

	return uint32(v), pos, nil
}

func parseSegment(b []byte, target *[]Segment, allowLeadingZeros bool) error {
	start := 0
	for pos := 0; pos <= len(b); pos++ {
		if pos < len(b) && b[pos] != '.' {
			continue
		}

		data := b[start:pos]

		if len(data) == 0 {
			return fmt.Errorf("no data: %w", ErrBadVersionFormat)
		}

		isNumeric, err := validateSegment(data, allowLeadingZeros)
		if err != nil && !errors.Is(err, ErrDisallowedLeadingZero) {
			return err
		}

		text := string(data)

		var n uint64

		if isNumeric {
			n, _ = strconv.ParseUint(text, 10, 64)
		}

		*target = append(*target, Segment{
			text:      text,
			numeric:   n,
			isNumeric: isNumeric,
		})
		start = pos + 1
	}

	return nil
}

func isValidByte(c byte) bool {
	if c < 64 {
		return (mask0Through63 & (1 << c)) != 0
	}

	if c < 128 {
		return (mask64Through127 & (1<<c - 64)) != 0
	}
	return false
}

// validateSegment validates the data segment and returns whether it's numeric is not.
func validateSegment(b []byte, allowLeadingZeros bool) (bool, error) {
	isNumeric := true

	for _, c := range b {
		if !isValidByte(c) {
			return false, fmt.Errorf("invalid character %#U in data segment: %w", c, ErrBadVersionFormat)
		}

		if c < '0' || c > '9' {
			isNumeric = false
		}
	}

	if !allowLeadingZeros && isNumeric && len(b) > 1 && b[0] == '0' {
		return false, fmt.Errorf("invalid numeric segment: %w", ErrDisallowedLeadingZero)
	}

	return isNumeric, nil
}
