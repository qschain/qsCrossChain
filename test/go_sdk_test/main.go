package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strings"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client/transaction"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/keys"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"go.uber.org/zap"
	"time"
)

// ABI
// const AbiABI = "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousAdmin\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAdmin\",\"type\":\"address\"}],\"name\":\"AdminChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"beacon\",\"type\":\"address\"}],\"name\":\"BeaconUpgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"oldContract\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newContract\",\"type\":\"address\"}],\"name\":\"ContractUpgraded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint32\",\"name\":\"index\",\"type\":\"uint32\"}],\"name\":\"GuardianSetAdded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"sequence\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"nonce\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"consistencyLevel\",\"type\":\"uint8\"}],\"name\":\"LogMessagePublished\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"implementation\",\"type\":\"address\"}],\"name\":\"Upgraded\",\"type\":\"event\"},{\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"inputs\":[],\"name\":\"chainId\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCurrentGuardianSetIndex\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"index\",\"type\":\"uint32\"}],\"name\":\"getGuardianSet\",\"outputs\":[{\"components\":[{\"internalType\":\"address[]\",\"name\":\"keys\",\"type\":\"address[]\"},{\"internalType\":\"uint32\",\"name\":\"expirationTime\",\"type\":\"uint32\"}],\"internalType\":\"structStructs.GuardianSet\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getGuardianSetExpiry\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"governanceActionIsConsumed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"governanceChainId\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"governanceContract\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"impl\",\"type\":\"address\"}],\"name\":\"isInitialized\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"messageFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"emitter\",\"type\":\"address\"}],\"name\":\"nextSequence\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedVM\",\"type\":\"bytes\"}],\"name\":\"parseAndVerifyVM\",\"outputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"timestamp\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"nonce\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"emitterChainId\",\"type\":\"uint16\"},{\"internalType\":\"bytes32\",\"name\":\"emitterAddress\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"sequence\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"consistencyLevel\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"guardianSetIndex\",\"type\":\"uint32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"guardianIndex\",\"type\":\"uint8\"}],\"internalType\":\"structStructs.Signature[]\",\"name\":\"signatures\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"internalType\":\"structStructs.VM\",\"name\":\"vm\",\"type\":\"tuple\"},{\"internalType\":\"bool\",\"name\":\"valid\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedUpgrade\",\"type\":\"bytes\"}],\"name\":\"parseContractUpgrade\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"module\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"action\",\"type\":\"uint8\"},{\"internalType\":\"uint16\",\"name\":\"chain\",\"type\":\"uint16\"},{\"internalType\":\"address\",\"name\":\"newContract\",\"type\":\"address\"}],\"internalType\":\"structGovernanceStructs.ContractUpgrade\",\"name\":\"cu\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedUpgrade\",\"type\":\"bytes\"}],\"name\":\"parseGuardianSetUpgrade\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"module\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"action\",\"type\":\"uint8\"},{\"internalType\":\"uint16\",\"name\":\"chain\",\"type\":\"uint16\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"keys\",\"type\":\"address[]\"},{\"internalType\":\"uint32\",\"name\":\"expirationTime\",\"type\":\"uint32\"}],\"internalType\":\"structStructs.GuardianSet\",\"name\":\"newGuardianSet\",\"type\":\"tuple\"},{\"internalType\":\"uint32\",\"name\":\"newGuardianSetIndex\",\"type\":\"uint32\"}],\"internalType\":\"structGovernanceStructs.GuardianSetUpgrade\",\"name\":\"gsu\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedSetMessageFee\",\"type\":\"bytes\"}],\"name\":\"parseSetMessageFee\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"module\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"action\",\"type\":\"uint8\"},{\"internalType\":\"uint16\",\"name\":\"chain\",\"type\":\"uint16\"},{\"internalType\":\"uint256\",\"name\":\"messageFee\",\"type\":\"uint256\"}],\"internalType\":\"structGovernanceStructs.SetMessageFee\",\"name\":\"smf\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedTransferFees\",\"type\":\"bytes\"}],\"name\":\"parseTransferFees\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"module\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"action\",\"type\":\"uint8\"},{\"internalType\":\"uint16\",\"name\":\"chain\",\"type\":\"uint16\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"recipient\",\"type\":\"bytes32\"}],\"internalType\":\"structGovernanceStructs.TransferFees\",\"name\":\"tf\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"encodedVM\",\"type\":\"bytes\"}],\"name\":\"parseVM\",\"outputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"timestamp\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"nonce\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"emitterChainId\",\"type\":\"uint16\"},{\"internalType\":\"bytes32\",\"name\":\"emitterAddress\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"sequence\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"consistencyLevel\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"guardianSetIndex\",\"type\":\"uint32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"guardianIndex\",\"type\":\"uint8\"}],\"internalType\":\"structStructs.Signature[]\",\"name\":\"signatures\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"internalType\":\"structStructs.VM\",\"name\":\"vm\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_vm\",\"type\":\"bytes\"}],\"name\":\"submitContractUpgrade\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_vm\",\"type\":\"bytes\"}],\"name\":\"submitNewGuardianSet\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_vm\",\"type\":\"bytes\"}],\"name\":\"submitSetMessageFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_vm\",\"type\":\"bytes\"}],\"name\":\"submitTransferFees\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"guardianIndex\",\"type\":\"uint8\"}],\"internalType\":\"structStructs.Signature[]\",\"name\":\"signatures\",\"type\":\"tuple[]\"},{\"components\":[{\"internalType\":\"address[]\",\"name\":\"keys\",\"type\":\"address[]\"},{\"internalType\":\"uint32\",\"name\":\"expirationTime\",\"type\":\"uint32\"}],\"internalType\":\"structStructs.GuardianSet\",\"name\":\"guardianSet\",\"type\":\"tuple\"}],\"name\":\"verifySignatures\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"valid\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"timestamp\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"nonce\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"emitterChainId\",\"type\":\"uint16\"},{\"internalType\":\"bytes32\",\"name\":\"emitterAddress\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"sequence\",\"type\":\"uint64\"},{\"internalType\":\"uint8\",\"name\":\"consistencyLevel\",\"type\":\"uint8\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"guardianSetIndex\",\"type\":\"uint32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"},{\"internalType\":\"uint8\",\"name\":\"guardianIndex\",\"type\":\"uint8\"}],\"internalType\":\"structStructs.Signature[]\",\"name\":\"signatures\",\"type\":\"tuple[]\"},{\"internalType\":\"bytes32\",\"name\":\"hash\",\"type\":\"bytes32\"}],\"internalType\":\"structStructs.VM\",\"name\":\"vm\",\"type\":\"tuple\"}],\"name\":\"verifyVM\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"valid\",\"type\":\"bool\"},{\"internalType\":\"string\",\"name\":\"reason\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"nonce\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"payload\",\"type\":\"bytes\"},{\"internalType\":\"uint8\",\"name\":\"consistencyLevel\",\"type\":\"uint8\"}],\"name\":\"publishMessage\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"sequence\",\"type\":\"uint64\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"initialGuardians\",\"type\":\"address[]\"},{\"internalType\":\"uint16\",\"name\":\"chainId\",\"type\":\"uint16\"},{\"internalType\":\"uint16\",\"name\":\"governanceChainId\",\"type\":\"uint16\"},{\"internalType\":\"bytes32\",\"name\":\"governanceContract\",\"type\":\"bytes32\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"
// 测试ABI
//const AbiABI = "[{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"initialValue\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"}],\"name\":\"DataStored\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"get\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"}],\"name\":\"store\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]"

const AbiABI = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"}],\"name\":\"DataStored\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"get111\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"}],\"name\":\"store111\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"storedData\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

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
	Topics          [][]byte // 事件主题
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
	/*parsedABI, err := abi.JSON(strings.NewReader(AbiABI))
	if err != nil {
		return nil, err
	}*/
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
			fmt.Println("==========是目标合约交易==================")
			if len(txInfo.Log) > 0 {
				fmt.Println(txInfo.Log)
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

// WatchEvents 监听合约事件（持续监听）
func (c *TronContractClient) WatchEvents(ctx context.Context, filter *TronEventFilter, eventChan chan<- *ContractEvent) error {
	fmt.Println("=============watchevents======================")
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
		ethLog, err := c.convertTronLogToEthLog(log)
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
func (c *TronContractClient) convertTronLogToEthLog(tronLog *core.TransactionInfo_Log) (*types.Log, error) {
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
	fmt.Println("=====eventSigHash=========")
	fmt.Println(eventSigHash.Hex())
	// 根据事件签名哈希查找对应的事件
	var eventABI abi.Event
	found := false

	for _, event := range parsedABI.Events {
		// 计算事件签名的哈希
		fmt.Println(event.Name)
		fmt.Println(event.Inputs)
		//eventSig := fmt.Sprintf("%s(%s)", event.Name, event.Inputs)
		eventSig := BuildEventSignature(event)
		calculatedSigHash := crypto.Keccak256Hash([]byte(eventSig))
		fmt.Println("=======calculatedsighash========================")
		fmt.Println(calculatedSigHash)
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
	if filter.Topics != nil {
		if !c.matchTopics(ethLog.Topics[1:], filter.Topics) {
			return nil, fmt.Errorf("topics do not match filter")
		}
	}

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
	// 检查长度是否为20字节
	if len(addressBytes) != 21 {
		return "", fmt.Errorf("expected 20-byte address, got %d bytes", len(addressBytes))
	}

	// 添加波场主网前缀 (0x41)
	/*prefixed := make([]byte, 21)
	prefixed[0] = 0x41 // 主网前缀
	copy(prefixed[1:], addressBytes)*/

	// 使用波场SDK的Base58Check编码
	return common.EncodeCheck(addressBytes), nil
}

// ContractAddressToBase58 从TransactionInfo获取合约地址并转换为Base58
func ContractAddressToBase58(txInfo *core.TransactionInfo) (string, error) {
	if txInfo == nil {
		return "", fmt.Errorf("transaction info is nil")
	}

	// 检查是否是21字节地址
	if len(txInfo.ContractAddress) != 21 {
		return "", fmt.Errorf("contract address is not 21 bytes, got %d bytes", len(txInfo.ContractAddress))
	}

	return BytesToTronBase58(txInfo.ContractAddress)
}

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

// encodeConstructorArgs 编码构造函数参数
func encodeConstructorArgs(initialValue int64) (string, error) {
	// 将 int64 转换为 big.Int
	value := big.NewInt(initialValue)

	// 将 uint256 编码为 64 个字符的16进制字符串（32字节）
	encoded := common.LeftPadBytes(value.Bytes(), 32)
	encodedHex := common.Bytes2Hex(encoded)

	log.Printf("   构造函数参数编码: %s (初始值: %d)", encodedHex, initialValue)
	return encodedHex, nil
}

// createAndSendDeployTransaction 创建并发送部署交易
func createAndSendDeployTransaction(
	grpcClient *client.GrpcClient,
	privateKey *btcec.PrivateKey,
	fromAddr string,
	bytecode string,
) (string, error) {
	// 转换为波场地址格式
	fromTronAddr := address.HexToAddress(fromAddr)

	log.Printf("   创建合约部署交易...")
	log.Printf("   部署者: %s", fromTronAddr.String())
	log.Printf("   字节码长度: %d 字节", len(bytecode)/2)

	// 创建智能合约部署交易
	bytecodeHex, _ := common.Hex2Bytes(bytecode)
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
			OriginAddress:   fromTronAddr.Bytes(),
			ContractAddress: []byte{},
			Abi:             &core.SmartContract_ABI{Entrys: []*core.SmartContract_ABI_Entry{}},
			Bytecode:        bytecodeHex,
			CallValue:       0,
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
	txExtension.Transaction.RawData.FeeLimit = 100_000_000
	tx := txExtension.Transaction

	signedTx, err := transaction.SignTransaction(tx, privateKey)
	if err != nil {
		return "", fmt.Errorf("签名交易失败: %v", err)
	}
	txID := common.BytesToHexString(txExtension.GetTxid())
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
func deployContract() {
	c := client.NewGrpcClient("localhost:50051")
	err := c.Start(client.GRPCInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Stop()
	privateKeyHex1 := "6c1505933bb9d95b85134734aa5286c88aa076406c3c6f556bb5a0f2b1bb4b53"
	bytecodeFile := "./SimpleStorage.bin"
	initialValue := int64(42)
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
	// 编码构造函数参数 (uint256)
	encodedArgs, err := encodeConstructorArgs(initialValue)
	if err != nil {
		fmt.Printf("编码构造函数参数失败: %v", err)
		return
	}
	fmt.Println("==========encodedArgs=========================")
	fmt.Println(encodedArgs)
	// 合并字节码和构造函数参数
	finalBytecode := contractBytecode + encodedArgs

	txID, err := createAndSendDeployTransaction(
		c,
		privateKey,
		fromAddress.Hex(),
		finalBytecode,
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
func getWitnesses() {
	c := client.NewGrpcClient("localhost:50051")
	err := c.Start(client.GRPCInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Stop()
	block, err := c.Client.GetNowBlock(context.Background(), &api.EmptyMessage{})
	if err != nil {
		fmt.Println(err)
		return
	}
	latestNum := block.BlockHeader.RawData.Number
	for i := 0; i < int(latestNum); i++ {
		blockExten, err := c.GetBlockByNum(int64(i))
		if err != nil {
			fmt.Println(err)
			continue
		}
		witAddr := blockExten.BlockHeader.RawData.WitnessAddress
		witAddrStr := common.EncodeCheck(witAddr)
		fmt.Print(i)
		fmt.Print("    ")
		fmt.Println(witAddrStr)
	}

}
func filter() {
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

func main() {
	//hexStr := "0a020005220873b49021c3a2f02540d4abd5b9b0335a8e01081f1289010a31747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e54726967676572536d617274436f6e747261637412540a1541d9215ac939d80cac35d201ee8d550cc03c3219b012154122e57b8df38ffe9273aa33e2198b9ef414946ed82224c8e25f62000000000000000000000000000000000000000000000000000000000000006470d1fdaeafb03390018094ebdc03"
	//getTransactionById()

	//transferTRX()
	//getWitnesses()
	//deployContract()
	filter()
}

