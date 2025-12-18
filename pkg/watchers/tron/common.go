package tron

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"

	//"github.com/ethereum/go-ethereum/rpc"
	ethCommon "github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	ethRpc "github.com/ethereum/go-ethereum/rpc"
)

type BlockMarshaller struct {
	Number *hexutil.Big
	Hash   common.Hash    `json:"hash"`
	Time   hexutil.Uint64 `json:"timestamp"`

	// L1BlockNumber is the L1 block number in which an Arbitrum batch containing this block was submitted.
	// This field is only populated when connecting to Arbitrum.
	L1BlockNumber *hexutil.Big
}

type NewBlock struct {
	Number        *big.Int
	Hash          common.Hash
	Time          uint64
	L1BlockNumber *big.Int // This is only populated on some chains (Arbitrum)
	Finality      FinalityLevel
}

func (b *NewBlock) Copy(f FinalityLevel) *NewBlock {
	return &NewBlock{
		Number:        b.Number,
		Hash:          b.Hash,
		Time:          b.Time,
		L1BlockNumber: b.L1BlockNumber,
		Finality:      f,
	}
}

// Connector exposes Wormhole-specific interactions with an EVM-based network
type Connector interface {
	NetworkName() string
	ContractAddress() address.Address
	GetCurrentGuardianSetIndex(ctx context.Context) (uint32, error)
	GetGuardianSet(ctx context.Context, index uint32) (StructsGuardianSet, error)
	//WatchLogMessagePublished(ctx context.Context, errC chan error, sink chan<- *ethabi.AbiLogMessagePublished) (event.Subscription, error)
	WatchLogMessagePublished(ctx context.Context, _ chan error, sink chan<- *AbiLogMessagePublished)
	TransactionReceipt(ctx context.Context, txHash ethCommon.Hash) (*core.TransactionInfo, error)
	TimeOfBlockByHash(ctx context.Context, hash ethCommon.Hash) (uint64, error)
	ParseLogMessagePublished(log ethTypes.Log) (*AbiLogMessagePublished, error)
	//SubscribeForBlocks(ctx context.Context, errC chan error, sink chan<- *NewBlock) (ethereum.Subscription, error)
	SubscribeForBlocks(ctx context.Context, errC chan error, sink chan<- *NewBlock)
	GetLatest(ctx context.Context) (latest, finalized, safe uint64, err error)
	RawCallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error
	RawBatchCallContext(ctx context.Context, b []ethRpc.BatchElem) error
	Client() *client.GrpcClient
	SubscribeNewHead(ctx context.Context, ch chan<- *NewBlock) (ethereum.Subscription, error)
}
