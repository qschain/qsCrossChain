package tron

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
)

func GetLatestBlock(ctx context.Context, conn Connector) (*NewBlock, error) {
	return GetBlockByFinality(ctx, conn, Latest)
}

func GetBlockByFinality(ctx context.Context, conn Connector, blockFinality FinalityLevel) (*NewBlock, error) {
	return GetBlock(ctx, conn, blockFinality.String(), blockFinality)
}

func GetBlockByNumberUint64(ctx context.Context, conn Connector, blockNum uint64, blockFinality FinalityLevel) (*NewBlock, error) {
	return GetBlock(ctx, conn, "0x"+fmt.Sprintf("%x", blockNum), blockFinality)
}

func GetBlock(ctx context.Context, conn Connector, str string, blockFinality FinalityLevel) (*NewBlock, error) {
	timeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var m api.BlockExtention
	err := conn.RawCallContext(timeout, &m, "eth_getBlockByNumber", str, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get block for %s: %w", str, err)
	}
	if m.GetBlockHeader().GetRawData().Number == 0 {
		return nil, fmt.Errorf("failed to unmarshal block for %s: Number is nil", str)
	}
	n := m.GetBlockHeader().GetRawData().GetNumber()
	return &NewBlock{
		Number:        big.NewInt(n),
		Time:          uint64(m.BlockHeader.RawData.Timestamp),
		Hash:          common.BytesToHash(m.Blockid),
		L1BlockNumber: nil,
		Finality:      Latest,
	}, nil
}
