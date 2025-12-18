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
	/*headSink := make(chan *ethTypes.Header, 2)
	headerSubscription, err := c.Connector.Client().SubscribeNewHead(ctx, headSink)
	if err != nil {
		return nil, err
	}*/
	go func() {
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

	}()
	// The purpose of this is to map events from the geth event channel to the new block event channel.
	/*common.RunWithScissors(ctx, errC, "tron_connector_subscribe_for_block", func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case ev := <-headSink:
				if ev == nil {
					c.logger.Error("new header event is nil")
					continue
				}
				if ev.Number == nil {
					c.logger.Error("new header block number is nil")
					continue
				}
				block := &NewBlock{
					Number:   ev.Number,
					Time:     ev.Time,
					Hash:     ev.Hash(),
					Finality: Finalized,
				}
				sink <- block              //nolint:channelcheck // This channel is buffered, if it backs up, we will just stop polling until it clears
				sink <- block.Copy(Safe)   //nolint:channelcheck // This channel is buffered, if it backs up, we will just stop polling until it clears
				sink <- block.Copy(Latest) //nolint:channelcheck // This channel is buffered, if it backs up, we will just stop polling until it clears
			}
		}
	})*/

	//return headerSubscription, err
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
