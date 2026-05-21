package bookref

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// ABSItemPrefix matches the real-ABS "li_..." item identifier convention.
// We prefix outbound IDs with this so the official ABS clients see the
// same shape they'd see talking to a stock ABS server. Decode strips the
// prefix when it's present; IDs without the prefix (legacy emissions,
// admin-pasted URLs, host-proxied SPA requests) still decode correctly.
const ABSItemPrefix = "li_"

// Encode prefixes backend-scoped book IDs with the presentation library ID
// and the "li_" sentinel real ABS clients expect. Backwards compatible:
// the legacy "<libraryID>:<base64>" form still decodes via Decode below.
func Encode(libraryID int64, backendBookID string) string {
	if libraryID <= 0 {
		return ABSItemPrefix + backendBookID
	}
	return ABSItemPrefix + fmt.Sprintf("%d:%s", libraryID, base64.RawURLEncoding.EncodeToString([]byte(backendBookID)))
}

// Decode returns the presentation library ID and backend book ID. Accepts
// three shapes:
//
//   - "li_<libraryID>:<base64>" — current canonical form
//   - "<libraryID>:<base64>"    — pre-prefix legacy form, still emitted in
//     URLs the operator may have bookmarked
//   - "<raw>" / "li_<raw>"      — backend-native ID with no library prefix
//     (libraryID 0, encoded=false)
//
// The encoded boolean reports whether the input carried a library prefix;
// callers use it to decide whether to re-encode the result before round-
// tripping it back into a response.
func Decode(ref string) (int64, string, bool) {
	ref = strings.TrimPrefix(ref, ABSItemPrefix)
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
