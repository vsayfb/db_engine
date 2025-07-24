package segment

import "github.com/oklog/ulid/v2"

const THRESHOLD = 1024 * 400

func newSegmentFileName() string {
	return ulid.Make().String() + ".bin"
}
