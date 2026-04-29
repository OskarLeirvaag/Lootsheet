package proto

import (
	"fmt"
	"strconv"
	"strings"
)

// ProtocolVersion is the wire protocol version exchanged during AUTH.
// Increment this when the protocol changes in a backwards-incompatible way
// (new required fields, changed semantics, removed methods, etc.).
// Additive changes (new optional fields, new methods) do not require a bump.
const ProtocolVersion uint32 = 1

// AppVersion is the current application version, exchanged during AUTH so
// each side can check compatibility against its minimum requirement.
const AppVersion = "0.6.2"

// MinServerVersion is the oldest server version this client can work with.
// Bump this when the client starts relying on a server feature that older
// servers don't have.
const MinServerVersion = "0.6.2"

// MinClientVersion is the oldest client version this server will accept.
// Bump this when the server changes behaviour that older clients can't handle.
const MinClientVersion = "0.1.0"

// CompareVersions compares two semver strings (major.minor.patch).
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
// Malformed versions sort before well-formed ones.
func CompareVersions(a, b string) int {
	ap := parseVersion(a)
	bp := parseVersion(b)
	for i := range 3 {
		if ap[i] < bp[i] {
			return -1
		}
		if ap[i] > bp[i] {
			return 1
		}
	}
	return 0
}

// VersionCompatible reports whether peerVersion >= minRequired.
func VersionCompatible(peerVersion, minRequired string) bool {
	if peerVersion == "" {
		return false
	}
	return CompareVersions(peerVersion, minRequired) >= 0
}

// VersionMismatchError returns a human-readable error for version incompatibility.
func VersionMismatchError(peerKind, peerVersion, minRequired string) error {
	label := peerVersion
	if label == "" {
		label = "(unknown)"
	}
	return fmt.Errorf(
		"%s version %s is too old (minimum required: %s) — upgrade the %s",
		peerKind, label, minRequired, peerKind,
	)
}

func parseVersion(s string) [3]int {
	var v [3]int
	parts := strings.SplitN(s, ".", 4) //nolint:mnd // major.minor.patch + overflow
	if len(parts) > 0 {
		v[0], _ = strconv.Atoi(parts[0])
	}
	if len(parts) > 1 {
		v[1], _ = strconv.Atoi(parts[1])
	}
	if len(parts) > 2 {
		v[2], _ = strconv.Atoi(parts[2])
	}
	return v
}
