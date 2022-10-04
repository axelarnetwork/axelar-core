package errors_test

import (
	"testing"

	errors2 "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/utils/errors"
)

func TestKeyVals(t *testing.T) {
	var err error = errors.With(errors2.New("test"), "key", "val")
	err = errors2.Wrap(err, "wrapped")

	assert.EqualValues(t, []interface{}{"key", "val"}, errors.KeyVals(err))
}
