package tofnd

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate checks if the criminal list is valid
func (m *MessageOut_CriminalList) Validate() error {
	if len(m.Criminals) == 0 {
		return fmt.Errorf("missing criminals")
	}

	criminalSeen := make(map[string]bool)

	for i, criminal := range m.Criminals {
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

		if i < len(m.Criminals)-1 && !m.Less(i, i+1) {
			return fmt.Errorf("criminals have to be sorted in ascending order")
		}

		criminalSeen[criminal.String()] = true
	}

	return nil
}

// Validate checks if the sign result is valid
func (m *MessageOut_SignResult) Validate() error {
	if signature := m.GetSignature(); signature != nil {
		if _, err := btcec.ParseDERSignature(signature, btcec.S256()); err != nil {
			return err
		}

		return nil
	}

	if criminalList := m.GetCriminals(); criminalList != nil {
		if err := criminalList.Validate(); err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("missing signature or criminals")
}

// Validate checks if the keygen result is valid
func (m *MessageOut_KeygenResult) Validate() error {
	if keygenData := m.GetData(); keygenData != nil {
		pubKeyBytes := keygenData.GetPubKey()
		if pubKeyBytes == nil {
			return fmt.Errorf("pubkey is nil")
		}
		groupRecoverInfo := keygenData.GetGroupRecoverInfo()
		if groupRecoverInfo == nil {
			return fmt.Errorf("group recovery info is nil")
		}
		privateRecoverInfo := keygenData.GetPrivateRecoverInfo()
		if privateRecoverInfo == nil {
			return fmt.Errorf("private recovery info is nil")
		}
		_, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
		if err != nil {
			return err
		}

		return nil
	}

	if criminalList := m.GetCriminals(); criminalList != nil {
		if err := criminalList.Validate(); err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("missing pubkey or criminals")
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
