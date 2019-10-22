package core

import (
	"bytes"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/hooks"
)

const metaChainIdentifier uint8 = 255

// IsSmartContractAddress verifies if a set address is of type smart contract
func IsSmartContractAddress(rcvAddress []byte) bool {
	isEmptyAddress := bytes.Equal(rcvAddress, make([]byte, len(rcvAddress)))
	if isEmptyAddress {
		return true
	}

	isSCAddress := bytes.Equal(rcvAddress[:(hooks.NumInitCharactersForScAddress-hooks.VMTypeLen)],
		make([]byte, hooks.NumInitCharactersForScAddress-hooks.VMTypeLen))
	if isSCAddress {
		return true
	}

	return false
}

// IsMetaChainShardId verifies if the identifier is of type metachain
func IsMetaChainShardId(identifier []byte) bool {
	for i := 0; i < len(identifier); i++ {
		if identifier[i] != metaChainIdentifier {
			return false
		}
	}

	return true
}