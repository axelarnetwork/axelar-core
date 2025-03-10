package errors_test

import (
	goerrors "errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/utils/errors"
)

func TestKeyVals(t *testing.T) {
	var err error = errors.With(goerrors.New("test"), "key", "val")
	err = fmt.Errorf("wrapped: %w", err)

	assert.EqualValues(t, []interface{}{"key", "val"}, errors.KeyVals(err))
}
