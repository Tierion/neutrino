package neutrino

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"sync"
)


// Mempool is used when we are downloading unconfirmed transactions.
// We will use this object to track which transactions we've already
// downloaded so that we don't download them more than once.
type Mempool struct {
	downloadedTxs map[chainhash.Hash]bool
	mtx           sync.RWMutex
	callbacks     []func(tx *btcutil.Tx, block *btcjson.BlockDetails)
	watchedAddrs  []btcutil.Address
}

// NewMempool returns an initialized Mempool
func NewMempool() *Mempool {
	return &Mempool{
		downloadedTxs: make(map[chainhash.Hash]bool),
		mtx:           sync.RWMutex{},
	}
}

// RegisterCallback will register a callback that will fire when a transaction
// matching a watched address enters the Mempool.
func (mp *Mempool) RegisterCallback(onRecvTx func(tx *btcutil.Tx, block *btcjson.BlockDetails)) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	mp.callbacks = append(mp.callbacks, onRecvTx)
}

// HaveTransaction returns whether or not the passed transaction already exists
// in the Mempool.
func (mp *Mempool) HaveTransaction(hash *chainhash.Hash) bool {
	mp.mtx.RLock()
	defer mp.mtx.RUnlock()
	return mp.downloadedTxs[*hash]
}

// AddTransaction adds a new transaction to the Mempool and
// maybe calls back if it matches any watched addresses.
func (mp *Mempool) AddTransaction(tx *btcutil.Tx) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	mp.downloadedTxs[*tx.Hash()] = true

	ro := defaultRescanOptions()
	WatchAddrs(mp.watchedAddrs...)(ro)
	if ok, err := ro.paysWatchedAddr(tx); ok && err == nil {
		for _, cb := range mp.callbacks {
			cb(tx, nil)
		}
	}
}

// Clear will remove all transactions from the Mempool. This
// should be done whenever a new block is accepted.
func (mp *Mempool) Clear() {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	mp.downloadedTxs = make(map[chainhash.Hash]bool)
}

// NotifyReceived stores addresses to watch
func (mp *Mempool) NotifyReceived(addrs []btcutil.Address) {
	mp.mtx.Lock()
	defer mp.mtx.Unlock()
	mp.watchedAddrs = append(mp.watchedAddrs, addrs...)
}