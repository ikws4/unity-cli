package cmd

import (
	"strconv"
	"strings"
)

// DefaultVersion is the source-build version. Increment it with every feature.
const DefaultVersion = "0.8.0"

type semanticVersion struct {
	major      int
	minor      int
	patch      int
	prerelease string
}

func isNewerVersion(candidate, current string) bool {
	if strings.TrimSpace(candidate) == "" {
		return false
	}
	candidateVersion, candidateOK := parseSemanticVersion(candidate)
	currentVersion, currentOK := parseSemanticVersion(current)
	if !candidateOK || !currentOK {
		return normalizeVersion(candidate) != normalizeVersion(current)
	}
	return compareSemanticVersions(candidateVersion, currentVersion) > 0
}

func parseSemanticVersion(value string) (semanticVersion, bool) {
	value = normalizeVersion(value)
	if buildIndex := strings.IndexByte(value, '+'); buildIndex >= 0 {
		value = value[:buildIndex]
	}

	prerelease := ""
	if prereleaseIndex := strings.IndexByte(value, '-'); prereleaseIndex >= 0 {
		prerelease = value[prereleaseIndex+1:]
		value = value[:prereleaseIndex]
		if prerelease == "" {
			return semanticVersion{}, false
		}
	}

	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return semanticVersion{}, false
	}
	values := [3]int{}
	for i, part := range parts {
		if part == "" || (len(part) > 1 && part[0] == '0') {
			return semanticVersion{}, false
		}
		parsed, err := strconv.Atoi(part)
		if err != nil || parsed < 0 {
			return semanticVersion{}, false
		}
		values[i] = parsed
	}

	return semanticVersion{
		major:      values[0],
		minor:      values[1],
		patch:      values[2],
		prerelease: prerelease,
	}, true
}

func compareSemanticVersions(left, right semanticVersion) int {
	leftCore := [3]int{left.major, left.minor, left.patch}
	rightCore := [3]int{right.major, right.minor, right.patch}
	for i := range leftCore {
		if leftCore[i] < rightCore[i] {
			return -1
		}
		if leftCore[i] > rightCore[i] {
			return 1
		}
	}

	if left.prerelease == right.prerelease {
		return 0
	}
	if left.prerelease == "" {
		return 1
	}
	if right.prerelease == "" {
		return -1
	}
	return comparePrerelease(left.prerelease, right.prerelease)
}

func comparePrerelease(left, right string) int {
	leftParts := strings.Split(left, ".")
	rightParts := strings.Split(right, ".")
	limit := len(leftParts)
	if len(rightParts) < limit {
		limit = len(rightParts)
	}

	for i := 0; i < limit; i++ {
		leftNumber, leftNumeric := numericIdentifier(leftParts[i])
		rightNumber, rightNumeric := numericIdentifier(rightParts[i])
		switch {
		case leftNumeric && rightNumeric:
			if leftNumber < rightNumber {
				return -1
			}
			if leftNumber > rightNumber {
				return 1
			}
		case leftNumeric:
			return -1
		case rightNumeric:
			return 1
		default:
			if leftParts[i] < rightParts[i] {
				return -1
			}
			if leftParts[i] > rightParts[i] {
				return 1
			}
		}
	}

	if len(leftParts) < len(rightParts) {
		return -1
	}
	if len(leftParts) > len(rightParts) {
		return 1
	}
	return 0
}

func numericIdentifier(value string) (int, bool) {
	if value == "" || (len(value) > 1 && value[0] == '0') {
		return 0, false
	}
	parsed, err := strconv.Atoi(value)
	return parsed, err == nil
}
