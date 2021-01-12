package block

import (
	"context"
	"log"

	"github.com/gookit/color"
	d "github.com/itzmeanjan/ette/app/data"
	"github.com/itzmeanjan/ette/app/db"
)

// PublishBlock - Attempts to publish block data to Redis pubsub channel
func PublishBlock(block *db.PackedBlock, redis *d.RedisInfo) {

	if err := redis.Client.Publish(context.Background(), "block", &d.Block{
		Hash:                block.Block.Hash,
		Number:              block.Block.Number,
		Time:                block.Block.Time,
		ParentHash:          block.Block.ParentHash,
		Difficulty:          block.Block.Difficulty,
		GasUsed:             block.Block.GasUsed,
		GasLimit:            block.Block.GasLimit,
		Nonce:               block.Block.Nonce,
		Miner:               block.Block.Miner,
		Size:                block.Block.Size,
		TransactionRootHash: block.Block.TransactionRootHash,
		ReceiptRootHash:     block.Block.ReceiptRootHash,
	}).Err(); err != nil {
		log.Print(color.Red.Sprintf("[!] Failed to publish block %d in channel : %s", block.Block.Number, err.Error()))
		return
	}

	PublishTxs(block.Block.Number, block.Transactions, redis)

}

// PublishTxs - Publishes all transactions in a block to redis pubsub
// channel
func PublishTxs(blockNumber uint64, txs []*db.PackedTransaction, redis *d.RedisInfo) {

	if txs == nil {
		return
	}

	for _, t := range txs {
		PublishTx(blockNumber, t, redis)
	}

}

// PublishTx - Publishes tx & events in tx, related data to respective
// Redis pubsub channel
func PublishTx(blockNumber uint64, tx *db.PackedTransaction, redis *d.RedisInfo) {

	if tx == nil {
		return
	}

	var pTx *d.Transaction

	if tx.Tx.To == "" {
		// This is a contract creation tx
		pTx = &d.Transaction{
			Hash:      tx.Tx.Hash,
			From:      tx.Tx.From,
			Contract:  tx.Tx.Contract,
			Value:     tx.Tx.Value,
			Data:      tx.Tx.Data,
			Gas:       tx.Tx.Gas,
			GasPrice:  tx.Tx.GasPrice,
			Cost:      tx.Tx.Cost,
			Nonce:     tx.Tx.Nonce,
			State:     tx.Tx.State,
			BlockHash: tx.Tx.BlockHash,
		}
	} else {
		// This is a normal tx, so we keep contract field empty
		pTx = &d.Transaction{
			Hash:      tx.Tx.Hash,
			From:      tx.Tx.From,
			To:        tx.Tx.To,
			Value:     tx.Tx.Value,
			Data:      tx.Tx.Data,
			Gas:       tx.Tx.Gas,
			GasPrice:  tx.Tx.GasPrice,
			Cost:      tx.Tx.Cost,
			Nonce:     tx.Tx.Nonce,
			State:     tx.Tx.State,
			BlockHash: tx.Tx.BlockHash,
		}
	}

	if err := redis.Client.Publish(context.Background(), "transaction", pTx).Err(); err != nil {
		log.Print(color.Red.Sprintf("[!] Failed to publish transaction from block %d : %s", blockNumber, err.Error()))
		return
	}

	PublishEvents(blockNumber, tx.Events, redis)

}

// PublishEvents - Iterate over all events & try to publish them on
// redis pubsub channel
func PublishEvents(blockNumber uint64, events []*db.Events, redis *d.RedisInfo) {

	if events == nil {
		return
	}

	for _, e := range events {
		PublishEvent(blockNumber, e, redis)
	}

}

// PublishEvent - Publishing event/ log entry to redis pub-sub topic, to be captured by subscribers
// and sent to client application, who are interested in this piece of data
// after applying filter
func PublishEvent(blockNumber uint64, event *db.Events, redis *d.RedisInfo) {

	if event == nil {
		return
	}

	if err := redis.Client.Publish(context.Background(), "event", &d.Event{
		Origin:          event.Origin,
		Index:           event.Index,
		Topics:          event.Topics,
		Data:            event.Data,
		TransactionHash: event.TransactionHash,
		BlockHash:       event.BlockHash,
	}).Err(); err != nil {
		log.Print(color.Red.Sprintf("[!] Failed to publish event from block %d : %s", blockNumber, err.Error()))
		return
	}

}