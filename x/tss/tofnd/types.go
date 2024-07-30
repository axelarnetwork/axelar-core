package tofnd

import (
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	ec "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate checks if the criminal list is valid
func (m *MessageOut_CriminalList) Validate() error {
	if len(m.Criminals) == 0 {
		return errors.New("missing criminals")
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
			return errors.New("criminals have to be sorted in ascending order")
		}

		criminalSeen[criminal.String()] = true
	}

	return nil
}

// Validate checks if the sign result is valid
func (m *MessageOut_SignResult) Validate() error {
	if signature := m.GetSignature(); signature != nil {
		if _, err := ec.ParseDERSignature(signature); err != nil {
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

	return errors.New("missing signature or criminals")
}

// Validate checks if the keygen result is valid
func (m *MessageOut_KeygenResult) Validate() error {
	if keygenData := m.GetData(); keygenData != nil {
		pubKeyBytes := keygenData.GetPubKey()
		if pubKeyBytes == nil {
			return errors.New("pubkey is nil")
		}
		groupRecoverInfo := keygenData.GetGroupRecoverInfo()
		if groupRecoverInfo == nil {
			return errors.New("group recovery info is nil")
		}
		privateRecoverInfo := keygenData.GetPrivateRecoverInfo()
		if privateRecoverInfo == nil {
			return errors.New("private recovery info is nil")
		}
		_, err := btcec.ParsePubKey(pubKeyBytes)
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

	return errors.New("missing pubkey or criminals")
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
