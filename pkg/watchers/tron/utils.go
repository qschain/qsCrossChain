package tron

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"golang.org/x/crypto/sha3"
)

// 常量定义
const (
	// Tron地址前缀字节
	TronAddressPrefix = 0x41

	// Tron和以太坊地址长度（字节）
	AddressLength = 20

	// 校验和长度（字节）
	ChecksumLength = 4
)

// AddressToHex 将Tron地址转换为十六进制格式
func AddressToHex(tronAddress string) (string, error) {
	bytes, err := common.DecodeCheck(tronAddress)
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(bytes), nil
}

// HexToAddress 将十六进制地址转换为Tron地址
func HexToAddress(hexAddress string) (string, error) {
	if strings.HasPrefix(hexAddress, "0x") {
		hexAddress = hexAddress[2:]
	}
	bytes, err := hex.DecodeString(hexAddress)
	if err != nil {
		return "", err
	}

	return common.EncodeCheck(bytes), nil
}

// IsValidTronAddress 验证Tron地址是否有效
func IsValidTronAddress(address string) bool {
	_, err := common.DecodeCheck(address)
	return err == nil
}

// calculateChecksum 计算双SHA256校验和
func calculateChecksum(data []byte) []byte {
	// 第一次SHA256
	firstHash := sha256.Sum256(data)

	// 第二次SHA256
	secondHash := sha256.Sum256(firstHash[:])

	// 取前4字节作为校验和
	return secondHash[:ChecksumLength]
}

// verifyChecksum 验证Base58Check校验和
func verifyChecksum(data []byte) bool {
	if len(data) < ChecksumLength {
		return false
	}

	// 数据部分（不包括校验和）
	dataPart := data[:len(data)-ChecksumLength]

	// 计算数据部分的校验和
	computedChecksum := calculateChecksum(dataPart)

	// 比较校验和
	actualChecksum := data[len(data)-ChecksumLength:]

	for i := 0; i < ChecksumLength; i++ {
		if computedChecksum[i] != actualChecksum[i] {
			return false
		}
	}

	return true
}

// 修改后的函数，返回EIP-55校验和地址
func TronAddressToEth(tronAddrStr string) (string, error) {
	decoded, err := address.Base58ToAddress(tronAddrStr)
	if err != nil {
		return "", nil
	}
	if len(decoded) < 21 {
		return "", nil
	}

	addressBytes := decoded[1:21]
	hexAddress := common.BytesToHexString(addressBytes)

	// 生成EIP-55校验和地址
	checksumAddress := toChecksumAddress(hexAddress)

	return checksumAddress, nil
}

// 计算波场地址校验和
func calculateTronChecksum(data []byte) []byte {
	// 第一次SHA256
	hash1 := sha256.Sum256(data)
	// 第二次SHA256
	hash2 := sha256.Sum256(hash1[:])
	return hash2[:]
}
func EthAddressToTron(ethAddrStr string) (string, error) {
	TronAddressPrefix := byte(0x41)
	// 移除可能的0x前缀
	ethAddrStr = strings.TrimPrefix(ethAddrStr, "0x")

	// 验证以太坊地址长度
	if len(ethAddrStr) != 40 && len(ethAddrStr) != 42 {
		return "", errors.New("以太坊地址长度无效")
	}

	// 如果地址包含EIP-55校验和，转换为小写（只取20字节部分）
	if len(ethAddrStr) == 42 {
		ethAddrStr = ethAddrStr[2:]
	}

	// 解码为字节
	addressBytes, err := hex.DecodeString(ethAddrStr)
	if err != nil {
		return "", errors.New("十六进制解码失败")
	}

	if len(addressBytes) != 20 {
		return "", errors.New("地址字节长度无效")
	}

	// 构建波场地址：前缀 + 地址
	tronBytes := make([]byte, 0, 25)
	tronBytes = append(tronBytes, TronAddressPrefix)
	tronBytes = append(tronBytes, addressBytes...)

	// 计算双SHA256校验和
	checksum := calculateTronChecksum(tronBytes)
	tronBytes = append(tronBytes, checksum[:4]...)

	// Base58编码
	tronAddr := base58.Encode(tronBytes)

	return tronAddr, nil
}

// 转换为EIP-55校验和地址
func toChecksumAddress(address string) string {
	// 确保地址没有0x前缀且小写
	address = strings.ToLower(strings.TrimPrefix(address, "0x"))

	// 计算地址的Keccak256哈希
	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(address))
	hashBytes := hash.Sum(nil)
	hashHex := hex.EncodeToString(hashBytes)

	// 构建校验和地址
	var result strings.Builder
	result.WriteString("0x")

	// 遍历地址的每个字符
	for i, c := range address {
		// 检查哈希值对应位置的值
		hashChar := hashHex[i]
		// 如果哈希值在对应位置大于7，则将该字符大写
		if (hashChar >= '8' && hashChar <= '9') || (hashChar >= 'a' && hashChar <= 'f') {
			result.WriteString(strings.ToUpper(string(c)))
		} else {
			result.WriteRune(c)
		}
	}

	return result.String()
}

// TronToEthereum 将波场地址转换为以太坊地址
// 参数: tronAddress - 波场地址（以T开头）
// 返回: 以太坊地址（0x开头）和可能的错误
/*func TronToEthereum(tronAddress string) (string, error) {
	if tronAddress == "" {
		return "", errors.New("波场地址不能为空")
	}

	// 移除可能的空格
	tronAddress = strings.TrimSpace(tronAddress)

	// 验证波场地址格式
	if !strings.HasPrefix(tronAddress, "T") {
		return "", errors.New("无效的波场地址格式：应以T开头")
	}

	// 解码Base58Check编码的波场地址
	decoded := base58.Decode(tronAddress)
	if len(decoded) != 25 { // 20字节地址 + 1字节版本 + 4字节校验和
		return "", errors.New("无效的波场地址长度")
	}

	// 提取地址主体（去除版本字节和校验和）
	// decoded[0] 是版本字节 (0x41)
	// decoded[1:21] 是20字节的地址主体
	// decoded[21:] 是4字节的校验和
	addressBody := decoded[1:21]

	if len(addressBody) != AddressLength {
		return "", errors.New("地址主体长度不正确")
	}

	// 验证校验和
	if !verifyChecksum(decoded) {
		return "", errors.New("地址校验和验证失败")
	}

	// 转换为以太坊地址格式（0x + 40个十六进制字符）
	ethereumAddress := "0x" + hex.EncodeToString(addressBody)

	// 验证以太坊地址格式
	if !ethCommon.IsHexAddress(ethereumAddress) {
		return "", errors.New("转换后的地址不是有效的以太坊地址")
	}

	return ethereumAddress, nil
}*/

// 编码调用合约的方法
func EncodeFunctionCall(methodName string, parameters []interface{}, abi *abi.ABI) ([]byte, error) {
	// 获取方法
	method, exists := abi.Methods[methodName]
	if !exists {
		return nil, fmt.Errorf("方法 %s 不存在于合约ABI中", methodName)
	}
	// 编码参数
	encodedData, err := method.Inputs.Pack(parameters...)
	if err != nil {
		return nil, fmt.Errorf("编码参数失败: %v", err)
	}
	// 添加方法选择器
	methodID := method.ID
	fullData := append(methodID, encodedData...)

	return fullData, nil
}
func ParseCallResult(transactionExtension *api.TransactionExtention, methodName string, abi *abi.ABI) (interface{}, error) {
	if transactionExtension == nil {
		return nil, errors.New("交易扩展信息为空")
	}
	if transactionExtension.Result != nil && !transactionExtension.Result.Result {
		return nil, errors.New(string(transactionExtension.Result.Message))
	}
	// 获取常量结果
	if transactionExtension.ConstantResult == nil || len(transactionExtension.ConstantResult) == 0 {
		return nil, errors.New("无常量结果返回")
	}

	rawData := transactionExtension.ConstantResult[0]

	// 尝试解码数据
	decodedData, err := decodeFunctionResult(methodName, rawData, abi)
	if err != nil {
		return nil, err
	}
	return decodedData, nil
}

func decodeFunctionResult(methodName string, rawData []byte, abi *abi.ABI) (interface{}, error) {
	method, exists := abi.Methods[methodName]
	if !exists {
		return nil, fmt.Errorf("方法 %s 不存在于合约ABI中", methodName)
	}

	if len(method.Outputs) == 0 {
		return nil, nil
	}

	// 解码输出
	decodedValues, err := method.Outputs.Unpack(rawData)
	if err != nil {
		return nil, fmt.Errorf("解码输出失败: %v", err)
	}

	// 如果只有一个返回值，直接返回
	if len(decodedValues) == 1 {
		return decodedValues[0], nil
	}

	// 否则返回切片
	return decodedValues, nil
}
func ConvertTronLogToEthLog(tronLog *core.TransactionInfo_Log) (*types.Log, error) {
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
		Address: address,
		Topics:  topics,
		Data:    tronLog.Data,
	}, nil
}
