package tron

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
)

// var cclAbi = "[{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"configure\",\"inputs\":[{\"name\":\"config\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getConfiguration\",\"inputs\":[{\"name\":\"emitterAddress\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"ConfigSet\",\"inputs\":[{\"name\":\"emitterAddress\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"config\",\"type\":\"bytes32\",\"indexed\":false,\"internalType\":\"bytes32\"}],\"anonymous\":false}]"
var cclAbi = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"emitterAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"config\",\"type\":\"bytes32\"}],\"name\":\"ConfigSet\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"VERSION\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"config\",\"type\":\"bytes32\"}],\"name\":\"configure\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"emitterAddress\",\"type\":\"address\"}],\"name\":\"getConfiguration\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// GetConfiguration is a free data retrieval call binding the contract method 0xc44b11f7.
//
// Solidity: function getConfiguration(address emitterAddress) view returns(bytes32)
func GetConfiguration(client *client.GrpcClient, emitterAddress ethCommon.Address, cclAddr string) ([32]byte, error) {
	//emitterAddr := emitterAddress.Hex()
	abiCcl, err := abi.JSON(strings.NewReader(cclAbi))

	params := make([]interface{}, 0)
	params = append(params, emitterAddress)
	// 编码函数调用数据
	encodeDataByte, err := EncodeFunctionCall("getConfiguration", params, &abiCcl)
	if err != nil {
		return [32]byte{}, fmt.Errorf("encodeFunctionCall error: %v", err)
	}

	/*encodedData := "0xc44b11f7" + emitterAddr


	encodeDataByte, err := common.DecodeCheck(encodedData)
	if err != nil {
		return [32]byte{}, err
	}*/
	// 创建调用参数
	contractAddressBytes, err := common.DecodeCheck(cclAddr)
	if err != nil {
		return [32]byte{}, fmt.Errorf("ccl decode contractaddress error: %v", err)
	}
	fromAddress := "410000000000000000000000000000000000000000" //设置为默认值
	callerAddressBytes, err := common.HexStringToBytes(fromAddress)
	if err != nil {
		return [32]byte{}, fmt.Errorf("ccl decode fromaddress error: %v", err)
	}
	triggerContract := &core.TriggerSmartContract{
		OwnerAddress:    callerAddressBytes,
		ContractAddress: contractAddressBytes,
		Data:            encodeDataByte,
		CallValue:       0, //默认为0
		TokenId:         0, //默认为0
		CallTokenValue:  0, //默认为0
	}

	transactionExtension, err := client.Client.TriggerConstantContract(context.Background(), triggerContract)
	if err != nil {
		return [32]byte{}, fmt.Errorf("ccl TriggerConstantContract fail: %v", err)
	}

	// 解析结果
	result, err := ParseCallResult(transactionExtension, "getConfiguration", &abiCcl)
	if err != nil {
		return [32]byte{}, err
	}
	value, ok := result.([32]byte)
	if !ok {
		return [32]byte{}, fmt.Errorf("ccl parsecallresult error")
	}
	return value, nil
}
