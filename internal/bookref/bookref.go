package bookref

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// Encode prefixes backend-scoped book IDs with the presentation library ID.
func Encode(libraryID int64, backendBookID string) string {
	if libraryID <= 0 {
		return backendBookID
	}
	return fmt.Sprintf("%d:%s", libraryID, base64.RawURLEncoding.EncodeToString([]byte(backendBookID)))
}

// Decode returns the presentation library ID and backend book ID. Legacy IDs
// without a library prefix are returned unchanged with libraryID 0.
func Decode(ref string) (int64, string, bool) {
	parts := strings.SplitN(ref, ":", 2)
	if len(parts) != 2 {
		return 0, ref, false
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		return 0, ref, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, ref, false
	}
	return id, string(raw), true
}
