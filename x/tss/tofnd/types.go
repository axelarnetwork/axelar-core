package tofnd

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate checks if the sign result is valid
func (m *MessageOut_SignResult) Validate() error {
	if signature := m.GetSignature(); signature != nil {
		if _, err := btcec.ParseDERSignature(signature, btcec.S256()); err != nil {
			return err
		}

		return nil
	}

	if criminals := m.GetCriminals(); criminals != nil {
		if len(criminals.Criminals) == 0 {
			return fmt.Errorf("missing criminals")
		}

		criminalSeen := make(map[string]bool)

		for i, criminal := range criminals.Criminals {
			if criminalSeen[criminal.String()] {
				return fmt.Errorf("duplicate criminal %s", criminal.String())
			}

			_, err := sdk.ValAddressFromBech32(criminal.GetPartyUid())
			if err != nil {
				return fmt.Errorf("invalid criminal address %s", criminal.GetPartyUid())
			}

			if criminal.CrimeType != CRIME_TYPE_MALICIOUS && criminal.CrimeType != CRIME_TYPE_NON_MALICIOUS {
				return fmt.Errorf("invalid crime type %s", criminal.CrimeType.String())
			}

			if i < len(criminals.Criminals)-1 && !criminals.Less(i, i+1) {
				return fmt.Errorf("criminals have to be sorted in ascending order")
			}

			criminalSeen[criminal.String()] = true
		}

		return nil
	}

	return fmt.Errorf("missing signature or criminals")
}

// Len returns the number of criminals in the criminal list
func (m *MessageOut_CriminalList) Len() int {
	return len(m.Criminals)
}

// Swap swaps the criminals at given indexes in the criminal list
func (m *MessageOut_CriminalList) Swap(i, j int) {
	m.Criminals[i], m.Criminals[j] = m.Criminals[j], m.Criminals[i]
}

// Less returns true if the criminal at index i is considered less than the criminal at index j in the criminal list
func (m *MessageOut_CriminalList) Less(i, j int) bool {
	return m.Criminals[i].String() < m.Criminals[j].String()
}
