package utils

import (
	"testing"
	"time"

	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	durationHour := gogoprototypes.DurationProto(time.Hour)
	durationMinute := gogoprototypes.DurationProto(time.Minute)

	actualHour := Hash(durationHour)
	assert.Len(t, actualHour, 32)

	actualMinute := Hash(durationMinute)
	assert.Len(t, actualMinute, 32)

	assert.NotEqual(t, actualHour, actualMinute)
}
