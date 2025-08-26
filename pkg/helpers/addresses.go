package helpers

import (
	"fmt"

	"github.com/stellar/go/strkey"
	"github.com/stellar/go/xdr"
)

func ContractAddressToScAddress(tokenContractStr string) (xdr.ScAddress, error) {
	var contractAddress xdr.ScAddress
	tokenAddress, err := strkey.Decode(strkey.VersionByteContract, tokenContractStr)
	if err != nil {
		return contractAddress, fmt.Errorf("error decoding token contract: %v", err)
	}

	var tokenAddressHash xdr.ContractId
	copy(tokenAddressHash[:], tokenAddress)

	contractAddress, err = xdr.NewScAddress(xdr.ScAddressTypeScAddressTypeContract, tokenAddressHash)
	if err != nil {
		return contractAddress, fmt.Errorf("error creating contract address: %v", err)
	}

	return contractAddress, nil
}

func StellarAddressToScAddress(address string) (xdr.ScAddress, error) {
	var scAddress xdr.ScAddress
	// Decode the address string to get the raw bytes
	addressBytes, err := strkey.Decode(strkey.VersionByteAccountID, address)
	if err != nil {
		return scAddress, err
	}

	// Convert to Uint256 for the AccountId
	var addressUint256 xdr.Uint256
	copy(addressUint256[:], addressBytes)

	balanceAccount, err := xdr.NewAccountId(xdr.PublicKeyTypePublicKeyTypeEd25519, addressUint256)
	if err != nil {
		return scAddress, err
	}

	scAddress, err = xdr.NewScAddress(xdr.ScAddressTypeScAddressTypeAccount, balanceAccount)
	if err != nil {
		return scAddress, err
	}

	return scAddress, nil
}

func EncodeContractAddress(contractId xdr.ContractId) (string, error) {
	encodedStr, err := strkey.Encode(strkey.VersionByteContract, contractId[:])
	if err != nil {
		return "", fmt.Errorf("failed to encode address: %v", err)
	}
	return encodedStr, nil
}
