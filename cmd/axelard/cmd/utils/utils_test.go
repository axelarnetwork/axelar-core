package utils

import (
	"strconv"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
)

func TestOverwriteFlagDefaults(t *testing.T) {
	strGen := rand.Strings(0, 1000).Distinct()
	intGen := rand.PInt64Gen()

	cmd := &cobra.Command{}

	stringFlagName := rand.StrBetween(5, 20)
	defaultString := strGen.Next()

	intFlagName := rand.StrBetween(5, 20)
	defaultInt := intGen.Next()

	cmd.Flags().String(stringFlagName, defaultString, strGen.Next())
	cmd.Flags().Int64(intFlagName, defaultInt, strGen.Next())

	newDefaultString := strGen.Next()
	newDefaultInt := intGen.Next()
	unknownFlag := strGen.Next()
	defaults := map[string]string{
		stringFlagName: newDefaultString,
		intFlagName:    strconv.FormatInt(newDefaultInt, 10),
		unknownFlag:    strGen.Next(),
	}

	OverwriteFlagDefaults(cmd, defaults)

	f1 := cmd.Flags().Lookup(stringFlagName)
	f2 := cmd.Flags().Lookup(intFlagName)
	assert.Equal(t, f1.DefValue, newDefaultString)
	assert.Equal(t, f1.Value.String(), defaultString)
	assert.False(t, f1.Changed)
	assert.Equal(t, f2.DefValue, strconv.FormatInt(newDefaultInt, 10))
	assert.Equal(t, f2.Value.String(), strconv.FormatInt(defaultInt, 10))
	assert.False(t, f2.Changed)
}

func TestOverwriteFlagValues(t *testing.T) {
	strGen := rand.Strings(0, 1000).Distinct()
	intGen := rand.PInt64Gen()

	cmd := &cobra.Command{}

	stringFlagName := rand.StrBetween(5, 20)
	defaultString := strGen.Next()

	intFlagName := rand.StrBetween(5, 20)
	defaultInt := intGen.Next()

	cmd.Flags().String(stringFlagName, defaultString, strGen.Next())
	cmd.Flags().Int64(intFlagName, defaultInt, strGen.Next())

	newString := strGen.Next()
	newInt := intGen.Next()
	unknownFlag := strGen.Next()
	values := map[string]string{
		stringFlagName: newString,
		intFlagName:    strconv.FormatInt(newInt, 10),
		unknownFlag:    strGen.Next(),
	}

	OverwriteFlagValues(cmd, values)

	f1 := cmd.Flags().Lookup(stringFlagName)
	f2 := cmd.Flags().Lookup(intFlagName)
	assert.Equal(t, f1.DefValue, defaultString)
	assert.Equal(t, f1.Value.String(), newString)
	assert.True(t, f1.Changed)
	assert.Equal(t, f2.DefValue, strconv.FormatInt(defaultInt, 10))
	assert.Equal(t, f2.Value.String(), strconv.FormatInt(newInt, 10))
	assert.True(t, f2.Changed)
}
