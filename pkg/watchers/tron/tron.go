package tron

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fbsobreira/gotron-sdk/pkg/address"

	//ethAbi "github.com/certusone/wormhole/node/pkg/watchers/evm/connectors/ethabi"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/fbsobreira/gotron-sdk/pkg/common"

	"github.com/ethereum/go-ethereum"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	ethRpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"go.uber.org/zap"
)

type TronBaseConnector struct {
	networkId       string
	contractAddress string
	fromAddress     string
	logger          *zap.Logger
	client          *client.GrpcClient
	abi             *abi.ABI
}

func NewTronBaseConnector(networkId, connectUrl, address string, logger *zap.Logger) (*TronBaseConnector, error) {
	// 创建gRPC连接
	url := strings.Split(connectUrl, "//")
	if len(url) != 2 {
		return nil, errors.New("tronRPC ERROR")
	}
	grpcClient := client.NewGrpcClient(url[1])
	err := grpcClient.Start(client.GRPCInsecure())
	if err != nil {
		return nil, err
	}
	abiJson, err := abi.JSON(strings.NewReader(AbiABI))
	if err != nil {
		return nil, err
	}
	fromAddress := "410000000000000000000000000000000000000000" //设置为默认值
	return &TronBaseConnector{
		networkId:       networkId,
		contractAddress: address,
		fromAddress:     fromAddress,
		logger:          logger,
		client:          grpcClient,
		abi:             &abiJson,
	}, nil
}

func (tc *TronBaseConnector) CallContract(methodName string, parameters []interface{}) (interface{}, error) {
	// 验证参数
	if methodName == "" {
		return nil, errors.New("methodName can not be empty")
	}
	// 编码函数调用数据
	encodedData, err := EncodeFunctionCall(methodName, parameters, tc.abi)
	if err != nil {
		return nil, fmt.Errorf("encodeFunctionCall error: %v", err)
	}

	// 创建调用参数
	contractAddressBytes, err := common.DecodeCheck(tc.contractAddress)
	if err != nil {
		return nil, fmt.Errorf("decodecheck contract error: %v", err)
	}
	callerAddressBytes, err := common.HexStringToBytes(tc.fromAddress)
	if err != nil {
		return nil, fmt.Errorf("decodecheck fromAddress error: %v", err)
	}

	triggerContract := &core.TriggerSmartContract{
		OwnerAddress:    callerAddressBytes,
		ContractAddress: contractAddressBytes,
		Data:            encodedData,
		CallValue:       0, //默认为0
		TokenId:         0, //默认为0
		CallTokenValue:  0, //默认为0
	}

	transactionExtension, err := tc.client.Client.TriggerConstantContract(context.Background(), triggerContract)
	if err != nil {
		return nil, fmt.Errorf("triggerConstantContract error: %v", err)
	}

	// 解析结果
	return ParseCallResult(transactionExtension, methodName, tc.abi)
}

func (tc *TronBaseConnector) NetworkName() string {
	return tc.networkId
}

func (tc *TronBaseConnector) ContractAddress() address.Address {
	bytes, err := common.DecodeCheck(tc.contractAddress)
	if err != nil {
		panic(err)
	}
	addrHex := hex.EncodeToString(bytes)
	addr := address.HexToAddress(addrHex)
	return addr
}

// function getCurrentGuardianSetIndex() view returns(uint32)
func (tc *TronBaseConnector) GetCurrentGuardianSetIndex(ctx context.Context) (uint32, error) {
	guardianIndexArr := make([]interface{}, 0)
	data, err := tc.CallContract("getCurrentGuardianSetIndex", guardianIndexArr)
	if err != nil {
		return 0, err
	}
	value, ok := data.(uint32)
	if ok {
		return value, nil
	}
	return 0, errors.New("获取guardianSet Index failed")
}

// function getGuardianSet(uint32 index) view returns((address[],uint32))
type StructsGuardianSet struct {
	Keys           []ethCommon.Address
	ExpirationTime uint32
}

func (tc *TronBaseConnector) GetGuardianSet(ctx context.Context, index uint32) (StructsGuardianSet, error) {
	guardianIndexArr := make([]interface{}, 0)
	guardianIndexArr = append(guardianIndexArr, index)
	structsGuardianSet := new(StructsGuardianSet)
	data, err := tc.CallContract("getGuardianSet", guardianIndexArr)
	if err != nil {
		return *structsGuardianSet, err
	}

	dataByte, err := json.Marshal(data)
	if err != nil {
		return *structsGuardianSet, err
	}
	err = json.Unmarshal(dataByte, structsGuardianSet)
	if err != nil {
		return *structsGuardianSet, err
	}
	return *structsGuardianSet, nil
}

type AbiLogMessagePublished struct {
	Sender           ethCommon.Address
	Sequence         uint64
	Nonce            uint32
	Payload          []byte
	ConsistencyLevel uint8
	Raw              types.Log // Blockchain specific contextual infos
}

func (tc *TronBaseConnector) WatchLogMessagePublished(ctx context.Context, _ chan error, sink chan<- *AbiLogMessagePublished) {
	contractClient := NewTronContractClient(
		tc.contractAddress,
		tc.client,
		tc.logger,
	)

	//ctx := context.Background()
	filter := &TronEventFilter{
		ContractAddress: tc.contractAddress,
		EventName:       "LogMessagePublished",
		StartBlock:      50000000,
		Topics:          []byte{221, 242, 82, 173, 27, 226, 200, 155, 105, 194, 176, 104, 252, 55, 141, 170, 149, 43, 167, 241, 99, 196, 161, 22, 40, 245, 90, 77, 245, 35, 179, 239},
	}

	//eventChan := make(chan *ContractEvent, 100)
	logChan := make(chan *types.Log, 100)

	go func() {
		err := contractClient.WatchEvents1(ctx, filter, logChan, tc.abi)
		if err != nil {
			tc.logger.Error("Failed to watch events", zap.Error(err))
			close(logChan)
			return
		}
		close(logChan)
	}()
	for log := range logChan {
		abiLogMessagePublish := new(AbiLogMessagePublished)
		event := "LogMessagePublished"
		_, exists := tc.abi.Events[event]
		if !exists {
			tc.logger.Error("LogMessagePublished不存在ABI合约中")
			panic(errors.New("LogMessagePublished不存在ABI合约中"))
		}
		if len(log.Topics) == 0 {
			panic("errNoEventSignature")
		}
		if log.Topics[0] != tc.abi.Events[event].ID {
			panic("errEventSignatureMismatch")
		}
		if len(log.Data) > 0 {
			if err := tc.abi.UnpackIntoInterface(abiLogMessagePublish, event, log.Data); err != nil {
				panic(err)
			}
		}
		abiLogMessagePublish.Raw.BlockNumber = log.BlockNumber
		var indexed abi.Arguments
		for _, arg := range tc.abi.Events[event].Inputs {
			if arg.Indexed {
				indexed = append(indexed, arg)
			}
		}
		abi.ParseTopics(abiLogMessagePublish, indexed, log.Topics[1:])
		sink <- abiLogMessagePublish

	}

}

func (tc *TronBaseConnector) TransactionReceipt(ctx context.Context, txHash ethCommon.Hash) (*core.TransactionInfo, error) {
	tx, err := tc.client.GetTransactionInfoByID(txHash.Hex())
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (tc *TronBaseConnector) TimeOfBlockByHash(ctx context.Context, hash ethCommon.Hash) (uint64, error) {
	block, err := tc.client.GetBlockByID(hash.Hex())
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	return uint64(block.GetBlockHeader().GetRawData().Timestamp), err
}

func parseLogMessagePublicshed(log ethTypes.Log, abi1 *abi.ABI) (*AbiLogMessagePublished, error) {
	if len(log.Topics) == 0 {
		return nil, errors.New("log.topics can not be empty")
	}
	event := new(AbiLogMessagePublished)
	if log.Topics[0] != abi1.Events["LogMessagePublished"].ID {
		return nil, errors.New("log.topics[0] is not same with events[LogMessagePublished]")
	}
	if len(log.Data) > 0 {
		if err := abi1.UnpackIntoInterface(event, "LogMessagePublished", log.Data); err != nil {
			return nil, err
		}
	}
	var indexed abi.Arguments
	for _, arg := range abi1.Events["LogMessagePublished"].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	err := abi.ParseTopics(event, indexed, log.Topics[1:])
	if err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

func (tc *TronBaseConnector) ParseLogMessagePublished(log ethTypes.Log) (*AbiLogMessagePublished, error) {
	return parseLogMessagePublicshed(log, tc.abi)
}

func (tc *TronBaseConnector) SubscribeForBlocks1(ctx context.Context, errC chan error, sink chan<- *NewBlock) (ethereum.Subscription, error) {
	panic("not implemented")
}

func (c *TronBaseConnector) SubscribeForBlocks(ctx context.Context, errC chan error, sink chan<- *NewBlock) {

}

func (tc *TronBaseConnector) GetLatest(ctx context.Context) (latest, finalized, safe uint64, err error) {
	panic("not implemented")
}

func (tc *TronBaseConnector) RawCallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	/*value, ok := args[0].(int64)
	if !ok {
		return errors.New("获取区块高度参数错误")

	}
	blockExtension, err := tc.client.GetBlockByNum(value)*/
	blockExtension, err := tc.client.GetNowBlock()
	if err != nil {
		return err
	}
	result = blockExtension
	return nil

}

func (e *TronBaseConnector) RawBatchCallContext(ctx context.Context, b []ethRpc.BatchElem) error {
	//todo
	return nil
}

func (tc *TronBaseConnector) Client() *client.GrpcClient {
	return tc.client
}

func (tc *TronBaseConnector) SubscribeNewHead(ctx context.Context, ch chan<- *NewBlock) (ethereum.Subscription, error) {
	//todo
	return nil, nil
}
