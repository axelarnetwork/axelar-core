package types

import "bytes"

var pauseSkipMarker = []byte("axelar/emergency-pause-skip")

// PauseSkipMarker returns the marker a proposer puts first in a block to skip the pause in that block
func PauseSkipMarker() []byte {
	return bytes.Clone(pauseSkipMarker)
}

// IsPauseSkipMarker returns true if the tx bytes are the pause skip marker
func IsPauseSkipMarker(tx []byte) bool {
	return bytes.Equal(tx, pauseSkipMarker)
}
