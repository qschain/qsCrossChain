package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/contract"
	"github.com/gogo/protobuf/proto"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client/transaction"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/keys"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"time"

	"github.com/btcsuite/btcutil/base58"
	"github.com/ethereum/go-ethereum/accounts/abi"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

const AbiABI = ""

// https://github.com/fbsobreira/gotron-sdk/blob/master/docs/examples.md
func transferTRX() {
	c := client.NewGrpcClient("localhost:50051")
	err := c.Start(client.GRPCInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer c.Stop()

	// 私钥（十六进制字符串，不带 0x 前缀）
	privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	privateKeyHex2 := "55fec466e1988cf6ef419162452b18ef639a9004677a2d8ff9f858a22a01bed8"
	//fromAddress := "TVmHbCwgiKfubGZsDA8q3gKh6kiJSAjYpM" // 对应私钥的地址
	//toAddress := "TQN9nFQq8FiyFdycYQ1Foqo6FKvj9G6qjK"
	funderKey1, _ := keys.GetPrivateKeyFromHex(privateKeyHex1)
	funderAddr1 := address.BTCECPrivkeyToAddress(funderKey1)
	funderKey2, _ := keys.GetPrivateKeyFromHex(privateKeyHex2)
	funderAddr2 := address.BTCECPrivkeyToAddress(funderKey2)

	tx, err := c.Transfer(funderAddr1.String(), funderAddr2.String(), 100_000_00) // 10 TRX = 10 * 1e6 sun
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := transaction.SignTransaction(tx.Transaction, funderKey1)
	if err != nil {
		log.Fatal(err)
	}

	result, err := c.Broadcast(signedTx)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !result.Result ||
		result.Code != api.Return_SUCCESS {
		fmt.Printf("Broadcast failed: (%d) %s", result.Code, result.Message)
	}

	fmt.Printf("Transaction ID: %x\n", tx.Txid)
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

// readBytecodeFromFile 从文件读取字节码
func readBytecodeFromFile(filename string) (string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return "", fmt.Errorf("字节码文件不存在: %s", filename)
	}

	// 读取文件内容
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	// 转换为字符串并清理
	bytecode := strings.TrimSpace(string(data))

	// 移除可能的0x前缀
	if strings.HasPrefix(bytecode, "0x") {
		bytecode = bytecode[2:]
	}

	// 移除所有空白字符（包括换行符）
	bytecode = strings.Join(strings.Fields(bytecode), "")

	log.Printf("   从文件 %s 读取字节码", filename)
	log.Printf("   原始长度: %d 字符", len(string(data)))
	log.Printf("   清理后长度: %d 字符", len(bytecode))

	return bytecode, nil
}

// 修改后的函数，返回EIP-55校验和地址
func tronAddressToEth(tronAddrStr string) {
	decoded, err := address.Base58ToAddress(tronAddrStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(decoded) < 21 {
		fmt.Println(err)
		return
	}

	addressBytes := decoded[1:21]
	hexAddress := common.BytesToHexString(addressBytes)

	// 生成EIP-55校验和地址
	checksumAddress := toChecksumAddress(hexAddress)

	fmt.Println(checksumAddress)
}

// 计算波场地址校验和
func calculateTronChecksum(data []byte) []byte {
	// 第一次SHA256
	hash1 := sha256.Sum256(data)
	// 第二次SHA256
	hash2 := sha256.Sum256(hash1[:])
	return hash2[:]
}
func ethAddressToTron(ethAddrStr string) {
	TronAddressPrefix := byte(0x41)
	// 移除可能的0x前缀
	ethAddrStr = strings.TrimPrefix(ethAddrStr, "0x")

	// 验证以太坊地址长度
	if len(ethAddrStr) != 40 && len(ethAddrStr) != 42 {
		fmt.Println("以太坊地址长度无效")
		return
	}

	// 如果地址包含EIP-55校验和，转换为小写（只取20字节部分）
	if len(ethAddrStr) == 42 {
		ethAddrStr = ethAddrStr[2:]
	}

	// 解码为字节
	addressBytes, err := hex.DecodeString(ethAddrStr)
	if err != nil {
		fmt.Println("十六进制解码失败")
		return
	}

	if len(addressBytes) != 20 {
		fmt.Println("地址字节长度无效")
		return
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

	fmt.Println(tronAddr)
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

func getTransactionById() {
	// 创建gRPC连接
	grpcClient := client.NewGrpcClient("localhost:50051")
	// ... 初始化连接
	err := grpcClient.Start(client.GRPCInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer grpcClient.Stop()
	tx, err := grpcClient.GetTransactionInfoByID("c68a19c6df917a09b41f0dad0fcc8adb35a530934abe882ffeb674c3842e6edf")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(tx)
	fmt.Println(tx.GetLog())

}
func CallConstantFunction(contractAddr string, methodStr string, encodeDataStr string, abiStr string) {
	c := client.NewGrpcClient("localhost:50051")
	err := c.Start(client.GRPCInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Stop()
	abiJson, err := abi.JSON(strings.NewReader(abiStr))
	if err != nil {
		fmt.Println(err)
		return
	}
	encodeDataByte, err := common.FromHex(encodeDataStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	contractAddressBytes, err := common.DecodeCheck(contractAddr)
	triggerContract := &core.TriggerSmartContract{
		OwnerAddress:    []byte{},
		ContractAddress: contractAddressBytes,
		Data:            encodeDataByte,
		CallValue:       0, //默认为0
		TokenId:         0, //默认为0
		CallTokenValue:  0, //默认为0
	}
	transactionExtension, err := c.Client.TriggerConstantContract(context.Background(), triggerContract)
	if err != nil {
		fmt.Println(err)
		return
	}
	result, err := decodeFunctionResult(methodStr, transactionExtension.ConstantResult[0], &abiJson)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("=========================")
	fmt.Println(result)
	value, ok := result.(ethCommon.Address)
	if !ok {
		fmt.Println("err")
		return
	}
	ethAddressToTron(value.String())
}

// 执行交易代码
func createAndSendTriggerTransaction(
	grpcClient *client.GrpcClient,
	privateKey *btcec.PrivateKey,
	fromAddr string,
	bytecode string,
	abiStr string,
	contractAddr string,
) (string, error) {
	fromTronAddr := address.HexToAddress(fromAddr)
	/*abi,err := contract.JSONtoABI(abiStr)
	if err != nil{
		fmt.Println(err)
		return "",err
	}*/
	bytecodeHex, _ := common.FromHex(bytecode)
	ctx := context.Background()
	decodedContractAddr, err := address.Base58ToAddress(contractAddr)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	contractTrigger := &core.TriggerSmartContract{
		OwnerAddress:    fromTronAddr.Bytes(),
		ContractAddress: decodedContractAddr,
		Data:            bytecodeHex,
		CallValue:       0,
		CallTokenValue:  0,
		TokenId:         0,
	}
	txExtension, err := grpcClient.Client.TriggerContract(ctx, contractTrigger)
	if err != nil {
		return "", fmt.Errorf("创建交易失败: %v", err)
	}
	/*tx, err := grpcClient.Client.TriggerContract(ctx, trigger)
	if err != nil {
		return "", fmt.Errorf("创建交易失败: %v", err)
	}*/
	txExtension.Transaction.RawData.FeeLimit = 1000_000_000
	tx := txExtension.Transaction

	signedTx, err := transaction.SignTransaction(tx, privateKey)
	if err != nil {
		return "", fmt.Errorf("签名交易失败: %v", err)
	}
	rawData := signedTx.GetRawData()
	rawDataBytes, err := proto.Marshal(rawData)
	if err != nil {
		return "", fmt.Errorf("获取id失败%v", err)
	}
	//txID := common.BytesToHexString(txExtension.Txid)
	hash := sha256.Sum256(rawDataBytes)
	txID := hex.EncodeToString(hash[:])
	fmt.Printf("   交易签名成功")

	// 广播交易
	result, err := grpcClient.Broadcast(signedTx)
	if err != nil {
		return "", fmt.Errorf("广播交易失败: %v", err)
	}
	fmt.Println(result)
	if !result.Result {
		errorMsg := string(result.Message)
		if result.Code != api.Return_SUCCESS {
			errorMsg = fmt.Sprintf("错误代码:%d,消息:%s", result.Code, string(result.Message))
		}
		// 常见错误处理
		if strings.Contains(errorMsg, "contract validate") {
			errorMsg += " (合约验证失败，请检查字节码和参数)"
		} else if strings.Contains(errorMsg, "bandwidth") {
			errorMsg += " (带宽不足，请质押TRX获取带宽或等待)"
		} else if strings.Contains(errorMsg, "energy") {
			errorMsg += " (能量不足，请质押TRX获取能量)"
		}
		return "", fmt.Errorf("交易广播失败: %s", result.Message)
	}

	// 返回交易ID
	fmt.Printf("   交易广播成功，交易ID: %s", txID)

	return txID, nil

}

// createAndSendDeployTransaction 创建并发送部署交易
func createAndSendDeployTransaction(
	grpcClient *client.GrpcClient,
	privateKey *btcec.PrivateKey,
	fromAddr string,
	bytecode string,
	abiStr string,
) (string, error) {
	// 转换为波场地址格式
	fromTronAddr := address.HexToAddress(fromAddr)
	abi, err := contract.JSONtoABI(abiStr)
	log.Printf("  执行合约交易...")
	log.Printf("   部署者: %s", fromTronAddr.String())
	log.Printf("   字节码长度: %d 字节", len(bytecode)/2)

	// 创建智能合约部署交易
	bytecodeHex, _ := common.FromHex(bytecode)
	/*trigger := &core.TriggerSmartContract{
		OwnerAddress:    fromTronAddr.Bytes(),
		ContractAddress: []byte{}, // 空，表示部署新合约
		CallValue:       0,        // 不发送 TRX
		Data:            bytecodeHex,
	}*/

	// 创建交易
	ctx := context.Background()
	contractCreate := &core.CreateSmartContract{
		OwnerAddress: fromTronAddr.Bytes(),
		NewContract: &core.SmartContract{
			OriginAddress:     fromTronAddr.Bytes(),
			ContractAddress:   []byte{},
			Abi:               abi,
			Bytecode:          bytecodeHex,
			CallValue:         0,
			OriginEnergyLimit: 5000000,
		},
		CallTokenValue: 0,
		TokenId:        0,
	}
	txExtension, err := grpcClient.Client.DeployContract(ctx, contractCreate)
	if err != nil {
		return "", fmt.Errorf("创建交易失败: %v", err)
	}
	/*tx, err := grpcClient.Client.TriggerContract(ctx, trigger)
	if err != nil {
		return "", fmt.Errorf("创建交易失败: %v", err)
	}*/
	txExtension.Transaction.RawData.FeeLimit = 1000_000_000
	tx := txExtension.Transaction

	signedTx, err := transaction.SignTransaction(tx, privateKey)
	if err != nil {
		return "", fmt.Errorf("签名交易失败: %v", err)
	}
	rawData := signedTx.GetRawData()
	rawDataBytes, err := proto.Marshal(rawData)
	if err != nil {
		return "", fmt.Errorf("获取id失败%v", err)
	}
	//txID := common.BytesToHexString(txExtension.Txid)
	hash := sha256.Sum256(rawDataBytes)
	txID := hex.EncodeToString(hash[:])
	fmt.Printf("   交易签名成功")

	// 广播交易
	result, err := grpcClient.Broadcast(signedTx)
	if err != nil {
		return "", fmt.Errorf("广播交易失败: %v", err)
	}
	fmt.Println(result)
	if !result.Result {
		errorMsg := string(result.Message)
		if result.Code != api.Return_SUCCESS {
			errorMsg = fmt.Sprintf("错误代码:%d,消息:%s", result.Code, string(result.Message))
		}
		// 常见错误处理
		if strings.Contains(errorMsg, "contract validate") {
			errorMsg += " (合约验证失败，请检查字节码和参数)"
		} else if strings.Contains(errorMsg, "bandwidth") {
			errorMsg += " (带宽不足，请质押TRX获取带宽或等待)"
		} else if strings.Contains(errorMsg, "energy") {
			errorMsg += " (能量不足，请质押TRX获取能量)"
		}
		return "", fmt.Errorf("交易广播失败: %s", result.Message)
	}

	// 返回交易ID
	fmt.Printf("   交易广播成功，交易ID: %s", txID)

	return txID, nil
}

func createAndSendDeployTransactiondeploysol(
	grpcClient *client.GrpcClient,
	privateKey *btcec.PrivateKey,
	fromAddr string,
	bytecode string,
	abiStr string,
) (string, error) {
	// 转换为波场地址格式
	fromTronAddr := address.HexToAddress(fromAddr)
	abi, err := contract.JSONtoABI(abiStr)
	log.Printf("   创建合约部署交易...")
	log.Printf("   部署者: %s", fromTronAddr.String())
	log.Printf("   字节码长度: %d 字节", len(bytecode)/2)

	// 创建智能合约部署交易
	bytecodeHex, _ := common.FromHex(bytecode)
	/*trigger := &core.TriggerSmartContract{
		OwnerAddress:    fromTronAddr.Bytes(),
		ContractAddress: []byte{}, // 空，表示部署新合约
		CallValue:       0,        // 不发送 TRX
		Data:            bytecodeHex,
	}*/

	// 创建交易
	ctx := context.Background()
	contractCreate := &core.CreateSmartContract{
		OwnerAddress: fromTronAddr.Bytes(),
		NewContract: &core.SmartContract{
			OriginAddress:     fromTronAddr.Bytes(),
			ContractAddress:   []byte{},
			Abi:               abi,
			Bytecode:          bytecodeHex,
			CallValue:         0,
			OriginEnergyLimit: 5000000,
		},
		CallTokenValue: 0,
		TokenId:        0,
	}
	txExtension, err := grpcClient.Client.DeployContract(ctx, contractCreate)
	if err != nil {
		return "", fmt.Errorf("创建交易失败: %v", err)
	}
	/*tx, err := grpcClient.Client.TriggerContract(ctx, trigger)
	if err != nil {
		return "", fmt.Errorf("创建交易失败: %v", err)
	}*/
	txExtension.Transaction.RawData.FeeLimit = 1000_000_000
	tx := txExtension.Transaction

	signedTx, err := transaction.SignTransaction(tx, privateKey)
	if err != nil {
		return "", fmt.Errorf("签名交易失败: %v", err)
	}
	rawData := signedTx.GetRawData()
	rawDataBytes, err := proto.Marshal(rawData)
	if err != nil {
		return "", fmt.Errorf("获取id失败%v", err)
	}
	//txID := common.BytesToHexString(txExtension.Txid)
	hash := sha256.Sum256(rawDataBytes)
	txID := hex.EncodeToString(hash[:])
	fmt.Printf("   交易签名成功")

	// 广播交易
	result, err := grpcClient.Broadcast(signedTx)
	if err != nil {
		return "", fmt.Errorf("广播交易失败: %v", err)
	}
	fmt.Println(result)
	if !result.Result {
		errorMsg := string(result.Message)
		if result.Code != api.Return_SUCCESS {
			errorMsg = fmt.Sprintf("错误代码:%d,消息:%s", result.Code, string(result.Message))
		}
		// 常见错误处理
		if strings.Contains(errorMsg, "contract validate") {
			errorMsg += " (合约验证失败，请检查字节码和参数)"
		} else if strings.Contains(errorMsg, "bandwidth") {
			errorMsg += " (带宽不足，请质押TRX获取带宽或等待)"
		} else if strings.Contains(errorMsg, "energy") {
			errorMsg += " (能量不足，请质押TRX获取能量)"
		}
		return "", fmt.Errorf("交易广播失败: %s", result.Message)
	}

	// 返回交易ID
	fmt.Printf("   交易广播成功，交易ID: %s", txID)

	return txID, nil
}

// waitForDeployment 等待部署完成
func waitForDeployment(grpcClient *client.GrpcClient, txID string) (string, int64, error) {
	maxRetries := 30
	retryInterval := 3 * time.Second

	fmt.Printf("   开始等待交易确认，最多尝试 %d 次，每次间隔 %v", maxRetries, retryInterval)

	for i := 0; i < maxRetries; i++ {
		fmt.Printf("   检查交易状态 (%d/%d)...", i+1, maxRetries)

		// 获取交易信息
		txInfo, err := grpcClient.GetTransactionInfoByID(txID)
		if err != nil {
			fmt.Printf("     获取交易信息失败: %v", err)
			time.Sleep(retryInterval)
			continue
		}

		if txInfo == nil {
			fmt.Println("     交易信息为空")
			time.Sleep(retryInterval)
			continue
		}

		// 检查交易状态
		if txInfo.Receipt != nil {
			if txInfo.Receipt.Result == core.Transaction_Result_SUCCESS {
				// 获取合约地址
				if txInfo.ContractAddress != nil && len(txInfo.ContractAddress) > 0 {
					contractAddr := address.Address(txInfo.ContractAddress)
					fmt.Printf("      交易确认成功!")
					fmt.Printf("      合约地址: %s", contractAddr.String())
					fmt.Printf("      区块号: %d", txInfo.BlockNumber)
					fmt.Printf("      燃料消耗: %d", txInfo.Receipt.EnergyUsage)
					return contractAddr.String(), txInfo.BlockNumber, nil
				} else {
					fmt.Println("     交易成功但未生成合约地址")
				}
			} else if txInfo.Receipt.Result != core.Transaction_Result_SUCCESS {
				fmt.Printf("     交易失败，原因: %v", txInfo.Receipt.Result)
				if txInfo.ResMessage != nil && len(txInfo.ResMessage) > 0 {
					fmt.Printf("     错误信息: %s", string(txInfo.ResMessage))
				}
				return "", 0, fmt.Errorf("交易失败: %v", txInfo.Receipt.Result)
			}
		} else {
			fmt.Println("     交易收据为空，交易可能尚未打包")
		}

		time.Sleep(retryInterval)
	}

	return "", 0, fmt.Errorf("等待交易确认超时")
}

func deployImplementContract() {
	c := client.NewGrpcClient("localhost:50051")
	err := c.Start(client.GRPCInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Stop()
	privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	bytecodeFile := "./abi_bin/implement.bin"
	abiFile := "./abi_bin/implement.abi"
	//privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	//privateKeyHex2 := "55fec466e1988cf6ef419162452b18ef639a9004677a2d8ff9f858a22a01bed8"
	//fromAddress := "TVmHbCwgiKfubGZsDA8q3gKh6kiJSAjYpM" // 对应私钥的地址
	//toAddress := "TQN9nFQq8FiyFdycYQ1Foqo6FKvj9G6qjK"
	privateKey, _ := keys.GetPrivateKeyFromHex(privateKeyHex1)
	fromAddress := address.BTCECPrivkeyToAddress(privateKey)
	contractBytecode, err := readBytecodeFromFile(bytecodeFile)
	if err != nil {
		fmt.Printf("读取字节码文件失败: %v", err)
		return
	}
	contractAbi, err := readBytecodeFromFile(abiFile)
	if err != nil {
		fmt.Printf("读取abi失败:%v", err)
		return
	}
	// 编码构造函数参数 (uint256)
	/*encodedArgs, err := encodeConstructorArgs(initialValue)
	if err != nil {
		fmt.Printf("编码构造函数参数失败: %v", err)
		return
	}
	fmt.Println("==========encodedArgs=========================")
	fmt.Println(encodedArgs)
	// 合并字节码和构造函数参数
	finalBytecode := contractBytecode + encodedArgs*/
	finalBytecode := contractBytecode

	txID, err := createAndSendDeployTransaction(
		c,
		privateKey,
		fromAddress.Hex(),
		finalBytecode,
		contractAbi,
	)
	if err != nil {
		fmt.Printf("创建交易失败: %v", err)
		return
	}
	contractAddress, blockNumber, err := waitForDeployment(c, txID)
	if err != nil {
		fmt.Printf("部署失败: %v", err)
		return
	}
	// ==================== 输出结果 ====================
	fmt.Println("\n🎉 合约部署成功!")
	fmt.Println("=====================================")
	fmt.Printf("合约地址: %s", contractAddress)
	fmt.Printf("交易ID: %s", txID)
	fmt.Printf("区块号: %d", blockNumber)
}
func deployCclContract() {
	c := client.NewGrpcClient("localhost:50051")
	err := c.Start(client.GRPCInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Stop()
	privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	bytecodeFile := "./abi_bin/ccl.bin"
	abiFile := "./abi_bin/ccl.abi"
	//privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	//privateKeyHex2 := "55fec466e1988cf6ef419162452b18ef639a9004677a2d8ff9f858a22a01bed8"
	//fromAddress := "TVmHbCwgiKfubGZsDA8q3gKh6kiJSAjYpM" // 对应私钥的地址
	//toAddress := "TQN9nFQq8FiyFdycYQ1Foqo6FKvj9G6qjK"
	privateKey, _ := keys.GetPrivateKeyFromHex(privateKeyHex1)
	fromAddress := address.BTCECPrivkeyToAddress(privateKey)
	contractBytecode, err := readBytecodeFromFile(bytecodeFile)
	if err != nil {
		fmt.Printf("读取字节码文件失败: %v", err)
		return
	}
	contractAbi, err := readBytecodeFromFile(abiFile)
	if err != nil {
		fmt.Printf("读取abi失败:%v", err)
		return
	}
	// 编码构造函数参数 (uint256)
	/*encodedArgs, err := encodeConstructorArgs(initialValue)
	if err != nil {
		fmt.Printf("编码构造函数参数失败: %v", err)
		return
	}
	fmt.Println("==========encodedArgs=========================")
	fmt.Println(encodedArgs)
	// 合并字节码和构造函数参数
	finalBytecode := contractBytecode + encodedArgs*/
	finalBytecode := contractBytecode

	txID, err := createAndSendDeployTransaction(
		c,
		privateKey,
		fromAddress.Hex(),
		finalBytecode,
		contractAbi,
	)
	if err != nil {
		fmt.Printf("创建交易失败: %v", err)
		return
	}
	contractAddress, blockNumber, err := waitForDeployment(c, txID)
	if err != nil {
		fmt.Printf("部署失败: %v", err)
		return
	}
	// ==================== 输出结果 ====================
	fmt.Println("\n🎉 合约部署成功!")
	fmt.Println("=====================================")
	fmt.Printf("合约地址: %s", contractAddress)
	fmt.Printf("交易ID: %s", txID)
	fmt.Printf("区块号: %d", blockNumber)
}

func deploySetupContract() {
	c := client.NewGrpcClient("localhost:50051")
	err := c.Start(client.GRPCInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Stop()
	privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	bytecodeFile := "./abi_bin/setup.bin"
	abiFile := "./abi_bin/setup.abi"
	//initialValue := int64(42)
	/*privateKey, fromAddress, err := getAccountFromPrivateKey(privateKeyHex1)
	if err != nil {
		fmt.Printf("账户初始化失败: %v", err)
		return
	}*/
	//privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	//privateKeyHex2 := "55fec466e1988cf6ef419162452b18ef639a9004677a2d8ff9f858a22a01bed8"
	//fromAddress := "TVmHbCwgiKfubGZsDA8q3gKh6kiJSAjYpM" // 对应私钥的地址
	//toAddress := "TQN9nFQq8FiyFdycYQ1Foqo6FKvj9G6qjK"
	privateKey, _ := keys.GetPrivateKeyFromHex(privateKeyHex1)
	fromAddress := address.BTCECPrivkeyToAddress(privateKey)
	contractBytecode, err := readBytecodeFromFile(bytecodeFile)
	if err != nil {
		fmt.Printf("读取字节码文件失败: %v", err)
		return
	}
	contractAbi, err := readBytecodeFromFile(abiFile)
	if err != nil {
		fmt.Printf("读取abi失败:%v", err)
		return
	}
	// 编码构造函数参数 (uint256)
	/*encodedArgs, err := encodeConstructorArgs(initialValue)
	if err != nil {
		fmt.Printf("编码构造函数参数失败: %v", err)
		return
	}
	fmt.Println("==========encodedArgs=========================")
	fmt.Println(encodedArgs)
	// 合并字节码和构造函数参数
	finalBytecode := contractBytecode + encodedArgs*/
	finalBytecode := contractBytecode

	txID, err := createAndSendDeployTransaction(
		c,
		privateKey,
		fromAddress.Hex(),
		finalBytecode,
		contractAbi,
	)
	if err != nil {
		fmt.Printf("创建交易失败: %v", err)
		return
	}
	contractAddress, blockNumber, err := waitForDeployment(c, txID)
	if err != nil {
		fmt.Printf("部署失败: %v", err)
		return
	}
	// ==================== 输出结果 ====================
	fmt.Println("\n🎉 合约部署成功!")
	fmt.Println("=====================================")
	fmt.Printf("合约地址: %s", contractAddress)
	fmt.Printf("交易ID: %s", txID)
	fmt.Printf("区块号: %d", blockNumber)
}

func triggerContract(deployContractAddr string) {
	c := client.NewGrpcClient("localhost:50051")
	err := c.Start(client.GRPCInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Stop()
	privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	//bytecodeFile := "./abi_bin/deploy.bin"
	abiFile := "./abi_bin/deploy.abi"
	//initialValue := int64(42)
	/*privateKey, fromAddress, err := getAccountFromPrivateKey(privateKeyHex1)
	if err != nil {
		fmt.Printf("账户初始化失败: %v", err)
		return
	}*/
	//privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	//privateKeyHex2 := "55fec466e1988cf6ef419162452b18ef639a9004677a2d8ff9f858a22a01bed8"
	//fromAddress := "TVmHbCwgiKfubGZsDA8q3gKh6kiJSAjYpM" // 对应私钥的地址
	//toAddress := "TQN9nFQq8FiyFdycYQ1Foqo6FKvj9G6qjK"
	privateKey, _ := keys.GetPrivateKeyFromHex(privateKeyHex1)
	fromAddress := address.BTCECPrivkeyToAddress(privateKey)
	contractAddress := deployContractAddr

	/*contractBytecode, err := readBytecodeFromFile(bytecodeFile)
	if err != nil {
		fmt.Printf("读取字节码文件失败: %v", err)
		return
	}*/
	contractAbi, err := readBytecodeFromFile(abiFile)
	if err != nil {
		fmt.Printf("读取abi失败:%v", err)
		return
	}
	// 编码构造函数参数 (uint256)
	/*encodedArgs, err := encodeConstructorArgs(initialValue)
	if err != nil {
		fmt.Printf("编码构造函数参数失败: %v", err)
		return
	}
	fmt.Println("==========encodedArgs=========================")
	fmt.Println(encodedArgs)
	// 合并字节码和构造函数参数
	finalBytecode := contractBytecode + encodedArgs*/
	finalBytecode := "0x78e7d1fb00000000000000000000000000000000000000000000000000000000000000a000000000000000000000000000000000000000000000000000000000000000de0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000005390000000000000000000000000000000000000000000000000000000000000001000000000000000000000000d2a3b9bb6e710ef61396c8baf8adc4655c3429f4"
	txID, err := createAndSendTriggerTransaction(
		c,
		privateKey,
		fromAddress.Hex(),
		finalBytecode,
		contractAbi,
		contractAddress,
	)
	if err != nil {
		fmt.Printf("创建交易失败: %v", err)
		return
	}
	contractAddress, blockNumber, err := waitForDeployment(c, txID)
	if err != nil {
		fmt.Printf("部署失败: %v", err)
		return
	}
	// ==================== 输出结果 ====================
	fmt.Println("\n🎉 合约执行成功!")
	fmt.Println("=====================================")
	fmt.Printf("合约地址: %s", contractAddress)
	fmt.Printf("交易ID: %s", txID)
	fmt.Printf("区块号: %d", blockNumber)
}
func triggerContractVAAVerfiy(tokenContractAddr string) {
	c := client.NewGrpcClient("localhost:50051")
	err := c.Start(client.GRPCInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Stop()
	privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	//bytecodeFile := "./abi_bin/deploy.bin"
	abiFile := "./abi_bin/token.abi"
	//initialValue := int64(42)
	/*privateKey, fromAddress, err := getAccountFromPrivateKey(privateKeyHex1)
	if err != nil {
		fmt.Printf("账户初始化失败: %v", err)
		return
	}*/
	//privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	//privateKeyHex2 := "55fec466e1988cf6ef419162452b18ef639a9004677a2d8ff9f858a22a01bed8"
	//fromAddress := "TVmHbCwgiKfubGZsDA8q3gKh6kiJSAjYpM" // 对应私钥的地址
	//toAddress := "TQN9nFQq8FiyFdycYQ1Foqo6FKvj9G6qjK"
	privateKey, _ := keys.GetPrivateKeyFromHex(privateKeyHex1)
	fromAddress := address.BTCECPrivkeyToAddress(privateKey)
	contractAddress := tokenContractAddr

	/*contractBytecode, err := readBytecodeFromFile(bytecodeFile)
	if err != nil {
		fmt.Printf("读取字节码文件失败: %v", err)
		return
	}*/
	contractAbi, err := readBytecodeFromFile(abiFile)
	if err != nil {
		fmt.Printf("读取abi失败:%v", err)
		return
	}
	// 编码构造函数参数 (uint256)
	/*encodedArgs, err := encodeConstructorArgs(initialValue)
	if err != nil {
		fmt.Printf("编码构造函数参数失败: %v", err)
		return
	}
	fmt.Println("==========encodedArgs=========================")
	fmt.Println(encodedArgs)
	// 合并字节码和构造函数参数
	finalBytecode := contractBytecode + encodedArgs*/
	finalBytecode := "0xb02275b10000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000015b01000000000100420b902a931db97f614b379d7a811a3c37df9ca4a96b072c0b5ab2c4fcb7e6e01e3c47342a2ed4786a855d1aab2d0924b56f774c886569e1e4848b8ec69caa7b006943a04e000000030002000000000000000000000000d73f34428098b44a589f13ad15a0e3d2efe92dbd0000000000000000c8000000000000000000000000d85403039f10faa0ad58a1caeb6137eb3f04096760f91c4983cd584ea4c48485b610cc900d867e1f806ae08ae922a90379466d3500000000000000000000000019e583b06050387337824d187b00892d6842664300000000000000000000000000000000000000000000000000000000000000420000000000000000000000000000000000000000000000000000000000000002000000000000000000000000d73f34428098b44a589f13ad15a0e3d2efe92dbd000000000000000000000000000000000000000000000000000000006943a04e0000000000"
	txID, err := createAndSendTriggerTransaction(
		c,
		privateKey,
		fromAddress.Hex(),
		finalBytecode,
		contractAbi,
		contractAddress,
	)
	if err != nil {
		fmt.Printf("创建交易失败: %v", err)
		return
	}
	contractAddress, blockNumber, err := waitForDeployment(c, txID)
	if err != nil {
		fmt.Printf("部署失败: %v", err)
		return
	}
	// ==================== 输出结果 ====================
	fmt.Println("\n合约执行成功!")
	fmt.Println("=====================================")
	fmt.Printf("合约地址: %s", contractAddress)
	fmt.Printf("交易ID: %s", txID)
	fmt.Printf("区块号: %d", blockNumber)
}
func triggerContractSendCrossChain(tokenContractAddr string) {
	c := client.NewGrpcClient("localhost:50051")
	err := c.Start(client.GRPCInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Stop()
	privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	//bytecodeFile := "./abi_bin/deploy.bin"
	abiFile := "./abi_bin/token.abi"
	//initialValue := int64(42)
	/*privateKey, fromAddress, err := getAccountFromPrivateKey(privateKeyHex1)
	if err != nil {
		fmt.Printf("账户初始化失败: %v", err)
		return
	}*/
	//privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	//privateKeyHex2 := "55fec466e1988cf6ef419162452b18ef639a9004677a2d8ff9f858a22a01bed8"
	//fromAddress := "TVmHbCwgiKfubGZsDA8q3gKh6kiJSAjYpM" // 对应私钥的地址
	//toAddress := "TQN9nFQq8FiyFdycYQ1Foqo6FKvj9G6qjK"
	privateKey, _ := keys.GetPrivateKeyFromHex(privateKeyHex1)
	fromAddress := address.BTCECPrivkeyToAddress(privateKey)
	contractAddress := tokenContractAddr

	/*contractBytecode, err := readBytecodeFromFile(bytecodeFile)
	if err != nil {
		fmt.Printf("读取字节码文件失败: %v", err)
		return
	}*/
	contractAbi, err := readBytecodeFromFile(abiFile)
	if err != nil {
		fmt.Printf("读取abi失败:%v", err)
		return
	}
	// 编码构造函数参数 (uint256)
	/*encodedArgs, err := encodeConstructorArgs(initialValue)
	if err != nil {
		fmt.Printf("编码构造函数参数失败: %v", err)
		return
	}
	fmt.Println("==========encodedArgs=========================")
	fmt.Println(encodedArgs)
	// 合并字节码和构造函数参数
	finalBytecode := contractBytecode + encodedArgs*/
	//finalBytecode := "0x9569327b000000000000000000000000000000000000000000000000000000000000000200000000000000000000000013cf0d18bcb898efd4e85ae2bb65a443ab86023c00000000000000000000000019e583b06050387337824d187b00892d6842664300000000000000000000000000000000000000000000000000000000000000610000000000000000000000000000000000000000000000000000000000000001"
	finalBytecode := "0x9569327b0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000d73f34428098b44a589f13ad15a0e3d2efe92dbd00000000000000000000000019e583b06050387337824d187b00892d68426643000000000000000000000000000000000000000000000000000000000000004d0000000000000000000000000000000000000000000000000000000000000002"
	txID, err := createAndSendTriggerTransaction(
		c,
		privateKey,
		fromAddress.Hex(),
		finalBytecode,
		contractAbi,
		contractAddress,
	)
	if err != nil {
		fmt.Printf("创建交易失败: %v", err)
		return
	}
	contractAddress, blockNumber, err := waitForDeployment(c, txID)
	if err != nil {
		fmt.Printf("部署失败: %v", err)
		return
	}
	// ==================== 输出结果 ====================
	fmt.Println("\n合约执行成功!")
	fmt.Println("=====================================")
	fmt.Printf("合约地址: %s", contractAddress)
	fmt.Printf("交易ID: %s", txID)
	fmt.Printf("区块号: %d", blockNumber)
}

func deployContractdeploy() {
	c := client.NewGrpcClient("localhost:50051")
	err := c.Start(client.GRPCInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Stop()
	privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	bytecodeFile := "./abi_bin/deploy.bin"
	abiFile := "./abi_bin/deploy.abi"
	/*implementAddrStr := "TJev6stQnr7PY5kBRqWyCahQKdFHeocYPa"
	setupAddrStr := "TXEPJ7hSJ8VaZEwkw3jyGG2qG2sS6USFhT"
	implementAddr, err := address.Base58ToAddress(implementAddrStr)
	fmt.Println(implementAddr)
	setupAddr, err := address.Base58ToAddress(setupAddrStr)
	fmt.Println(setupAddr)*/
	//initialValue := int64(42)
	/*privateKey, fromAddress, err := getAccountFromPrivateKey(privateKeyHex1)
	if err != nil {
		fmt.Printf("账户初始化失败: %v", err)
		return
	}*/
	//privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	//privateKeyHex2 := "55fec466e1988cf6ef419162452b18ef639a9004677a2d8ff9f858a22a01bed8"
	//fromAddress := "TVmHbCwgiKfubGZsDA8q3gKh6kiJSAjYpM" // 对应私钥的地址
	//toAddress := "TQN9nFQq8FiyFdycYQ1Foqo6FKvj9G6qjK"
	privateKey, _ := keys.GetPrivateKeyFromHex(privateKeyHex1)
	fromAddress := address.BTCECPrivkeyToAddress(privateKey)
	contractBytecode, err := readBytecodeFromFile(bytecodeFile)
	if err != nil {
		fmt.Printf("读取字节码文件失败: %v", err)
		return
	}
	contractAbi, err := readBytecodeFromFile(abiFile)
	if err != nil {
		fmt.Printf("读取abi失败:%v", err)
		return
	}
	// 编码构造函数参数 (uint256)
	//encodedArgs, err := encodeConstructorArgs(initialValue)
	if err != nil {
		fmt.Printf("编码构造函数参数失败: %v", err)
		return
	}
	fmt.Println("==========encodedArgs=========================")
	//fmt.Println(encodedArgs)
	// 合并字节码和构造函数参数
	/*finalBytecode := contractBytecode + encodedArgs*/
	finalBytecode := contractBytecode

	txID, err := createAndSendDeployTransactiondeploysol(
		c,
		privateKey,
		fromAddress.Hex(),
		finalBytecode,
		contractAbi,
	)
	if err != nil {
		fmt.Printf("创建交易失败: %v", err)
		return
	}
	contractAddress, blockNumber, err := waitForDeployment(c, txID)
	if err != nil {
		fmt.Printf("部署失败: %v", err)
		return
	}
	// ==================== 输出结果 ====================
	fmt.Println("\n🎉 合约部署成功!")
	fmt.Println("=====================================")
	fmt.Printf("合约地址: %s", contractAddress)
	fmt.Printf("交易ID: %s", txID)
	fmt.Printf("区块号: %d", blockNumber)
}

func deployContractToken() {
	c := client.NewGrpcClient("localhost:50051")
	err := c.Start(client.GRPCInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Stop()
	privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	//bytecodeFile := "./abi_bin/token.bin"
	abiFile := "./abi_bin/token.abi"
	/*implementAddrStr := "TJev6stQnr7PY5kBRqWyCahQKdFHeocYPa"
	setupAddrStr := "TXEPJ7hSJ8VaZEwkw3jyGG2qG2sS6USFhT"
	implementAddr, err := address.Base58ToAddress(implementAddrStr)
	fmt.Println(implementAddr)
	setupAddr, err := address.Base58ToAddress(setupAddrStr)
	fmt.Println(setupAddr)*/
	//initialValue := int64(42)
	/*privateKey, fromAddress, err := getAccountFromPrivateKey(privateKeyHex1)
	if err != nil {
		fmt.Printf("账户初始化失败: %v", err)
		return
	}*/
	//privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	//privateKeyHex2 := "55fec466e1988cf6ef419162452b18ef639a9004677a2d8ff9f858a22a01bed8"
	//fromAddress := "TVmHbCwgiKfubGZsDA8q3gKh6kiJSAjYpM" // 对应私钥的地址
	//toAddress := "TQN9nFQq8FiyFdycYQ1Foqo6FKvj9G6qjK"
	privateKey, _ := keys.GetPrivateKeyFromHex(privateKeyHex1)
	fromAddress := address.BTCECPrivkeyToAddress(privateKey)
	/*contractBytecode, err := readBytecodeFromFile(bytecodeFile)
	if err != nil {
		fmt.Printf("读取字节码文件失败: %v", err)
		return
	}*/
	contractAbi, err := readBytecodeFromFile(abiFile)
	if err != nil {
		fmt.Printf("读取abi失败:%v", err)
		return
	}
	// 编码构造函数参数 (uint256)
	/*encodedArgs, err := encodeConstructorArgs(initialValue)
	if err != nil {
		fmt.Printf("编码构造函数参数失败: %v", err)
		return
	}
	*/
	fmt.Println("==========encodedArgs=========================")
	//fmt.Println(encodedArgs)
	// 合并字节码和构造函数参数
	/*finalBytecode := contractBytecode + encodedArgs*/
	//记得修改wormhole的地址
	//constructorParams := "0x60c060405234801562000010575f80fd5b50d380156200001d575f80fd5b50d280156200002a575f80fd5b5060405162003b9138038062003b918339818101604052810190620000509190620006bd565b8282816003908162000063919062000998565b50806004908162000075919062000998565b5050505f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603620000e9576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401620000e09062000ada565b60405180910390fd5b8373ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff168152505060805173ffffffffffffffffffffffffffffffffffffffff16639a8a05926040518163ffffffff1660e01b8152600401602060405180830381865afa15801562000169573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906200018f919062000b36565b61ffff1660a08161ffff16815250505f811115620001ba57620001b93382620001c460201b60201c565b5b5050505062000c60565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160362000237575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016200022e919062000b77565b60405180910390fd5b6200024a5f83836200024e60201b60201c565b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603620002a2578060025f82825462000295919062000bbf565b9250508190555062000373565b5f805f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050818110156200032e578381836040517fe450d38c000000000000000000000000000000000000000000000000000000008152600401620003259392919062000c0a565b60405180910390fd5b8181035f808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603620003bc578060025f828254039250508190555062000406565b805f808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef8360405162000465919062000c45565b60405180910390a3505050565b5f604051905090565b5f80fd5b5f80fd5b5f74ffffffffffffffffffffffffffffffffffffffffff82169050919050565b620004ae8162000483565b8114620004b9575f80fd5b50565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f620004e782620004bc565b9050919050565b5f81519050620004fe81620004a3565b6200050981620004db565b905092915050565b5f80fd5b5f80fd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b620005618262000519565b810181811067ffffffffffffffff8211171562000583576200058262000529565b5b80604052505050565b5f6200059762000472565b9050620005a5828262000556565b919050565b5f67ffffffffffffffff821115620005c757620005c662000529565b5b620005d28262000519565b9050602081019050919050565b5f5b83811015620005fe578082015181840152602081019050620005e1565b5f8484015250505050565b5f6200061f6200061984620005aa565b6200058c565b9050828152602081018484840111156200063e576200063d62000515565b5b6200064b848285620005df565b509392505050565b5f82601f8301126200066a576200066962000511565b5b81516200067c84826020860162000609565b91505092915050565b5f819050919050565b620006998162000685565b8114620006a4575f80fd5b50565b5f81519050620006b7816200068e565b92915050565b5f805f8060808587031215620006d857620006d76200047b565b5b5f620006e787828801620004ee565b945050602085015167ffffffffffffffff8111156200070b576200070a6200047f565b5b620007198782880162000653565b935050604085015167ffffffffffffffff8111156200073d576200073c6200047f565b5b6200074b8782880162000653565b92505060606200075e87828801620006a7565b91505092959194509250565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680620007b957607f821691505b602082108103620007cf57620007ce62000774565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f60088302620008337fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82620007f6565b6200083f8683620007f6565b95508019841693508086168417925050509392505050565b5f819050919050565b5f620008806200087a620008748462000685565b62000857565b62000685565b9050919050565b5f819050919050565b6200089b8362000860565b620008b3620008aa8262000887565b84845462000802565b825550505050565b5f90565b620008c9620008bb565b620008d681848462000890565b505050565b5b81811015620008fd57620008f15f82620008bf565b600181019050620008dc565b5050565b601f8211156200094c576200091681620007d5565b6200092184620007e7565b8101602085101562000931578190505b620009496200094085620007e7565b830182620008db565b50505b505050565b5f82821c905092915050565b5f6200096e5f198460080262000951565b1980831691505092915050565b5f6200098883836200095d565b9150826002028217905092915050565b620009a3826200076a565b67ffffffffffffffff811115620009bf57620009be62000529565b5b620009cb8254620007a1565b620009d882828562000901565b5f60209050601f83116001811462000a0e575f8415620009f9578287015190505b62000a0585826200097b565b86555062000a74565b601f19841662000a1e86620007d5565b5f5b8281101562000a475784890151825560018201915060208501945060208101905062000a20565b8683101562000a67578489015162000a63601f8916826200095d565b8355505b6001600288020188555050505b505050505050565b5f82825260208201905092915050565b7f496e76616c696420776f726d686f6c65206164647265737300000000000000005f82015250565b5f62000ac260188362000a7c565b915062000acf8262000a8c565b602082019050919050565b5f6020820190508181035f83015262000af38162000ab4565b9050919050565b5f61ffff82169050919050565b62000b128162000afa565b811462000b1d575f80fd5b50565b5f8151905062000b308162000b07565b92915050565b5f6020828403121562000b4e5762000b4d6200047b565b5b5f62000b5d8482850162000b20565b91505092915050565b62000b7181620004db565b82525050565b5f60208201905062000b8c5f83018462000b66565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f62000bcb8262000685565b915062000bd88362000685565b925082820190508082111562000bf35762000bf262000b92565b5b92915050565b62000c048162000685565b82525050565b5f60608201905062000c1f5f83018662000b66565b62000c2e602083018562000bf9565b62000c3d604083018462000bf9565b949350505050565b5f60208201905062000c5a5f83018462000bf9565b92915050565b60805160a051612ed062000cc15f395f8181610a5d01528181610c810152610ece01525f8181610837015281816108c9015281816108ef0152818161097c01528181610b7801528181610cd301528181610f18015261121e0152612ed05ff3fe608060405260043610610122575f3560e01c8063889be5331161009f578063a9059cbb11610063578063a9059cbb1461053e578063b02275b114610592578063d294cd8e146105e6578063dcb2ee3c1461063b578063dd62ed3e1461068f57610129565b8063889be533146103f35780639569327b1461043657806395d89b41146104665780639886cdcd146104a85780639a8a0592146104fc57610129565b806340c10f19116100e657806340c10f191461029b57806342966c68146102db57806370a082311461031b5780637f18dc851461036f57806384acd1bb146103b157610129565b806306fdde031461012d578063095ea7b31461016f57806318160ddd146101c357806323b872dd14610205578063313ce5671461025957610129565b3661012957005b5f80fd5b348015610138575f80fd5b50d38015610144575f80fd5b50d28015610150575f80fd5b506101596106e3565b6040516101669190611aa0565b60405180910390f35b34801561017a575f80fd5b50d38015610186575f80fd5b50d28015610192575f80fd5b506101ad60048036038101906101a89190611b89565b610773565b6040516101ba9190611be1565b60405180910390f35b3480156101ce575f80fd5b50d380156101da575f80fd5b50d280156101e6575f80fd5b506101ef610795565b6040516101fc9190611c09565b60405180910390f35b348015610210575f80fd5b50d3801561021c575f80fd5b50d28015610228575f80fd5b50610243600480360381019061023e9190611c22565b61079e565b6040516102509190611be1565b60405180910390f35b348015610264575f80fd5b50d38015610270575f80fd5b50d2801561027c575f80fd5b506102856107cc565b6040516102929190611c8d565b60405180910390f35b3480156102a6575f80fd5b50d380156102b2575f80fd5b50d280156102be575f80fd5b506102d960048036038101906102d49190611b89565b6107d4565b005b3480156102e6575f80fd5b50d380156102f2575f80fd5b50d280156102fe575f80fd5b5061031960048036038101906103149190611ca6565b6107e2565b005b348015610326575f80fd5b50d38015610332575f80fd5b50d2801561033e575f80fd5b5061035960048036038101906103549190611cd1565b6107ef565b6040516103669190611c09565b60405180910390f35b34801561037a575f80fd5b50d38015610386575f80fd5b50d28015610392575f80fd5b5061039b610834565b6040516103a89190611c09565b60405180910390f35b3480156103bc575f80fd5b50d380156103c8575f80fd5b50d280156103d4575f80fd5b506103dd6108c7565b6040516103ea9190611d57565b60405180910390f35b3480156103fe575f80fd5b50d3801561040a575f80fd5b50d28015610416575f80fd5b5061041f6108eb565b60405161042d929190611d8c565b60405180910390f35b610450600480360381019061044b9190611e16565b610a0f565b60405161045d9190611eaf565b60405180910390f35b348015610471575f80fd5b50d3801561047d575f80fd5b50d28015610489575f80fd5b50610492610e1f565b60405161049f9190611aa0565b60405180910390f35b3480156104b3575f80fd5b50d380156104bf575f80fd5b50d280156104cb575f80fd5b506104e660048036038101906104e19190611efb565b610eaf565b6040516104f39190611be1565b60405180910390f35b348015610507575f80fd5b50d38015610513575f80fd5b50d2801561051f575f80fd5b50610528610ecc565b6040516105359190611f26565b60405180910390f35b348015610549575f80fd5b50d38015610555575f80fd5b50d28015610561575f80fd5b5061057c60048036038101906105779190611b89565b610ef0565b6040516105899190611be1565b60405180910390f35b34801561059d575f80fd5b50d380156105a9575f80fd5b50d280156105b5575f80fd5b506105d060048036038101906105cb9190611fa0565b610f12565b6040516105dd9190611be1565b60405180910390f35b3480156105f1575f80fd5b50d380156105fd575f80fd5b50d28015610609575f80fd5b50610624600480360381019061061f9190611fa0565b611219565b604051610632929190611feb565b60405180910390f35b348015610646575f80fd5b50d38015610652575f80fd5b50d2801561065e575f80fd5b5061067960048036038101906106749190611efb565b6112cb565b6040516106869190611be1565b60405180910390f35b34801561069a575f80fd5b50d380156106a6575f80fd5b50d280156106b2575f80fd5b506106cd60048036038101906106c89190612019565b6112f1565b6040516106da9190611c09565b60405180910390f35b6060600380546106f290612084565b80601f016020809104026020016040519081016040528092919081815260200182805461071e90612084565b80156107695780601f1061074057610100808354040283529160200191610769565b820191905f5260205f20905b81548152906001019060200180831161074c57829003601f168201915b5050505050905090565b5f8061077d611373565b905061078a81858561137a565b600191505092915050565b5f600254905090565b5f806107a8611373565b90506107b585828561138c565b6107c085858561141f565b60019150509392505050565b5f6012905090565b6107de828261150f565b5050565b6107ec338261158e565b50565b5f805f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16631a90a2196040518163ffffffff1660e01b8152600401602060405180830381865afa15801561089e573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906108c291906120c8565b905090565b7f000000000000000000000000000000000000000000000000000000000000000081565b5f807f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16639a8a05926040518163ffffffff1660e01b8152600401602060405180830381865afa158015610956573d5f803e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061097a9190612107565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16631a90a2196040518163ffffffff1660e01b8152600401602060405180830381865afa1580156109e3573d5f803e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610a0791906120c8565b915091509091565b5f82610a1a336107ef565b1015610a5b576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610a529061217c565b60405180910390fd5b7f000000000000000000000000000000000000000000000000000000000000000061ffff168661ffff1603610ac5576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610abc906121e4565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610b33576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610b2a9061224c565b60405180910390fd5b5f8311610b75576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610b6c906122b4565b60405180910390fd5b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16631a90a2196040518163ffffffff1660e01b8152600401602060405180830381865afa158015610bdf573d5f803e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610c0391906120c8565b905080341015610c48576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610c3f9061231c565b60405180910390fd5b610c52338561158e565b5f610c5c8761160d565b7f60f91c4983cd584ea4c48485b610cc900d867e1f806ae08ae922a90379466d3587877f0000000000000000000000000000000000000000000000000000000000000000610ca93061160d565b42604051602001610cc09796959493929190612358565b60405160208183030381529060405290507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663b19a437e83868460c86040518563ffffffff1660e01b8152600401610d309392919061245f565b60206040518083038185885af1158015610d4c573d5f803e3d5ffd5b50505050506040513d601f19601f82011682018060405250810190610d7191906124c5565b925081341115610dcb573373ffffffffffffffffffffffffffffffffffffffff166108fc8334610da1919061251d565b90811502906040515f60405180830381858888f19350505050158015610dc9573d5f803e3d5ffd5b505b7f1af839d04b116fbfc6c8e7054ff35eea7302c7eae01f249491a8fdb36c33200c83338a610df88b61160d565b8a8a604051610e0c96959493929190612550565b60405180910390a1505095945050505050565b606060048054610e2e90612084565b80601f0160208091040260200160405190810160405280929190818152602001828054610e5a90612084565b8015610ea55780601f10610e7c57610100808354040283529160200191610ea5565b820191905f5260205f20905b815481529060010190602001808311610e8857829003601f168201915b5050505050905090565b6005602052805f5260405f205f915054906101000a900460ff1681565b7f000000000000000000000000000000000000000000000000000000000000000081565b5f80610efa611373565b9050610f0781858561141f565b600191505092915050565b5f805f807f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663c0fd8bde87876040518363ffffffff1660e01b8152600401610f719291906125e9565b5f60405180830381865afa158015610f8b573d5f803e3d5ffd5b505050506040513d5f823e3d601f19601f82011682018060405250810190610fb39190612abd565b925092509250818190610ffc576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610ff39190611aa0565b60405180910390fd5b5060055f84610140015181526020019081526020015f205f9054906101000a900460ff1615611060576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161105790612b8f565b60405180910390fd5b5f805f805f805f8960e0015180602001905181019061107f9190612bf3565b96509650965096509650965096507f60f91c4983cd584ea4c48485b610cc900d867e1f806ae08ae922a90379466d3586146110ef576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016110e690612cda565b60405180910390fd5b6110f83061160d565b8714611139576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161113090612d42565b60405180910390fd5b62015180816111489190612d60565b4210611189576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161118090612ddd565b60405180910390fd5b611193858561150f565b600160055f8c610140015181526020019081526020015f205f6101000a81548160ff0219169083151502179055507f39f79d45ea2315f6088956a9bacc2d784f7b2b721ce52081d9bd3c063175e2b68a6101400151848488886040516111fd959493929190612dfb565b60405180910390a160019a505050505050505050505092915050565b5f60607f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663c0fd8bde85856040518363ffffffff1660e01b81526004016112779291906125e9565b5f60405180830381865afa158015611291573d5f803e3d5ffd5b505050506040513d5f823e3d601f19601f820116820180604052508101906112b99190612abd565b90915080925081935050509250929050565b5f60055f8381526020019081526020015f205f9054906101000a900460ff169050919050565b5f60015f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905092915050565b5f33905090565b611387838383600161162e565b505050565b5f61139784846112f1565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff811015611419578181101561140a578281836040517ffb8f41b200000000000000000000000000000000000000000000000000000000815260040161140193929190612e4c565b60405180910390fd5b61141884848484035f61162e565b5b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361148f575f6040517f96c6fd1e0000000000000000000000000000000000000000000000000000000081526004016114869190612e81565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16036114ff575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016114f69190612e81565b60405180910390fd5b61150a8383836117fd565b505050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361157f575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016115769190612e81565b60405180910390fd5b61158a5f83836117fd565b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16036115fe575f6040517f96c6fd1e0000000000000000000000000000000000000000000000000000000081526004016115f59190612e81565b60405180910390fd5b611609825f836117fd565b5050565b5f8173ffffffffffffffffffffffffffffffffffffffff165f1b9050919050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff160361169e575f6040517fe602df050000000000000000000000000000000000000000000000000000000081526004016116959190612e81565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361170e575f6040517f94280d620000000000000000000000000000000000000000000000000000000081526004016117059190612e81565b60405180910390fd5b8160015f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f208190555080156117f7578273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925846040516117ee9190611c09565b60405180910390a35b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361184d578060025f8282546118419190612d60565b9250508190555061191b565b5f805f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050818110156118d6578381836040517fe450d38c0000000000000000000000000000000000000000000000000000000081526004016118cd93929190612e4c565b60405180910390fd5b8181035f808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603611962578060025f82825403925050819055506119ac565b805f808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051611a099190611c09565b60405180910390a3505050565b5f81519050919050565b5f82825260208201905092915050565b5f5b83811015611a4d578082015181840152602081019050611a32565b5f8484015250505050565b5f601f19601f8301169050919050565b5f611a7282611a16565b611a7c8185611a20565b9350611a8c818560208601611a30565b611a9581611a58565b840191505092915050565b5f6020820190508181035f830152611ab88184611a68565b905092915050565b5f604051905090565b5f80fd5b5f80fd5b5f74ffffffffffffffffffffffffffffffffffffffffff82169050919050565b611afa81611ad1565b8114611b04575f80fd5b50565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f611b3082611b07565b9050919050565b5f81359050611b4581611af1565b611b4e81611b26565b905092915050565b5f819050919050565b611b6881611b56565b8114611b72575f80fd5b50565b5f81359050611b8381611b5f565b92915050565b5f8060408385031215611b9f57611b9e611ac9565b5b5f611bac85828601611b37565b9250506020611bbd85828601611b75565b9150509250929050565b5f8115159050919050565b611bdb81611bc7565b82525050565b5f602082019050611bf45f830184611bd2565b92915050565b611c0381611b56565b82525050565b5f602082019050611c1c5f830184611bfa565b92915050565b5f805f60608486031215611c3957611c38611ac9565b5b5f611c4686828701611b37565b9350506020611c5786828701611b37565b9250506040611c6886828701611b75565b9150509250925092565b5f60ff82169050919050565b611c8781611c72565b82525050565b5f602082019050611ca05f830184611c7e565b92915050565b5f60208284031215611cbb57611cba611ac9565b5b5f611cc884828501611b75565b91505092915050565b5f60208284031215611ce657611ce5611ac9565b5b5f611cf384828501611b37565b91505092915050565b5f819050919050565b5f611d1f611d1a611d1584611b07565b611cfc565b611b07565b9050919050565b5f611d3082611d05565b9050919050565b5f611d4182611d26565b9050919050565b611d5181611d37565b82525050565b5f602082019050611d6a5f830184611d48565b92915050565b5f61ffff82169050919050565b611d8681611d70565b82525050565b5f604082019050611d9f5f830185611d7d565b611dac6020830184611bfa565b9392505050565b611dbc81611d70565b8114611dc6575f80fd5b50565b5f81359050611dd781611db3565b92915050565b5f63ffffffff82169050919050565b611df581611ddd565b8114611dff575f80fd5b50565b5f81359050611e1081611dec565b92915050565b5f805f805f60a08688031215611e2f57611e2e611ac9565b5b5f611e3c88828901611dc9565b9550506020611e4d88828901611b37565b9450506040611e5e88828901611b37565b9350506060611e6f88828901611b75565b9250506080611e8088828901611e02565b9150509295509295909350565b5f67ffffffffffffffff82169050919050565b611ea981611e8d565b82525050565b5f602082019050611ec25f830184611ea0565b92915050565b5f819050919050565b611eda81611ec8565b8114611ee4575f80fd5b50565b5f81359050611ef581611ed1565b92915050565b5f60208284031215611f1057611f0f611ac9565b5b5f611f1d84828501611ee7565b91505092915050565b5f602082019050611f395f830184611d7d565b92915050565b5f80fd5b5f80fd5b5f80fd5b5f8083601f840112611f6057611f5f611f3f565b5b8235905067ffffffffffffffff811115611f7d57611f7c611f43565b5b602083019150836001820283011115611f9957611f98611f47565b5b9250929050565b5f8060208385031215611fb657611fb5611ac9565b5b5f83013567ffffffffffffffff811115611fd357611fd2611acd565b5b611fdf85828601611f4b565b92509250509250929050565b5f604082019050611ffe5f830185611bd2565b81810360208301526120108184611a68565b90509392505050565b5f806040838503121561202f5761202e611ac9565b5b5f61203c85828601611b37565b925050602061204d85828601611b37565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061209b57607f821691505b6020821081036120ae576120ad612057565b5b50919050565b5f815190506120c281611b5f565b92915050565b5f602082840312156120dd576120dc611ac9565b5b5f6120ea848285016120b4565b91505092915050565b5f8151905061210181611db3565b92915050565b5f6020828403121561211c5761211b611ac9565b5b5f612129848285016120f3565b91505092915050565b7f496e73756666696369656e742062616c616e63650000000000000000000000005f82015250565b5f612166601483611a20565b915061217182612132565b602082019050919050565b5f6020820190508181035f8301526121938161215a565b9050919050565b7f43616e6e6f742073656e6420746f2073616d6520636861696e000000000000005f82015250565b5f6121ce601983611a20565b91506121d98261219a565b602082019050919050565b5f6020820190508181035f8301526121fb816121c2565b9050919050565b7f496e76616c696420726563697069656e740000000000000000000000000000005f82015250565b5f612236601183611a20565b915061224182612202565b602082019050919050565b5f6020820190508181035f8301526122638161222a565b9050919050565b7f416d6f756e74206d7573742062652067726561746572207468616e20300000005f82015250565b5f61229e601d83611a20565b91506122a98261226a565b602082019050919050565b5f6020820190508181035f8301526122cb81612292565b9050919050565b7f496e73756666696369656e7420666565000000000000000000000000000000005f82015250565b5f612306601083611a20565b9150612311826122d2565b602082019050919050565b5f6020820190508181035f830152612333816122fa565b9050919050565b61234381611ec8565b82525050565b61235281611b26565b82525050565b5f60e08201905061236b5f83018a61233a565b612378602083018961233a565b6123856040830188612349565b6123926060830187611bfa565b61239f6080830186611d7d565b6123ac60a083018561233a565b6123b960c0830184611bfa565b98975050505050505050565b6123ce81611ddd565b82525050565b5f81519050919050565b5f82825260208201905092915050565b5f6123f8826123d4565b61240281856123de565b9350612412818560208601611a30565b61241b81611a58565b840191505092915050565b5f819050919050565b5f61244961244461243f84612426565b611cfc565b611c72565b9050919050565b6124598161242f565b82525050565b5f6060820190506124725f8301866123c5565b818103602083015261248481856123ee565b90506124936040830184612450565b949350505050565b6124a481611e8d565b81146124ae575f80fd5b50565b5f815190506124bf8161249b565b92915050565b5f602082840312156124da576124d9611ac9565b5b5f6124e7848285016124b1565b91505092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f61252782611b56565b915061253283611b56565b925082820390508181111561254a576125496124f0565b5b92915050565b5f60c0820190506125635f830189611ea0565b6125706020830188612349565b61257d6040830187611d7d565b61258a606083018661233a565b6125976080830185612349565b6125a460a0830184611bfa565b979650505050505050565b828183375f83830152505050565b5f6125c883856123de565b93506125d58385846125af565b6125de83611a58565b840190509392505050565b5f6020820190508181035f8301526126028184866125bd565b90509392505050565b5f80fd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b61264582611a58565b810181811067ffffffffffffffff821117156126645761266361260f565b5b80604052505050565b5f612676611ac0565b9050612682828261263c565b919050565b5f80fd5b61269481611c72565b811461269e575f80fd5b50565b5f815190506126af8161268b565b92915050565b5f815190506126c381611dec565b92915050565b5f815190506126d781611ed1565b92915050565b5f80fd5b5f67ffffffffffffffff8211156126fb576126fa61260f565b5b61270482611a58565b9050602081019050919050565b5f61272361271e846126e1565b61266d565b90508281526020810184848401111561273f5761273e6126dd565b5b61274a848285611a30565b509392505050565b5f82601f83011261276657612765611f3f565b5b8151612776848260208601612711565b91505092915050565b5f67ffffffffffffffff8211156127995761279861260f565b5b602082029050602081019050919050565b5f608082840312156127bf576127be61260b565b5b6127c9608061266d565b90505f6127d8848285016126c9565b5f8301525060206127eb848285016126c9565b60208301525060406127ff848285016126a1565b6040830152506060612813848285016126a1565b60608301525092915050565b5f61283161282c8461277f565b61266d565b9050808382526020820190506080840283018581111561285457612853611f47565b5b835b8181101561287d578061286988826127aa565b845260208401935050608081019050612856565b5050509392505050565b5f82601f83011261289b5761289a611f3f565b5b81516128ab84826020860161281f565b91505092915050565b5f61016082840312156128ca576128c961260b565b5b6128d561016061266d565b90505f6128e4848285016126a1565b5f8301525060206128f7848285016126b5565b602083015250604061290b848285016126b5565b604083015250606061291f848285016120f3565b6060830152506080612933848285016126c9565b60808301525060a0612947848285016124b1565b60a08301525060c061295b848285016126a1565b60c08301525060e082015167ffffffffffffffff81111561297f5761297e612687565b5b61298b84828501612752565b60e0830152506101006129a0848285016126b5565b6101008301525061012082015167ffffffffffffffff8111156129c6576129c5612687565b5b6129d284828501612887565b610120830152506101406129e8848285016126c9565b6101408301525092915050565b6129fe81611bc7565b8114612a08575f80fd5b50565b5f81519050612a19816129f5565b92915050565b5f67ffffffffffffffff821115612a3957612a3861260f565b5b612a4282611a58565b9050602081019050919050565b5f612a61612a5c84612a1f565b61266d565b905082815260208101848484011115612a7d57612a7c6126dd565b5b612a88848285611a30565b509392505050565b5f82601f830112612aa457612aa3611f3f565b5b8151612ab4848260208601612a4f565b91505092915050565b5f805f60608486031215612ad457612ad3611ac9565b5b5f84015167ffffffffffffffff811115612af157612af0611acd565b5b612afd868287016128b4565b9350506020612b0e86828701612a0b565b925050604084015167ffffffffffffffff811115612b2f57612b2e611acd565b5b612b3b86828701612a90565b9150509250925092565b7f56414120616c72656164792070726f63657373656400000000000000000000005f82015250565b5f612b79601583611a20565b9150612b8482612b45565b602082019050919050565b5f6020820190508181035f830152612ba681612b6d565b9050919050565b612bb681611ad1565b8114612bc0575f80fd5b50565b5f612bcd82611b07565b9050919050565b5f81519050612be281612bad565b612beb81612bc3565b905092915050565b5f805f805f805f60e0888a031215612c0e57612c0d611ac9565b5b5f612c1b8a828b016126c9565b9750506020612c2c8a828b016126c9565b9650506040612c3d8a828b01612bd4565b9550506060612c4e8a828b016120b4565b9450506080612c5f8a828b016120f3565b93505060a0612c708a828b016126c9565b92505060c0612c818a828b016120b4565b91505092959891949750929550565b7f496e76616c6964206d65737361676520747970650000000000000000000000005f82015250565b5f612cc4601483611a20565b9150612ccf82612c90565b602082019050919050565b5f6020820190508181035f830152612cf181612cb8565b9050919050565b7f4e6f742074617267657420636f6e7472616374000000000000000000000000005f82015250565b5f612d2c601383611a20565b9150612d3782612cf8565b602082019050919050565b5f6020820190508181035f830152612d5981612d20565b9050919050565b5f612d6a82611b56565b9150612d7583611b56565b9250828201905080821115612d8d57612d8c6124f0565b5b92915050565b7f56414120657870697265640000000000000000000000000000000000000000005f82015250565b5f612dc7600b83611a20565b9150612dd282612d93565b602082019050919050565b5f6020820190508181035f830152612df481612dbb565b9050919050565b5f60a082019050612e0e5f83018861233a565b612e1b6020830187611d7d565b612e28604083018661233a565b612e356060830185612349565b612e426080830184611bfa565b9695505050505050565b5f606082019050612e5f5f830186612349565b612e6c6020830185611bfa565b612e796040830184611bfa565b949350505050565b5f602082019050612e945f830184612349565b9291505056fea26474726f6e58221220e8d6430ef1d438bd9ace52cb91fbf2e8cd3048d0640b94f30f64b67fd02b836364736f6c63430008160033000000000000000000000000208ba0a882d151e36c16a93fe952f1d114ec5ee4000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000098968000000000000000000000000000000000000000000000000000000000000000036d696e000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000046d696e3100000000000000000000000000000000000000000000000000000000"
	constructorParams := "0x60c060405234801562000010575f80fd5b50d380156200001d575f80fd5b50d280156200002a575f80fd5b5060405162003b9138038062003b918339818101604052810190620000509190620006bd565b8282816003908162000063919062000998565b50806004908162000075919062000998565b5050505f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603620000e9576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401620000e09062000ada565b60405180910390fd5b8373ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff168152505060805173ffffffffffffffffffffffffffffffffffffffff16639a8a05926040518163ffffffff1660e01b8152600401602060405180830381865afa15801562000169573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906200018f919062000b36565b61ffff1660a08161ffff16815250505f811115620001ba57620001b93382620001c460201b60201c565b5b5050505062000c60565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160362000237575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016200022e919062000b77565b60405180910390fd5b6200024a5f83836200024e60201b60201c565b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1603620002a2578060025f82825462000295919062000bbf565b9250508190555062000373565b5f805f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050818110156200032e578381836040517fe450d38c000000000000000000000000000000000000000000000000000000008152600401620003259392919062000c0a565b60405180910390fd5b8181035f808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603620003bc578060025f828254039250508190555062000406565b805f808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef8360405162000465919062000c45565b60405180910390a3505050565b5f604051905090565b5f80fd5b5f80fd5b5f74ffffffffffffffffffffffffffffffffffffffffff82169050919050565b620004ae8162000483565b8114620004b9575f80fd5b50565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f620004e782620004bc565b9050919050565b5f81519050620004fe81620004a3565b6200050981620004db565b905092915050565b5f80fd5b5f80fd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b620005618262000519565b810181811067ffffffffffffffff8211171562000583576200058262000529565b5b80604052505050565b5f6200059762000472565b9050620005a5828262000556565b919050565b5f67ffffffffffffffff821115620005c757620005c662000529565b5b620005d28262000519565b9050602081019050919050565b5f5b83811015620005fe578082015181840152602081019050620005e1565b5f8484015250505050565b5f6200061f6200061984620005aa565b6200058c565b9050828152602081018484840111156200063e576200063d62000515565b5b6200064b848285620005df565b509392505050565b5f82601f8301126200066a576200066962000511565b5b81516200067c84826020860162000609565b91505092915050565b5f819050919050565b620006998162000685565b8114620006a4575f80fd5b50565b5f81519050620006b7816200068e565b92915050565b5f805f8060808587031215620006d857620006d76200047b565b5b5f620006e787828801620004ee565b945050602085015167ffffffffffffffff8111156200070b576200070a6200047f565b5b620007198782880162000653565b935050604085015167ffffffffffffffff8111156200073d576200073c6200047f565b5b6200074b8782880162000653565b92505060606200075e87828801620006a7565b91505092959194509250565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f6002820490506001821680620007b957607f821691505b602082108103620007cf57620007ce62000774565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f60088302620008337fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82620007f6565b6200083f8683620007f6565b95508019841693508086168417925050509392505050565b5f819050919050565b5f620008806200087a620008748462000685565b62000857565b62000685565b9050919050565b5f819050919050565b6200089b8362000860565b620008b3620008aa8262000887565b84845462000802565b825550505050565b5f90565b620008c9620008bb565b620008d681848462000890565b505050565b5b81811015620008fd57620008f15f82620008bf565b600181019050620008dc565b5050565b601f8211156200094c576200091681620007d5565b6200092184620007e7565b8101602085101562000931578190505b620009496200094085620007e7565b830182620008db565b50505b505050565b5f82821c905092915050565b5f6200096e5f198460080262000951565b1980831691505092915050565b5f6200098883836200095d565b9150826002028217905092915050565b620009a3826200076a565b67ffffffffffffffff811115620009bf57620009be62000529565b5b620009cb8254620007a1565b620009d882828562000901565b5f60209050601f83116001811462000a0e575f8415620009f9578287015190505b62000a0585826200097b565b86555062000a74565b601f19841662000a1e86620007d5565b5f5b8281101562000a475784890151825560018201915060208501945060208101905062000a20565b8683101562000a67578489015162000a63601f8916826200095d565b8355505b6001600288020188555050505b505050505050565b5f82825260208201905092915050565b7f496e76616c696420776f726d686f6c65206164647265737300000000000000005f82015250565b5f62000ac260188362000a7c565b915062000acf8262000a8c565b602082019050919050565b5f6020820190508181035f83015262000af38162000ab4565b9050919050565b5f61ffff82169050919050565b62000b128162000afa565b811462000b1d575f80fd5b50565b5f8151905062000b308162000b07565b92915050565b5f6020828403121562000b4e5762000b4d6200047b565b5b5f62000b5d8482850162000b20565b91505092915050565b62000b7181620004db565b82525050565b5f60208201905062000b8c5f83018462000b66565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f62000bcb8262000685565b915062000bd88362000685565b925082820190508082111562000bf35762000bf262000b92565b5b92915050565b62000c048162000685565b82525050565b5f60608201905062000c1f5f83018662000b66565b62000c2e602083018562000bf9565b62000c3d604083018462000bf9565b949350505050565b5f60208201905062000c5a5f83018462000bf9565b92915050565b60805160a051612ed062000cc15f395f8181610a5d01528181610c810152610ece01525f8181610837015281816108c9015281816108ef0152818161097c01528181610b7801528181610cd301528181610f18015261121e0152612ed05ff3fe608060405260043610610122575f3560e01c8063889be5331161009f578063a9059cbb11610063578063a9059cbb1461053e578063b02275b114610592578063d294cd8e146105e6578063dcb2ee3c1461063b578063dd62ed3e1461068f57610129565b8063889be533146103f35780639569327b1461043657806395d89b41146104665780639886cdcd146104a85780639a8a0592146104fc57610129565b806340c10f19116100e657806340c10f191461029b57806342966c68146102db57806370a082311461031b5780637f18dc851461036f57806384acd1bb146103b157610129565b806306fdde031461012d578063095ea7b31461016f57806318160ddd146101c357806323b872dd14610205578063313ce5671461025957610129565b3661012957005b5f80fd5b348015610138575f80fd5b50d38015610144575f80fd5b50d28015610150575f80fd5b506101596106e3565b6040516101669190611aa0565b60405180910390f35b34801561017a575f80fd5b50d38015610186575f80fd5b50d28015610192575f80fd5b506101ad60048036038101906101a89190611b89565b610773565b6040516101ba9190611be1565b60405180910390f35b3480156101ce575f80fd5b50d380156101da575f80fd5b50d280156101e6575f80fd5b506101ef610795565b6040516101fc9190611c09565b60405180910390f35b348015610210575f80fd5b50d3801561021c575f80fd5b50d28015610228575f80fd5b50610243600480360381019061023e9190611c22565b61079e565b6040516102509190611be1565b60405180910390f35b348015610264575f80fd5b50d38015610270575f80fd5b50d2801561027c575f80fd5b506102856107cc565b6040516102929190611c8d565b60405180910390f35b3480156102a6575f80fd5b50d380156102b2575f80fd5b50d280156102be575f80fd5b506102d960048036038101906102d49190611b89565b6107d4565b005b3480156102e6575f80fd5b50d380156102f2575f80fd5b50d280156102fe575f80fd5b5061031960048036038101906103149190611ca6565b6107e2565b005b348015610326575f80fd5b50d38015610332575f80fd5b50d2801561033e575f80fd5b5061035960048036038101906103549190611cd1565b6107ef565b6040516103669190611c09565b60405180910390f35b34801561037a575f80fd5b50d38015610386575f80fd5b50d28015610392575f80fd5b5061039b610834565b6040516103a89190611c09565b60405180910390f35b3480156103bc575f80fd5b50d380156103c8575f80fd5b50d280156103d4575f80fd5b506103dd6108c7565b6040516103ea9190611d57565b60405180910390f35b3480156103fe575f80fd5b50d3801561040a575f80fd5b50d28015610416575f80fd5b5061041f6108eb565b60405161042d929190611d8c565b60405180910390f35b610450600480360381019061044b9190611e16565b610a0f565b60405161045d9190611eaf565b60405180910390f35b348015610471575f80fd5b50d3801561047d575f80fd5b50d28015610489575f80fd5b50610492610e1f565b60405161049f9190611aa0565b60405180910390f35b3480156104b3575f80fd5b50d380156104bf575f80fd5b50d280156104cb575f80fd5b506104e660048036038101906104e19190611efb565b610eaf565b6040516104f39190611be1565b60405180910390f35b348015610507575f80fd5b50d38015610513575f80fd5b50d2801561051f575f80fd5b50610528610ecc565b6040516105359190611f26565b60405180910390f35b348015610549575f80fd5b50d38015610555575f80fd5b50d28015610561575f80fd5b5061057c60048036038101906105779190611b89565b610ef0565b6040516105899190611be1565b60405180910390f35b34801561059d575f80fd5b50d380156105a9575f80fd5b50d280156105b5575f80fd5b506105d060048036038101906105cb9190611fa0565b610f12565b6040516105dd9190611be1565b60405180910390f35b3480156105f1575f80fd5b50d380156105fd575f80fd5b50d28015610609575f80fd5b50610624600480360381019061061f9190611fa0565b611219565b604051610632929190611feb565b60405180910390f35b348015610646575f80fd5b50d38015610652575f80fd5b50d2801561065e575f80fd5b5061067960048036038101906106749190611efb565b6112cb565b6040516106869190611be1565b60405180910390f35b34801561069a575f80fd5b50d380156106a6575f80fd5b50d280156106b2575f80fd5b506106cd60048036038101906106c89190612019565b6112f1565b6040516106da9190611c09565b60405180910390f35b6060600380546106f290612084565b80601f016020809104026020016040519081016040528092919081815260200182805461071e90612084565b80156107695780601f1061074057610100808354040283529160200191610769565b820191905f5260205f20905b81548152906001019060200180831161074c57829003601f168201915b5050505050905090565b5f8061077d611373565b905061078a81858561137a565b600191505092915050565b5f600254905090565b5f806107a8611373565b90506107b585828561138c565b6107c085858561141f565b60019150509392505050565b5f6012905090565b6107de828261150f565b5050565b6107ec338261158e565b50565b5f805f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050919050565b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16631a90a2196040518163ffffffff1660e01b8152600401602060405180830381865afa15801561089e573d5f803e3d5ffd5b505050506040513d601f19601f820116820180604052508101906108c291906120c8565b905090565b7f000000000000000000000000000000000000000000000000000000000000000081565b5f807f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16639a8a05926040518163ffffffff1660e01b8152600401602060405180830381865afa158015610956573d5f803e3d5ffd5b505050506040513d601f19601f8201168201806040525081019061097a9190612107565b7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16631a90a2196040518163ffffffff1660e01b8152600401602060405180830381865afa1580156109e3573d5f803e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610a0791906120c8565b915091509091565b5f82610a1a336107ef565b1015610a5b576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610a529061217c565b60405180910390fd5b7f000000000000000000000000000000000000000000000000000000000000000061ffff168661ffff1603610ac5576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610abc906121e4565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1603610b33576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610b2a9061224c565b60405180910390fd5b5f8311610b75576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610b6c906122b4565b60405180910390fd5b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16631a90a2196040518163ffffffff1660e01b8152600401602060405180830381865afa158015610bdf573d5f803e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610c0391906120c8565b905080341015610c48576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610c3f9061231c565b60405180910390fd5b610c52338561158e565b5f610c5c8761160d565b7f60f91c4983cd584ea4c48485b610cc900d867e1f806ae08ae922a90379466d3587877f0000000000000000000000000000000000000000000000000000000000000000610ca93061160d565b42604051602001610cc09796959493929190612358565b60405160208183030381529060405290507f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663b19a437e83868460c86040518563ffffffff1660e01b8152600401610d309392919061245f565b60206040518083038185885af1158015610d4c573d5f803e3d5ffd5b50505050506040513d601f19601f82011682018060405250810190610d7191906124c5565b925081341115610dcb573373ffffffffffffffffffffffffffffffffffffffff166108fc8334610da1919061251d565b90811502906040515f60405180830381858888f19350505050158015610dc9573d5f803e3d5ffd5b505b7f1af839d04b116fbfc6c8e7054ff35eea7302c7eae01f249491a8fdb36c33200c83338a610df88b61160d565b8a8a604051610e0c96959493929190612550565b60405180910390a1505095945050505050565b606060048054610e2e90612084565b80601f0160208091040260200160405190810160405280929190818152602001828054610e5a90612084565b8015610ea55780601f10610e7c57610100808354040283529160200191610ea5565b820191905f5260205f20905b815481529060010190602001808311610e8857829003601f168201915b5050505050905090565b6005602052805f5260405f205f915054906101000a900460ff1681565b7f000000000000000000000000000000000000000000000000000000000000000081565b5f80610efa611373565b9050610f0781858561141f565b600191505092915050565b5f805f807f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663c0fd8bde87876040518363ffffffff1660e01b8152600401610f719291906125e9565b5f60405180830381865afa158015610f8b573d5f803e3d5ffd5b505050506040513d5f823e3d601f19601f82011682018060405250810190610fb39190612abd565b925092509250818190610ffc576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610ff39190611aa0565b60405180910390fd5b5060055f84610140015181526020019081526020015f205f9054906101000a900460ff1615611060576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161105790612b8f565b60405180910390fd5b5f805f805f805f8960e0015180602001905181019061107f9190612bf3565b96509650965096509650965096507f60f91c4983cd584ea4c48485b610cc900d867e1f806ae08ae922a90379466d3586146110ef576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016110e690612cda565b60405180910390fd5b6110f83061160d565b8714611139576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161113090612d42565b60405180910390fd5b62015180816111489190612d60565b4210611189576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161118090612ddd565b60405180910390fd5b611193858561150f565b600160055f8c610140015181526020019081526020015f205f6101000a81548160ff0219169083151502179055507f39f79d45ea2315f6088956a9bacc2d784f7b2b721ce52081d9bd3c063175e2b68a6101400151848488886040516111fd959493929190612dfb565b60405180910390a160019a505050505050505050505092915050565b5f60607f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff1663c0fd8bde85856040518363ffffffff1660e01b81526004016112779291906125e9565b5f60405180830381865afa158015611291573d5f803e3d5ffd5b505050506040513d5f823e3d601f19601f820116820180604052508101906112b99190612abd565b90915080925081935050509250929050565b5f60055f8381526020019081526020015f205f9054906101000a900460ff169050919050565b5f60015f8473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2054905092915050565b5f33905090565b611387838383600161162e565b505050565b5f61139784846112f1565b90507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff811015611419578181101561140a578281836040517ffb8f41b200000000000000000000000000000000000000000000000000000000815260040161140193929190612e4c565b60405180910390fd5b61141884848484035f61162e565b5b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361148f575f6040517f96c6fd1e0000000000000000000000000000000000000000000000000000000081526004016114869190612e81565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16036114ff575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016114f69190612e81565b60405180910390fd5b61150a8383836117fd565b505050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff160361157f575f6040517fec442f050000000000000000000000000000000000000000000000000000000081526004016115769190612e81565b60405180910390fd5b61158a5f83836117fd565b5050565b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16036115fe575f6040517f96c6fd1e0000000000000000000000000000000000000000000000000000000081526004016115f59190612e81565b60405180910390fd5b611609825f836117fd565b5050565b5f8173ffffffffffffffffffffffffffffffffffffffff165f1b9050919050565b5f73ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff160361169e575f6040517fe602df050000000000000000000000000000000000000000000000000000000081526004016116959190612e81565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361170e575f6040517f94280d620000000000000000000000000000000000000000000000000000000081526004016117059190612e81565b60405180910390fd5b8160015f8673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f208190555080156117f7578273ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff167f8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925846040516117ee9190611c09565b60405180910390a35b50505050565b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff160361184d578060025f8282546118419190612d60565b9250508190555061191b565b5f805f8573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f20549050818110156118d6578381836040517fe450d38c0000000000000000000000000000000000000000000000000000000081526004016118cd93929190612e4c565b60405180910390fd5b8181035f808673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f2081905550505b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1603611962578060025f82825403925050819055506119ac565b805f808473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020015f205f82825401925050819055505b8173ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef83604051611a099190611c09565b60405180910390a3505050565b5f81519050919050565b5f82825260208201905092915050565b5f5b83811015611a4d578082015181840152602081019050611a32565b5f8484015250505050565b5f601f19601f8301169050919050565b5f611a7282611a16565b611a7c8185611a20565b9350611a8c818560208601611a30565b611a9581611a58565b840191505092915050565b5f6020820190508181035f830152611ab88184611a68565b905092915050565b5f604051905090565b5f80fd5b5f80fd5b5f74ffffffffffffffffffffffffffffffffffffffffff82169050919050565b611afa81611ad1565b8114611b04575f80fd5b50565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f611b3082611b07565b9050919050565b5f81359050611b4581611af1565b611b4e81611b26565b905092915050565b5f819050919050565b611b6881611b56565b8114611b72575f80fd5b50565b5f81359050611b8381611b5f565b92915050565b5f8060408385031215611b9f57611b9e611ac9565b5b5f611bac85828601611b37565b9250506020611bbd85828601611b75565b9150509250929050565b5f8115159050919050565b611bdb81611bc7565b82525050565b5f602082019050611bf45f830184611bd2565b92915050565b611c0381611b56565b82525050565b5f602082019050611c1c5f830184611bfa565b92915050565b5f805f60608486031215611c3957611c38611ac9565b5b5f611c4686828701611b37565b9350506020611c5786828701611b37565b9250506040611c6886828701611b75565b9150509250925092565b5f60ff82169050919050565b611c8781611c72565b82525050565b5f602082019050611ca05f830184611c7e565b92915050565b5f60208284031215611cbb57611cba611ac9565b5b5f611cc884828501611b75565b91505092915050565b5f60208284031215611ce657611ce5611ac9565b5b5f611cf384828501611b37565b91505092915050565b5f819050919050565b5f611d1f611d1a611d1584611b07565b611cfc565b611b07565b9050919050565b5f611d3082611d05565b9050919050565b5f611d4182611d26565b9050919050565b611d5181611d37565b82525050565b5f602082019050611d6a5f830184611d48565b92915050565b5f61ffff82169050919050565b611d8681611d70565b82525050565b5f604082019050611d9f5f830185611d7d565b611dac6020830184611bfa565b9392505050565b611dbc81611d70565b8114611dc6575f80fd5b50565b5f81359050611dd781611db3565b92915050565b5f63ffffffff82169050919050565b611df581611ddd565b8114611dff575f80fd5b50565b5f81359050611e1081611dec565b92915050565b5f805f805f60a08688031215611e2f57611e2e611ac9565b5b5f611e3c88828901611dc9565b9550506020611e4d88828901611b37565b9450506040611e5e88828901611b37565b9350506060611e6f88828901611b75565b9250506080611e8088828901611e02565b9150509295509295909350565b5f67ffffffffffffffff82169050919050565b611ea981611e8d565b82525050565b5f602082019050611ec25f830184611ea0565b92915050565b5f819050919050565b611eda81611ec8565b8114611ee4575f80fd5b50565b5f81359050611ef581611ed1565b92915050565b5f60208284031215611f1057611f0f611ac9565b5b5f611f1d84828501611ee7565b91505092915050565b5f602082019050611f395f830184611d7d565b92915050565b5f80fd5b5f80fd5b5f80fd5b5f8083601f840112611f6057611f5f611f3f565b5b8235905067ffffffffffffffff811115611f7d57611f7c611f43565b5b602083019150836001820283011115611f9957611f98611f47565b5b9250929050565b5f8060208385031215611fb657611fb5611ac9565b5b5f83013567ffffffffffffffff811115611fd357611fd2611acd565b5b611fdf85828601611f4b565b92509250509250929050565b5f604082019050611ffe5f830185611bd2565b81810360208301526120108184611a68565b90509392505050565b5f806040838503121561202f5761202e611ac9565b5b5f61203c85828601611b37565b925050602061204d85828601611b37565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061209b57607f821691505b6020821081036120ae576120ad612057565b5b50919050565b5f815190506120c281611b5f565b92915050565b5f602082840312156120dd576120dc611ac9565b5b5f6120ea848285016120b4565b91505092915050565b5f8151905061210181611db3565b92915050565b5f6020828403121561211c5761211b611ac9565b5b5f612129848285016120f3565b91505092915050565b7f496e73756666696369656e742062616c616e63650000000000000000000000005f82015250565b5f612166601483611a20565b915061217182612132565b602082019050919050565b5f6020820190508181035f8301526121938161215a565b9050919050565b7f43616e6e6f742073656e6420746f2073616d6520636861696e000000000000005f82015250565b5f6121ce601983611a20565b91506121d98261219a565b602082019050919050565b5f6020820190508181035f8301526121fb816121c2565b9050919050565b7f496e76616c696420726563697069656e740000000000000000000000000000005f82015250565b5f612236601183611a20565b915061224182612202565b602082019050919050565b5f6020820190508181035f8301526122638161222a565b9050919050565b7f416d6f756e74206d7573742062652067726561746572207468616e20300000005f82015250565b5f61229e601d83611a20565b91506122a98261226a565b602082019050919050565b5f6020820190508181035f8301526122cb81612292565b9050919050565b7f496e73756666696369656e7420666565000000000000000000000000000000005f82015250565b5f612306601083611a20565b9150612311826122d2565b602082019050919050565b5f6020820190508181035f830152612333816122fa565b9050919050565b61234381611ec8565b82525050565b61235281611b26565b82525050565b5f60e08201905061236b5f83018a61233a565b612378602083018961233a565b6123856040830188612349565b6123926060830187611bfa565b61239f6080830186611d7d565b6123ac60a083018561233a565b6123b960c0830184611bfa565b98975050505050505050565b6123ce81611ddd565b82525050565b5f81519050919050565b5f82825260208201905092915050565b5f6123f8826123d4565b61240281856123de565b9350612412818560208601611a30565b61241b81611a58565b840191505092915050565b5f819050919050565b5f61244961244461243f84612426565b611cfc565b611c72565b9050919050565b6124598161242f565b82525050565b5f6060820190506124725f8301866123c5565b818103602083015261248481856123ee565b90506124936040830184612450565b949350505050565b6124a481611e8d565b81146124ae575f80fd5b50565b5f815190506124bf8161249b565b92915050565b5f602082840312156124da576124d9611ac9565b5b5f6124e7848285016124b1565b91505092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f61252782611b56565b915061253283611b56565b925082820390508181111561254a576125496124f0565b5b92915050565b5f60c0820190506125635f830189611ea0565b6125706020830188612349565b61257d6040830187611d7d565b61258a606083018661233a565b6125976080830185612349565b6125a460a0830184611bfa565b979650505050505050565b828183375f83830152505050565b5f6125c883856123de565b93506125d58385846125af565b6125de83611a58565b840190509392505050565b5f6020820190508181035f8301526126028184866125bd565b90509392505050565b5f80fd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b61264582611a58565b810181811067ffffffffffffffff821117156126645761266361260f565b5b80604052505050565b5f612676611ac0565b9050612682828261263c565b919050565b5f80fd5b61269481611c72565b811461269e575f80fd5b50565b5f815190506126af8161268b565b92915050565b5f815190506126c381611dec565b92915050565b5f815190506126d781611ed1565b92915050565b5f80fd5b5f67ffffffffffffffff8211156126fb576126fa61260f565b5b61270482611a58565b9050602081019050919050565b5f61272361271e846126e1565b61266d565b90508281526020810184848401111561273f5761273e6126dd565b5b61274a848285611a30565b509392505050565b5f82601f83011261276657612765611f3f565b5b8151612776848260208601612711565b91505092915050565b5f67ffffffffffffffff8211156127995761279861260f565b5b602082029050602081019050919050565b5f608082840312156127bf576127be61260b565b5b6127c9608061266d565b90505f6127d8848285016126c9565b5f8301525060206127eb848285016126c9565b60208301525060406127ff848285016126a1565b6040830152506060612813848285016126a1565b60608301525092915050565b5f61283161282c8461277f565b61266d565b9050808382526020820190506080840283018581111561285457612853611f47565b5b835b8181101561287d578061286988826127aa565b845260208401935050608081019050612856565b5050509392505050565b5f82601f83011261289b5761289a611f3f565b5b81516128ab84826020860161281f565b91505092915050565b5f61016082840312156128ca576128c961260b565b5b6128d561016061266d565b90505f6128e4848285016126a1565b5f8301525060206128f7848285016126b5565b602083015250604061290b848285016126b5565b604083015250606061291f848285016120f3565b6060830152506080612933848285016126c9565b60808301525060a0612947848285016124b1565b60a08301525060c061295b848285016126a1565b60c08301525060e082015167ffffffffffffffff81111561297f5761297e612687565b5b61298b84828501612752565b60e0830152506101006129a0848285016126b5565b6101008301525061012082015167ffffffffffffffff8111156129c6576129c5612687565b5b6129d284828501612887565b610120830152506101406129e8848285016126c9565b6101408301525092915050565b6129fe81611bc7565b8114612a08575f80fd5b50565b5f81519050612a19816129f5565b92915050565b5f67ffffffffffffffff821115612a3957612a3861260f565b5b612a4282611a58565b9050602081019050919050565b5f612a61612a5c84612a1f565b61266d565b905082815260208101848484011115612a7d57612a7c6126dd565b5b612a88848285611a30565b509392505050565b5f82601f830112612aa457612aa3611f3f565b5b8151612ab4848260208601612a4f565b91505092915050565b5f805f60608486031215612ad457612ad3611ac9565b5b5f84015167ffffffffffffffff811115612af157612af0611acd565b5b612afd868287016128b4565b9350506020612b0e86828701612a0b565b925050604084015167ffffffffffffffff811115612b2f57612b2e611acd565b5b612b3b86828701612a90565b9150509250925092565b7f56414120616c72656164792070726f63657373656400000000000000000000005f82015250565b5f612b79601583611a20565b9150612b8482612b45565b602082019050919050565b5f6020820190508181035f830152612ba681612b6d565b9050919050565b612bb681611ad1565b8114612bc0575f80fd5b50565b5f612bcd82611b07565b9050919050565b5f81519050612be281612bad565b612beb81612bc3565b905092915050565b5f805f805f805f60e0888a031215612c0e57612c0d611ac9565b5b5f612c1b8a828b016126c9565b9750506020612c2c8a828b016126c9565b9650506040612c3d8a828b01612bd4565b9550506060612c4e8a828b016120b4565b9450506080612c5f8a828b016120f3565b93505060a0612c708a828b016126c9565b92505060c0612c818a828b016120b4565b91505092959891949750929550565b7f496e76616c6964206d65737361676520747970650000000000000000000000005f82015250565b5f612cc4601483611a20565b9150612ccf82612c90565b602082019050919050565b5f6020820190508181035f830152612cf181612cb8565b9050919050565b7f4e6f742074617267657420636f6e7472616374000000000000000000000000005f82015250565b5f612d2c601383611a20565b9150612d3782612cf8565b602082019050919050565b5f6020820190508181035f830152612d5981612d20565b9050919050565b5f612d6a82611b56565b9150612d7583611b56565b9250828201905080821115612d8d57612d8c6124f0565b5b92915050565b7f56414120657870697265640000000000000000000000000000000000000000005f82015250565b5f612dc7600b83611a20565b9150612dd282612d93565b602082019050919050565b5f6020820190508181035f830152612df481612dbb565b9050919050565b5f60a082019050612e0e5f83018861233a565b612e1b6020830187611d7d565b612e28604083018661233a565b612e356060830185612349565b612e426080830184611bfa565b9695505050505050565b5f606082019050612e5f5f830186612349565b612e6c6020830185611bfa565b612e796040830184611bfa565b949350505050565b5f602082019050612e945f830184612349565b9291505056fea26474726f6e58221220e8d6430ef1d438bd9ace52cb91fbf2e8cd3048d0640b94f30f64b67fd02b836364736f6c63430008160033000000000000000000000000cb0a96af34a398763c6e8e54ff1a565a40599971000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000000098968000000000000000000000000000000000000000000000000000000000000000077869616f6d696e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000026675000000000000000000000000000000000000000000000000000000000000"
	//finalBytecode := contractBytecode + constructorParams
	finalBytecode := constructorParams
	txID, err := createAndSendDeployTransactiondeploysol(
		c,
		privateKey,
		fromAddress.Hex(),
		finalBytecode,
		contractAbi,
	)
	if err != nil {
		fmt.Printf("创建交易失败: %v", err)
		return
	}
	contractAddress, blockNumber, err := waitForDeployment(c, txID)
	if err != nil {
		fmt.Printf("部署失败: %v", err)
		return
	}
	// ==================== 输出结果 ====================
	fmt.Println("\n🎉 合约部署成功!")
	fmt.Println("=====================================")
	fmt.Printf("合约地址: %s", contractAddress)
	fmt.Printf("交易ID: %s", txID)
	fmt.Printf("区块号: %d", blockNumber)
}
func main() {

	//transferTRX()

	//部署ccl合约
	/*fmt.Println("=================deploy ccl==================")
	deployCclContract()

	//部署implementContract
	fmt.Println("=================deploy implement==================")
	deployImplementContract()

	//部署setup合约
	fmt.Println("=================deploy setup==================")
	deploySetupContract()*/

	/*implementAddr := "TEYTZY2bCfnusvg5quEaPY1JSRATq9Un9R"
	setupAddr := "TXRgu7i1nJSx56waWuisV9coKAeNDJcJeR"
	tronAddressToEth(implementAddr)
	tronAddressToEth(setupAddr)*/

	//部署deploy合约，每次部署前要记得修改deploy.bin的bytecode
	//deployContractdeploy()
	
	//callvalue must be 0
	/*deployContractAddr := "TXNHc8LfPKN3gzXGECK34x6R3hJxHb2fFc"
	triggerContract(deployContractAddr)*/


	//部署token合约，每次修改前记得修改参数
	//deployContractToken()

	//balanceof
	tokenAddr:="TVh3bjZzVyLPKvQNRGN7B21LLRbUURieJ2"
	abiStr, err := readBytecodeFromFile("./abi_bin/token.abi")
	if err != nil {
		fmt.Println(err)
		return
	}
	CallConstantFunction(tokenAddr, "balanceOf", "0x70a0823100000000000000000000000019e583b06050387337824d187b00892d68426643", abiStr)
	
	//cross chain transfer
	/*tokenAddr := "TVh3bjZzVyLPKvQNRGN7B21LLRbUURieJ2"
	triggerContractSendCrossChain(tokenAddr)*/
	
	
	//tron verify vaa
	/*tokenAddr := "TVh3bjZzVyLPKvQNRGN7B21LLRbUURieJ2"
	triggerContractVAAVerfiy(tokenAddr)*/

	/*deployContractAddr := "TXNHc8LfPKN3gzXGECK34x6R3hJxHb2fFc"
	abiStr, err := readBytecodeFromFile("./abi_bin/deploy.abi")
	if err != nil {
		fmt.Println(err)
		return
	}
	CallConstantFunction(deployContractAddr, "wormholeAddress", "0xa6ea2ee8", abiStr)*/

	/*abiStr, err := readBytecodeFromFile("./abi_bin/implement.abi")
	if err!=nil{
		fmt.Println(err)
		return
	}
	CallConstantFunction("TNCJVah3c33evEmgjZahqYHtg2gk5MLXqV","getCurrentGuardianSetIndex", "0x1cfe7951",abiStr)*/

	/*abiStr, err := readBytecodeFromFile("./abi_bin/implement.abi")
	if err!=nil{
		fmt.Println(err)
		return
	}
	CallConstantFunction("TNCJVah3c33evEmgjZahqYHtg2gk5MLXqV","getGuardianSet", "0xf951975a0000000000000000000000000000000000000000000000000000000000000000",abiStr)*/

	/*implementAddr := "TZFbv24UvG8Pa3L7VQynpXajnLK9FYmTya"
	setupAddr := "TRjMWWyrYm4n4As1SZA8W65WPuBauvFet3"
	tronAddressToEth(implementAddr)
	tronAddressToEth(setupAddr)*/
	
	/*tokenContractAddr := "TVh3bjZzVyLPKvQNRGN7B21LLRbUURieJ2"
	tronAddressToEth(tokenContractAddr)*/
}

