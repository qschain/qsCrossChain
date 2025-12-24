package tron

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/fbsobreira/gotron-sdk/pkg/client"

	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"go.uber.org/zap"
)

// ABI
//const AbiABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beacon\",\"type\":\"address\"}],\"name\":\"BeaconUpgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldContract\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newContract\",\"type\":\"address\"}],\"name\":\"ContractUpgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint32\",\"name\":\"index\",\"type\":\"uint32\"}],\"name\":\"GuardianSetAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"sequence\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"nonce\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"consistencyLevel\",\"type\":\"uint8\"}],\"name\":\"LogMessagePublished\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"inputs\":[],\"name\":\"chainId\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCurrentGuardianSetIndex\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"index\",\"type\":\"uint32\"}],\"name\":\"getGuardianSet\",\"outputs\":[{\"components\":[{\"internalType\":\"address[]\",\"name\":\"keys\",\"type\":\"address[]\"},{\"internalType\":\"uint32\",\"name\":\"expirationTime\",\"type\":\"uint32\"}],\"internalType\":\"structStructs.GuardianSet\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getGuardianSetExpiry\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"governanceActionIsConsumed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"governanceChainId\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"governanceContract\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"impl\",\"type\":\"address\"}],\"name\":\"isInitialized\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"messageFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"emitter\",\"type\":\"address\"}],\"name\":\"nextSequence\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedVM\",\"type\":\"bytes\"}],\"name\":\"parseAndVerifyVM\",\"outputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"timestamp\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"nonce\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"emitterChainId\",\"type\":\"uint16\"},{\"internalType\":\"bytes32\",\"name\":\"emitterAddress\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"sequence\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"consistencyLevel\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"guardianSetIndex\",\"type\":\"uint32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"guardianIndex\",\"type\":\"uint8\"}],\"internalType\":\"structStructs.Signature[]\",\"name\":\"signatures\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"internalType\":\"structStructs.VM\",\"name\":\"vm\",\"type\":\"tuple\"},{\"internalType\":\"bool\",\"name\":\"valid\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedUpgrade\",\"type\":\"bytes\"}],\"name\":\"parseContractUpgrade\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"module\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"action\",\"type\":\"uint8\"},{\"internalType\":\"uint16\",\"name\":\"chain\",\"type\":\"uint16\"},{\"internalType\":\"address\",\"name\":\"newContract\",\"type\":\"address\"}],\"internalType\":\"structGovernanceStructs.ContractUpgrade\",\"name\":\"cu\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedUpgrade\",\"type\":\"bytes\"}],\"name\":\"parseGuardianSetUpgrade\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"module\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"action\",\"type\":\"uint8\"},{\"internalType\":\"uint16\",\"name\":\"chain\",\"type\":\"uint16\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"keys\",\"type\":\"address[]\"},{\"internalType\":\"uint32\",\"name\":\"expirationTime\",\"type\":\"uint32\"}],\"internalType\":\"structStructs.GuardianSet\",\"name\":\"newGuardianSet\",\"type\":\"tuple\"},{\"internalType\":\"uint32\",\"name\":\"newGuardianSetIndex\",\"type\":\"uint32\"}],\"internalType\":\"structGovernanceStructs.GuardianSetUpgrade\",\"name\":\"gsu\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedSetMessageFee\",\"type\":\"bytes\"}],\"name\":\"parseSetMessageFee\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"module\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"action\",\"type\":\"uint8\"},{\"internalType\":\"uint16\",\"name\":\"chain\",\"type\":\"uint16\"},{\"internalType\":\"uint256\",\"name\":\"messageFee\",\"type\":\"uint256\"}],\"internalType\":\"structGovernanceStructs.SetMessageFee\",\"name\":\"smf\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedTransferFees\",\"type\":\"bytes\"}],\"name\":\"parseTransferFees\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"module\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"action\",\"type\":\"uint8\"},{\"internalType\":\"uint16\",\"name\":\"chain\",\"type\":\"uint16\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"recipient\",\"type\":\"bytes32\"}],\"internalType\":\"structGovernanceStructs.TransferFees\",\"name\":\"tf\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedVM\",\"type\":\"bytes\"}],\"name\":\"parseVM\",\"outputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"timestamp\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"nonce\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"emitterChainId\",\"type\":\"uint16\"},{\"internalType\":\"bytes32\",\"name\":\"emitterAddress\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"sequence\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"consistencyLevel\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"guardianSetIndex\",\"type\":\"uint32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"guardianIndex\",\"type\":\"uint8\"}],\"internalType\":\"structStructs.Signature[]\",\"name\":\"signatures\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"internalType\":\"structStructs.VM\",\"name\":\"vm\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_vm\",\"type\":\"bytes\"}],\"name\":\"submitContractUpgrade\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_vm\",\"type\":\"bytes\"}],\"name\":\"submitNewGuardianSet\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_vm\",\"type\":\"bytes\"}],\"name\":\"submitSetMessageFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_vm\",\"type\":\"bytes\"}],\"name\":\"submitTransferFees\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"guardianIndex\",\"type\":\"uint8\"}],\"internalType\":\"structStructs.Signature[]\",\"name\":\"signatures\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"keys\",\"type\":\"address[]\"},{\"internalType\":\"uint32\",\"name\":\"expirationTime\",\"type\":\"uint32\"}],\"internalType\":\"structStructs.GuardianSet\",\"name\":\"guardianSet\",\"type\":\"tuple\"}],\"name\":\"verifySignatures\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"valid\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"timestamp\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"nonce\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"emitterChainId\",\"type\":\"uint16\"},{\"internalType\":\"bytes32\",\"name\":\"emitterAddress\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"sequence\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"consistencyLevel\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"guardianSetIndex\",\"type\":\"uint32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"guardianIndex\",\"type\":\"uint8\"}],\"internalType\":\"structStructs.Signature[]\",\"name\":\"signatures\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"internalType\":\"structStructs.VM\",\"name\":\"vm\",\"type\":\"tuple\"}],\"name\":\"verifyVM\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"valid\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"nonce\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint8\",\"name\":\"consistencyLevel\",\"type\":\"uint8\"}],\"name\":\"publishMessage\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"sequence\",\"type\":\"uint64\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"initialGuardians\",\"type\":\"address[]\"},{\"internalType\":\"uint16\",\"name\":\"chainId\",\"type\":\"uint16\"},{\"internalType\":\"uint16\",\"name\":\"governanceChainId\",\"type\":\"uint16\"},{\"internalType\":\"bytes32\",\"name\":\"governanceContract\",\"type\":\"bytes32\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

const AbiABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beacon\",\"type\":\"address\"}],\"name\":\"BeaconUpgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldContract\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newContract\",\"type\":\"address\"}],\"name\":\"ContractUpgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint32\",\"name\":\"index\",\"type\":\"uint32\"}],\"name\":\"GuardianSetAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"sequence\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"nonce\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"consistencyLevel\",\"type\":\"uint8\"}],\"name\":\"LogMessagePublished\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"inputs\":[],\"name\":\"chainId\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"evmChainId\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCurrentGuardianSetIndex\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"index\",\"type\":\"uint32\"}],\"name\":\"getGuardianSet\",\"outputs\":[{\"components\":[{\"internalType\":\"address[]\",\"name\":\"keys\",\"type\":\"address[]\"},{\"internalType\":\"uint32\",\"name\":\"expirationTime\",\"type\":\"uint32\"}],\"internalType\":\"struct Structs.GuardianSet\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getGuardianSetExpiry\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"governanceActionIsConsumed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"governanceChainId\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"governanceContract\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"isFork\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"impl\",\"type\":\"address\"}],\"name\":\"isInitialized\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"messageFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"emitter\",\"type\":\"address\"}],\"name\":\"nextSequence\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedVM\",\"type\":\"bytes\"}],\"name\":\"parseAndVerifyVM\",\"outputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"timestamp\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"nonce\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"emitterChainId\",\"type\":\"uint16\"},{\"internalType\":\"bytes32\",\"name\":\"emitterAddress\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"sequence\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"consistencyLevel\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"guardianSetIndex\",\"type\":\"uint32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"guardianIndex\",\"type\":\"uint8\"}],\"internalType\":\"struct Structs.Signature[]\",\"name\":\"signatures\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"internalType\":\"struct Structs.VM\",\"name\":\"vm\",\"type\":\"tuple\"},{\"internalType\":\"bool\",\"name\":\"valid\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedUpgrade\",\"type\":\"bytes\"}],\"name\":\"parseContractUpgrade\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"module\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"action\",\"type\":\"uint8\"},{\"internalType\":\"uint16\",\"name\":\"chain\",\"type\":\"uint16\"},{\"internalType\":\"address\",\"name\":\"newContract\",\"type\":\"address\"}],\"internalType\":\"struct GovernanceStructs.ContractUpgrade\",\"name\":\"cu\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedUpgrade\",\"type\":\"bytes\"}],\"name\":\"parseGuardianSetUpgrade\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"module\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"action\",\"type\":\"uint8\"},{\"internalType\":\"uint16\",\"name\":\"chain\",\"type\":\"uint16\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"keys\",\"type\":\"address[]\"},{\"internalType\":\"uint32\",\"name\":\"expirationTime\",\"type\":\"uint32\"}],\"internalType\":\"struct Structs.GuardianSet\",\"name\":\"newGuardianSet\",\"type\":\"tuple\"},{\"internalType\":\"uint32\",\"name\":\"newGuardianSetIndex\",\"type\":\"uint32\"}],\"internalType\":\"struct GovernanceStructs.GuardianSetUpgrade\",\"name\":\"gsu\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedRecoverChainId\",\"type\":\"bytes\"}],\"name\":\"parseRecoverChainId\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"module\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"action\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"evmChainId\",\"type\":\"uint256\"},{\"internalType\":\"uint16\",\"name\":\"newChainId\",\"type\":\"uint16\"}],\"internalType\":\"struct GovernanceStructs.RecoverChainId\",\"name\":\"rci\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedSetMessageFee\",\"type\":\"bytes\"}],\"name\":\"parseSetMessageFee\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"module\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"action\",\"type\":\"uint8\"},{\"internalType\":\"uint16\",\"name\":\"chain\",\"type\":\"uint16\"},{\"internalType\":\"uint256\",\"name\":\"messageFee\",\"type\":\"uint256\"}],\"internalType\":\"struct GovernanceStructs.SetMessageFee\",\"name\":\"smf\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedTransferFees\",\"type\":\"bytes\"}],\"name\":\"parseTransferFees\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"module\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"action\",\"type\":\"uint8\"},{\"internalType\":\"uint16\",\"name\":\"chain\",\"type\":\"uint16\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"recipient\",\"type\":\"bytes32\"}],\"internalType\":\"struct GovernanceStructs.TransferFees\",\"name\":\"tf\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedVM\",\"type\":\"bytes\"}],\"name\":\"parseVM\",\"outputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"timestamp\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"nonce\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"emitterChainId\",\"type\":\"uint16\"},{\"internalType\":\"bytes32\",\"name\":\"emitterAddress\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"sequence\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"consistencyLevel\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"guardianSetIndex\",\"type\":\"uint32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"guardianIndex\",\"type\":\"uint8\"}],\"internalType\":\"struct Structs.Signature[]\",\"name\":\"signatures\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"internalType\":\"struct Structs.VM\",\"name\":\"vm\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"nonce\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint8\",\"name\":\"consistencyLevel\",\"type\":\"uint8\"}],\"name\":\"publishMessage\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"sequence\",\"type\":\"uint64\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"numGuardians\",\"type\":\"uint256\"}],\"name\":\"quorum\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"numSignaturesRequiredForQuorum\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_vm\",\"type\":\"bytes\"}],\"name\":\"submitContractUpgrade\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_vm\",\"type\":\"bytes\"}],\"name\":\"submitNewGuardianSet\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_vm\",\"type\":\"bytes\"}],\"name\":\"submitRecoverChainId\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_vm\",\"type\":\"bytes\"}],\"name\":\"submitSetMessageFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_vm\",\"type\":\"bytes\"}],\"name\":\"submitTransferFees\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"guardianIndex\",\"type\":\"uint8\"}],\"internalType\":\"struct Structs.Signature[]\",\"name\":\"signatures\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"keys\",\"type\":\"address[]\"},{\"internalType\":\"uint32\",\"name\":\"expirationTime\",\"type\":\"uint32\"}],\"internalType\":\"struct Structs.GuardianSet\",\"name\":\"guardianSet\",\"type\":\"tuple\"}],\"name\":\"verifySignatures\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"valid\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"timestamp\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"nonce\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"emitterChainId\",\"type\":\"uint16\"},{\"internalType\":\"bytes32\",\"name\":\"emitterAddress\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"sequence\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"consistencyLevel\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"guardianSetIndex\",\"type\":\"uint32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"guardianIndex\",\"type\":\"uint8\"}],\"internalType\":\"struct Structs.Signature[]\",\"name\":\"signatures\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"internalType\":\"struct Structs.VM\",\"name\":\"vm\",\"type\":\"tuple\"}],\"name\":\"verifyVM\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"valid\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]"

// 测试ABI
//const AbiABI = "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"initialValue\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"}],\"name\":\"DataStored\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"}],\"name\":\"store\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

//const AbiABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"}],\"name\":\"DataStored\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"get111\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"}],\"name\":\"store111\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"storedData\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// TronContract 波场合约客户端(整合Filterer和caller功能)
type TronContractClient struct {
	ContractAddress string
	Client          *client.GrpcClient
	ABI             string // 可选：合约ABI，用于事件解析
	Logger          *zap.Logger
}

// ContractEvent 合约事件结构
type ContractEvent struct {
	// 基础信息
	EventName          string // 事件名称
	ContractAddress    string // 合约地址（Base58格式）
	ContractHexAddress string // 合约地址（十六进制）

	// 事件数据
	RawTopics    [][]byte               // 原始主题
	RawData      []byte                 // 原始数据
	ParsedParams map[string]interface{} // 解析后的参数

	// 区块信息
	BlockNumber    int64  // 区块号
	BlockTimestamp int64  // 区块时间戳
	TxHash         string // 交易哈希
	TxIndex        uint   // 交易索引
	LogIndex       uint   // 日志索引
}

// TronEventFilter 波场事件过滤器（对应以太坊的Filterer）
type TronEventFilter struct {
	ContractAddress string
	EventName       string
	Topics          []byte // 事件主题
	StartBlock      int64
	EndBlock        *int64
}

// NewTronContractClient 创建波场合约客户端
func NewTronContractClient(contractAddress string, grpcClient *client.GrpcClient, logger *zap.Logger) *TronContractClient {
	return &TronContractClient{
		ContractAddress: contractAddress,
		Client:          grpcClient,
		Logger:          logger,
	}
}

// FilterEvents 过滤合约事件（类似以太坊的Filterer）
func (c *TronContractClient) FilterEvents(ctx context.Context, filter *TronEventFilter) ([]*ContractEvent, error) {
	var events []*ContractEvent

	// 确定结束区块
	endBlock := filter.EndBlock
	if endBlock == nil {
		// 获取最新区块
		block, err := c.Client.GetNowBlock()
		if err != nil {
			return nil, fmt.Errorf("failed to get latest block: %w", err)
		}
		currentBlock := block.GetBlockHeader().GetRawData().GetNumber()
		endBlock = &currentBlock
	}

	// 遍历区块范围内的交易
	for blockNum := filter.StartBlock; blockNum <= *endBlock; blockNum++ {
		block, err := c.Client.GetBlockByNum(blockNum)
		if err != nil {
			c.Logger.Warn("failed to get block", zap.Int64("block", blockNum), zap.Error(err))
			continue
		}
		// 获取区块时间戳
		blockTimestamp := block.GetBlockHeader().GetRawData().GetTimestamp()
		// 检查区块中的交易
		for txIndex, tx := range block.Transactions {
			if tx.Transaction.GetRawData() == nil {
				continue
			}
			txInfo, err := c.Client.GetTransactionInfoByID(common.BytesToHexString(tx.Txid))
			if err != nil {
				fmt.Println(err)
				continue
			}
			contractAddrStr, err := ContractAddressToBase58(txInfo)
			if err != nil {
				fmt.Println(err)
				continue
			}
			// 检查是否为目标合约的交易
			if contractAddrStr != filter.ContractAddress {
				continue
			}
			if len(txInfo.Log) > 0 {
				parsedEvents, err := c.parseLogs(
					txInfo.Log,
					filter,
					blockNum,
					blockTimestamp,
					common.BytesToHexString(tx.Txid),
					uint(txIndex),
				)
				if err != nil {
					c.Logger.Debug("failed to parse logs", zap.Error(err))
					continue
				}
				events = append(events, parsedEvents...)
			}
		}

		// 考虑性能，可以在这里添加批处理逻辑
	}

	return events, nil
}
func (c *TronContractClient) FilterEvents1(ctx context.Context, filter *TronEventFilter, implementAbi *abi.ABI) ([]*types.Log, error) {
	//var events []*ContractEvent
	var ethLogs []*types.Log
	// 确定结束区块
	endBlock := filter.EndBlock
	if endBlock == nil {
		// 获取最新区块
		block, err := c.Client.GetNowBlock()
		if err != nil {
			return nil, fmt.Errorf("failed to get latest block: %w", err)
		}
		currentBlock := block.GetBlockHeader().GetRawData().GetNumber()
		endBlock = &currentBlock
	}

	// 遍历区块范围内的交易
	for blockNum := filter.StartBlock; blockNum <= *endBlock; blockNum++ {
		block, err := c.Client.GetBlockByNum(blockNum)
		if err != nil {
			c.Logger.Warn("failed to get block", zap.Int64("block", blockNum), zap.Error(err))
			continue
		}
		// 获取区块时间戳
		blockTimestamp := block.GetBlockHeader().GetRawData().GetTimestamp()
		// 检查区块中的交易
		for txIndex, tx := range block.Transactions {
			if tx.Transaction.GetRawData() == nil {
				continue
			}
			txInfo, err := c.Client.GetTransactionInfoByID(common.BytesToHexString(tx.Txid))
			if err != nil {
				fmt.Println(err)
				continue
			}
			if len(txInfo.Log) > 0 {
				ethLog, err := c.parseLogs1(
					txInfo.Log,
					filter,
					blockNum,
					blockTimestamp,
					common.BytesToHexString(tx.Txid),
					uint(txIndex),
					implementAbi,
				)
				if err != nil {
					c.Logger.Debug("failed to parse logs", zap.Error(err))
					continue
				}
				ethLogs = append(ethLogs, ethLog...)
			}
		}

		// 考虑性能，可以在这里添加批处理逻辑
	}

	return ethLogs, nil
}

// WatchEvents 监听合约事件（持续监听）
func (c *TronContractClient) WatchEvents1(ctx context.Context, filter *TronEventFilter, logChan chan<- *types.Log, implementAbi *abi.ABI) error {
	var lastBlockNum int64 = 0
	for {
		lastCheckedBlock, err := c.Client.Client.GetNowBlock(ctx, nil)
		if err != nil {
			c.Logger.Error("failed to get latest block", zap.Error(err))
			time.Sleep(2 * time.Second)
			continue
		}
		//只监听固化块
		lastBlockNum = lastCheckedBlock.GetBlockHeader().GetRawData().GetNumber() - SolidNum
		if lastBlockNum >= 1 {
			break
		}
	}
	// 持续轮询新区块
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			currentBlock, err := c.Client.GetNowBlock()
			if err != nil {
				c.Logger.Error("failed to get latest block", zap.Error(err))
				time.Sleep(2 * time.Second)
				continue
			}
			currentBlockNum := currentBlock.GetBlockHeader().GetRawData().GetNumber()
			//计算当前固化快的高度
			currentBlockNum = currentBlockNum - SolidNum
			if currentBlockNum < 1 {
				continue
			}

			if currentBlockNum > lastBlockNum {
				// 检查新产生的区块
				for blockNum := lastBlockNum + 1; blockNum <= currentBlockNum; blockNum++ {
					filter.StartBlock = blockNum
					filter.EndBlock = &blockNum

					events, err := c.FilterEvents1(ctx, filter, implementAbi)
					if err != nil {
						c.Logger.Error("failed to filter events", zap.Error(err))
						continue
					}

					// 发送事件到通道
					for _, event := range events {
						select {
						case logChan <- event:
						case <-ctx.Done():
							return nil
						default:
							c.Logger.Warn("event channel is full, dropping event")
						}
					}
				}

				lastBlockNum = currentBlockNum
			}

			time.Sleep(2 * time.Second)
		}
	}
}

// WatchEvents 监听合约事件（持续监听）
func (c *TronContractClient) WatchEvents(ctx context.Context, filter *TronEventFilter, eventChan chan<- *ContractEvent) error {
	lastCheckedBlock, err := c.Client.Client.GetNowBlock(ctx, nil)
	if err != nil {
		return err
	}
	lastBlockNum := lastCheckedBlock.GetBlockHeader().GetRawData().GetNumber()

	// 持续轮询新区块
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			currentBlock, err := c.Client.GetNowBlock()
			if err != nil {
				c.Logger.Error("failed to get latest block", zap.Error(err))
				time.Sleep(2 * time.Second)
				continue
			}
			currentBlockNum := currentBlock.GetBlockHeader().GetRawData().GetNumber()
			if currentBlockNum > lastBlockNum {
				// 检查新产生的区块
				for blockNum := lastBlockNum + 1; blockNum <= currentBlockNum; blockNum++ {
					filter.StartBlock = blockNum
					filter.EndBlock = &blockNum

					events, err := c.FilterEvents(ctx, filter)
					if err != nil {
						c.Logger.Error("failed to filter events", zap.Error(err))
						continue
					}

					// 发送事件到通道
					for _, event := range events {
						select {
						case eventChan <- event:
						case <-ctx.Done():
							return nil
						default:
							c.Logger.Warn("event channel is full, dropping event")
						}
					}
				}

				lastBlockNum = currentBlockNum
			}

			time.Sleep(2 * time.Second)
		}
	}
}

// CallConstantFunction 调用只读函数（对应以太坊的Caller）
func (c *TronContractClient) CallConstantFunction(ctx context.Context, functionSelector string, parameters string) (*api.TransactionExtention, error) {
	// 触发常量合约调用（不上链）
	result, err := c.Client.TriggerConstantContract(
		"",
		c.ContractAddress,
		functionSelector,
		parameters,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to call constant function: %w", err)
	}

	// 检查调用结果
	if result.Result != nil && !(result.Result.Code == api.Return_SUCCESS) {
		return nil, fmt.Errorf("contract call failed: %s", result.Result.Message)
	}

	return result, nil
}

func (c *TronContractClient) parseLogs1(
	logs []*core.TransactionInfo_Log,
	filter *TronEventFilter,
	blockNumber int64,
	blockTimestamp int64,
	txHash string,
	txIndex uint,
	implementAbi *abi.ABI,
) ([]*types.Log, error) {
	var ethLogs []*types.Log
	for _, log := range logs {
		if log == nil || len(log.Topics) == 0 {
			continue
		}
		// 将波场日志转换为以太坊日志格式
		ethLog, err := c.convertTronLogToEthLog(log, blockNumber)
		if err != nil {
			c.Logger.Debug("failed to convert tron log", zap.Error(err))
			continue
		}
		if ethLog != nil {
			if ethLog.Topics[0] == implementAbi.Events[filter.EventName].ID {
				ethLogs = append(ethLogs, ethLog)
			}
		}
	}
	return ethLogs, nil
}

func (c *TronContractClient) parseLogs(
	logs []*core.TransactionInfo_Log,
	filter *TronEventFilter,
	blockNumber int64,
	blockTimestamp int64,
	txHash string,
	txIndex uint,
) ([]*ContractEvent, error) {
	var events []*ContractEvent
	for logIndex, log := range logs {
		if log == nil || len(log.Topics) == 0 {
			continue
		}
		// 将波场日志转换为以太坊日志格式
		ethLog, err := c.convertTronLogToEthLog(log, blockNumber)
		if err != nil {
			c.Logger.Debug("failed to convert tron log", zap.Error(err))
			continue
		}
		// 尝试解析事件
		event, err := c.parseSingleLog(ethLog, filter, blockNumber, blockTimestamp, txHash, txIndex, uint(logIndex))
		if err != nil {
			c.Logger.Debug("failed to parse log", zap.Error(err))
			continue
		}
		if event != nil {
			events = append(events, event)
		}
	}
	return events, nil
}

// convertTronLogToEthLog 将波场日志转换为以太坊日志格式
func (c *TronContractClient) convertTronLogToEthLog(tronLog *core.TransactionInfo_Log, blockNumber int64) (*types.Log, error) {
	if tronLog == nil {
		return nil, fmt.Errorf("log is nil")
	}

	// 转换地址（从20字节到common.Address）
	var address ethCommon.Address
	if len(tronLog.Address) >= 20 {
		// 取最后20字节（波场地址可能是21字节，第一个字节是网络前缀）
		startIdx := len(tronLog.Address) - 20
		if startIdx < 0 {
			startIdx = 0
		}
		address = ethCommon.BytesToAddress(tronLog.Address[startIdx:])
	} else if len(tronLog.Address) > 0 {
		// 如果不足20字节，补0
		padded := make([]byte, 20)
		copy(padded[20-len(tronLog.Address):], tronLog.Address)
		address = ethCommon.BytesToAddress(padded)
	}
	// 转换Topics
	topics := make([]ethCommon.Hash, len(tronLog.Topics))
	for i, topic := range tronLog.Topics {
		if len(topic) == 0 {
			continue
		}

		// 确保topic是32字节
		var topicHash ethCommon.Hash
		if len(topic) == 32 {
			copy(topicHash[:], topic)
		} else {
			// 如果不足32字节，补0
			padded := make([]byte, 32)
			copy(padded[32-len(topic):], topic)
			copy(topicHash[:], padded)
		}
		topics[i] = topicHash
	}

	return &types.Log{
		Address:     address,
		Topics:      topics,
		Data:        tronLog.Data,
		BlockNumber: uint64(blockNumber),
	}, nil
}

// 构建事件签名
func BuildEventSignature(event abi.Event) string {
	var params []string
	for _, input := range event.Inputs {
		params = append(params, input.Type.String())
	}
	return fmt.Sprintf("%s(%s)", event.Name, strings.Join(params, ","))
}

// parseSingleLog 解析单个日志
func (c *TronContractClient) parseSingleLog(
	ethLog *types.Log,
	filter *TronEventFilter,
	blockNumber int64,
	blockTimestamp int64,
	txHash string,
	txIndex uint,
	logIndex uint,
) (*ContractEvent, error) {
	parsedABI, err := abi.JSON(strings.NewReader(AbiABI))
	// 如果没有Topics，无法解析事件
	if len(ethLog.Topics) == 0 {
		return nil, fmt.Errorf("log has no topics")
	}

	// 获取事件签名哈希（Topics[0]）
	eventSigHash := ethLog.Topics[0]
	// 根据事件签名哈希查找对应的事件
	var eventABI abi.Event
	found := false

	for _, event := range parsedABI.Events {
		// 计算事件签名的哈希
		eventSig := BuildEventSignature(event)
		calculatedSigHash := crypto.Keccak256Hash([]byte(eventSig))
		if calculatedSigHash == eventSigHash {
			eventABI = event
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("event not found in ABI for signature hash: %s", eventSigHash.Hex())
	}

	// 如果指定了事件名称过滤，检查是否匹配
	if filter.EventName != "" && eventABI.Name != filter.EventName {
		return nil, fmt.Errorf("event name mismatch: expected %s, got %s", filter.EventName, eventABI.Name)
	}

	// 检查Topics过滤
	/*if filter.Topics != nil {
		if !c.matchTopics(ethLog.Topics[1:], filter.Topics) {
			return nil, fmt.Errorf("topics do not match filter")
		}
	}*/

	// 解析事件参数
	parsedData, err := c.parseEventData(ethLog, eventABI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse event data: %w", err)
	}

	// 转换合约地址格式
	contractBase58, err := c.hexAddressToBase58(ethLog.Address.Hex())
	if err != nil {
		c.Logger.Debug("failed to convert contract address", zap.Error(err))
	}

	// 构建事件对象
	event := &ContractEvent{
		EventName:          eventABI.Name,
		ContractAddress:    contractBase58,
		ContractHexAddress: ethLog.Address.Hex(),
		RawTopics:          c.convertEthTopicsToBytes(ethLog.Topics),
		RawData:            ethLog.Data,
		ParsedParams:       parsedData,
		BlockNumber:        blockNumber,
		BlockTimestamp:     blockTimestamp,
		TxHash:             txHash,
		TxIndex:            txIndex,
		LogIndex:           logIndex,
	}

	return event, nil
}

// parseEventData 解析事件数据
func (c *TronContractClient) parseEventData(ethLog *types.Log, eventABI abi.Event) (map[string]interface{}, error) {
	parsedData := make(map[string]interface{})

	// 使用ABI解析事件
	unpackedData, err := eventABI.Inputs.Unpack(ethLog.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack event data: %w", err)
	}

	// 处理索引参数（Topics）
	indexedParams := 0
	for i, input := range eventABI.Inputs {
		if input.Indexed {
			// 索引参数在Topics中（从Topics[1]开始）
			topicIdx := indexedParams + 1
			if topicIdx < len(ethLog.Topics) {
				topic := ethLog.Topics[topicIdx]

				var value interface{}
				switch input.Type.T {
				case abi.AddressTy:
					// 地址类型，取最后20字节
					addr := ethCommon.BytesToAddress(topic[12:])
					value = addr.Hex()
				case abi.IntTy, abi.UintTy:
					// 整数类型
					intValue := new(big.Int).SetBytes(topic[:])
					value = intValue
				case abi.BoolTy:
					// 布尔类型
					intValue := new(big.Int).SetBytes(topic[:])
					value = intValue.Uint64() != 0
				case abi.FixedBytesTy:
					// 固定字节数组
					size := input.Type.Size
					if size > 32 {
						size = 32
					}
					value = topic[:size]
				case abi.BytesTy, abi.StringTy:
					// 动态类型在data中，不在topic中
					continue
				default:
					// 其他类型
					value = topic[:]
				}

				parsedData[input.Name] = value
				indexedParams++
			}
		} else {
			// 非索引参数在Data中
			if i-indexedParams < len(unpackedData) {
				value := unpackedData[i-indexedParams]

				// 转换常见类型
				switch v := value.(type) {
				case ethCommon.Address:
					parsedData[input.Name] = v.Hex()
				case *big.Int:
					parsedData[input.Name] = v
				case []byte:
					parsedData[input.Name] = hex.EncodeToString(v)
				default:
					parsedData[input.Name] = v
				}
			}
		}
	}

	return parsedData, nil
}

// matchTopics 检查Topics是否匹配过滤器
func (c *TronContractClient) matchTopics(logTopics []ethCommon.Hash, filterTopics [][]byte) bool {
	if len(filterTopics) == 0 {
		return true
	}

	if len(logTopics) < len(filterTopics) {
		return false
	}

	for i, filterTopic := range filterTopics {
		if filterTopic == nil {
			continue // nil表示匹配任意值
		}

		if i >= len(logTopics) {
			return false
		}

		// 将过滤器topic转换为common.Hash
		var filterHash ethCommon.Hash
		if len(filterTopic) == 32 {
			copy(filterHash[:], filterTopic)
		} else {
			// 如果不足32字节，补0
			padded := make([]byte, 32)
			copy(padded[32-len(filterTopic):], filterTopic)
			copy(filterHash[:], padded)
		}

		if logTopics[i] != filterHash {
			return false
		}
	}

	return true
}

// hexAddressToBase58 将十六进制地址转换为Base58格式
func (c *TronContractClient) hexAddressToBase58(hexAddr string) (string, error) {
	// 去掉0x前缀
	cleanHex := strings.TrimPrefix(hexAddr, "0x")

	if len(cleanHex) == 0 {
		return "", fmt.Errorf("empty address")
	}

	// 十六进制解码
	addrBytes, err := hex.DecodeString(cleanHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex: %w", err)
	}

	// 确保是20字节
	if len(addrBytes) != 20 {
		return "", fmt.Errorf("expected 20-byte address, got %d bytes", len(addrBytes))
	}

	// 添加主网前缀 (0x41)
	prefixed := make([]byte, 21)
	prefixed[0] = 0x41 // 波场主网前缀
	copy(prefixed[1:], addrBytes)

	// Base58Check编码
	return common.EncodeCheck(prefixed), nil
}

// convertEthTopicsToBytes 将以太坊Topics转换为字节数组
func (c *TronContractClient) convertEthTopicsToBytes(ethTopics []ethCommon.Hash) [][]byte {
	topics := make([][]byte, len(ethTopics))
	for i, topic := range ethTopics {
		topics[i] = topic[:]
	}
	return topics
}

// BytesToTronBase58 将20字节地址转换为波场Base58格式
func BytesToTronBase58(addressBytes []byte) (string, error) {
	// 检查长度是否为21字节
	if len(addressBytes) != 21 {
		return "", fmt.Errorf("expected 20-byte address, got %d bytes", len(addressBytes))
	}
	// 使用波场SDK的Base58Check编码
	return common.EncodeCheck(addressBytes), nil
}

// ContractAddressToBase58 从TransactionInfo获取合约地址并转换为Base58
func ContractAddressToBase58(txInfo *core.TransactionInfo) (string, error) {
	if txInfo == nil {
		return "", fmt.Errorf("transaction info is nil")
	}
	return BytesToTronBase58(txInfo.ContractAddress)
}

func filter_test() {
	logger, _ := zap.NewDevelopment()

	// 创建gRPC连接
	grpcClient := client.NewGrpcClient("localhost:50051")
	// ... 初始化连接
	err := grpcClient.Start(client.GRPCInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer grpcClient.Stop()

	// 创建合约客户端（整合Filterer和Caller功能）
	contractClient := NewTronContractClient(
		"TKmpcyuqy3cXvUCBn47bcyy1kdfZ7X3bDB", //合约地址
		grpcClient,
		logger,
	)

	ctx := context.Background()

	// 1. 调用只读函数（类似Caller）
	/*result, err := contractClient.CallViewFunction(ctx, "balanceOf", "TXYZ...address...")
	if err != nil {
		logger.Error("Failed to call view function", zap.Error(err))
	}*/

	// 2. 监听事件（类似Filterer）
	filter := &TronEventFilter{
		ContractAddress: "TKmpcyuqy3cXvUCBn47bcyy1kdfZ7X3bDB",
		EventName:       "DataStored",
		StartBlock:      50000000,
	}

	eventChan := make(chan *ContractEvent, 100)
	go func() {
		err := contractClient.WatchEvents(ctx, filter, eventChan)
		if err != nil {
			logger.Error("Failed to watch events", zap.Error(err))
		}
	}()

	// 处理事件
	for event := range eventChan {
		fmt.Println(event)
		logger.Info("Received contract event",
			zap.String("address", event.ContractAddress),
			zap.Int64("block", event.BlockNumber),
		)
	}
}
