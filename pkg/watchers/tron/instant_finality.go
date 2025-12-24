package tron

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

// InstantFinalityConnector is used for chains that support instant finality. It uses the standard geth head sink to read blocks
// and publishes each block as latest, safe and finalized.
type InstantFinalityConnector struct {
	Connector
	logger *zap.Logger
}

func NewInstantFinalityConnector(baseConnector *TronBaseConnector, logger *zap.Logger) *InstantFinalityConnector {
	connector := &InstantFinalityConnector{
		Connector: baseConnector,
		logger:    logger,
	}
	return connector
}

func (c *InstantFinalityConnector) SubscribeForBlocks(ctx context.Context, errC chan error, sink chan<- *NewBlock) {
	for {
		blockExtension, err := c.Connector.Client().GetNowBlock()
		if err != nil {
			c.logger.Error(err.Error())
			errC <- errors.New(err.Error())
			continue
		}
		newBlock := new(NewBlock)
		newBlock.Number = big.NewInt(blockExtension.BlockHeader.RawData.Number)
		newBlock.Hash = ethCommon.BytesToHash(blockExtension.Blockid)
		newBlock.Time = uint64(blockExtension.BlockHeader.RawData.Timestamp)
		newBlock.Finality = Latest
		sink <- newBlock
		sink <- newBlock.Copy(Safe)
		sink <- newBlock.Copy(Latest)
		time.Sleep(3000)
	}
}

func (c *InstantFinalityConnector) GetLatest(ctx context.Context) (latest, finalized, safe uint64, err error) {
	latestBlock, err := GetBlockByFinality(ctx, c.Connector, Latest)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get latest block: %w", err)
	}

	latest = latestBlock.Number.Uint64()
	finalized = latest
	safe = latest
	return
}
