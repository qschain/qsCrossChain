package tron

import (
	"context"
	"fmt"
	"time"

	"github.com/certusone/wormhole/node/pkg/watchers/evm"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"

	"github.com/certusone/wormhole/node/pkg/common"
	eth_common "github.com/ethereum/go-ethereum/common"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

var (
	// SECURITY: Hardcoded ABI identifier for the LogMessagePublished topic. When using the watcher, we don't need this
	// since the node will only hand us pre-filtered events. In this case, we need to manually verify it
	// since ParseLogMessagePublished will only verify whether it parses.
	LogMessagePublishedTopic = eth_common.HexToHash("0x6eb224fb001ed210e379b335e35efe88672a8ce935d981a6896b27ffdf52a3b2")
)

// MessageEventsForTransaction returns the lockup events for a given transaction.
// Returns the block number and a list of MessagePublication events.
func MessageEventsForTransaction(
	ctx context.Context,
	tronConn Connector,
	contract string,
	chainId vaa.ChainID,
	tx eth_common.Hash) (*core.TransactionInfo, uint64, []*common.MessagePublication, error) {

	// Get transactions logs from transaction
	receipt, err := tronConn.TransactionReceipt(ctx, tx)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	// Bail early when the transaction receipt status is anything other than
	// 1 (success). In theory, this check isn't strictly necessary - a failed
	// transaction cannot emit logs and will trigger neither subscription
	// messages nor have log messages in its receipt.
	//
	// However, relying on that invariant is brittle - we connect to a lot of
	// EVM-compatible chains which might accidentally break this API contract
	// and return logs for failed transactions. Check explicitly instead.
	if receipt.Result == 1 {
		return nil, 0, nil, fmt.Errorf("non-success transaction status: %d", receipt.Result)
	}

	// Get block
	blockTime, err := tronConn.TimeOfBlockByHash(ctx, tx)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to get block time: %w", err)
	}

	msgs := make([]*common.MessagePublication, 0, len(receipt.Log))

	// Extract logs
	for _, l := range receipt.Log {
		if l == nil {
			continue
		}
		ethLog, err := ConvertTronLogToEthLog(l)
		if err != nil {
			continue
		}
		// SECURITY: Skip logs not produced by our contract.
		contractToEthAddrStr, err := TronAddressToEth(contract)
		if err != nil {
			continue
		}
		contractToEthAddr := eth_common.HexToAddress(contractToEthAddrStr)

		if ethLog.Address != contractToEthAddr {
			continue
		}

		if ethLog.Topics[0] != LogMessagePublishedTopic {
			continue
		}

		ev, err := tronConn.ParseLogMessagePublished(*ethLog)
		if err != nil {
			return nil, 0, nil, fmt.Errorf("failed to parse log: %w", err)
		}

		message := &common.MessagePublication{
			TxID:             ev.Raw.TxHash.Bytes(),
			Timestamp:        time.Unix(int64(blockTime), 0), // #nosec G115 -- This conversion is safe indefinitely
			Nonce:            ev.Nonce,
			Sequence:         ev.Sequence,
			EmitterChain:     chainId,
			EmitterAddress:   evm.PadAddress(ev.Sender),
			Payload:          ev.Payload,
			ConsistencyLevel: ev.ConsistencyLevel,
		}

		msgs = append(msgs, message)
	}

	return receipt, uint64(receipt.BlockNumber), msgs, nil
}
