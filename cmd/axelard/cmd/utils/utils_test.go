package utils

import (
	"strconv"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
)

func TestOverwriteFlagDefaults(t *testing.T) {
	strGen := rand.Strings(0, 1000).Distinct()
	intGen := rand.PInt64Gen()

	var (
		stringFlagName, intFlagName string
		defaultString               string
		defaultInt                  int64
	)

	setup := func() *cobra.Command {
		cmd := &cobra.Command{}

		stringFlagName = rand.StrBetween(5, 20)
		defaultString = strGen.Next()

		intFlagName = rand.StrBetween(5, 20)
		defaultInt = intGen.Next()

		cmd.Flags().String(stringFlagName, defaultString, strGen.Next())
		cmd.Flags().Int64(intFlagName, defaultInt, strGen.Next())
		return cmd
	}
	testCases := []struct {
		label     string
		updateVal bool
	}{
		{"only update defaults", false},
		{"also update current value", true},
	}

	repeats := 100

	for _, testCase := range testCases {
		t.Run(testCase.label, testutils.Func(func(t *testing.T) {
			cmd := setup()

			newDefaultString := strGen.Next()
			newDefaultInt := intGen.Next()

			unknownFlag := strGen.Next()

			defaults := map[string]string{
				stringFlagName: newDefaultString,
				intFlagName:    strconv.FormatInt(newDefaultInt, 10),
				unknownFlag:    strGen.Next(),
			}

			OverwriteFlagDefaults(cmd, defaults, testCase.updateVal)

			f1 := cmd.Flags().Lookup(stringFlagName)
			f2 := cmd.Flags().Lookup(intFlagName)

			assert.Equal(t, f1.DefValue, newDefaultString)
			assert.Equal(t, f2.DefValue, strconv.FormatInt(newDefaultInt, 10))

			if testCase.updateVal {
				assert.Equal(t, f1.Value.String(), newDefaultString)
				assert.True(t, f1.Changed)
				assert.Equal(t, f2.Value.String(), strconv.FormatInt(newDefaultInt, 10))
				assert.True(t, f2.Changed)
			} else {
				assert.Equal(t, f1.Value.String(), defaultString)
				assert.False(t, f1.Changed)
				assert.Equal(t, f2.Value.String(), strconv.FormatInt(defaultInt, 10))
				assert.False(t, f2.Changed)
			}

		}).Repeat(repeats))
	}
}
