package protocol

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/groupcache/lru"
	log "github.com/sirupsen/logrus"

	"btmSign/bytom/consensus"
	"btmSign/bytom/consensus/bcrp"
	"btmSign/bytom/event"
	"btmSign/bytom/protocol/bc"
	"btmSign/bytom/protocol/bc/types"
	"btmSign/bytom/protocol/state"
)

// msg type
const (
	MsgNewTx = iota
	MsgRemoveTx
	logModule = "protocol"
)

var (
	maxCachedErrTxs = 1000
	maxMsgChSize    = 1000
	maxNewTxNum     = 10000
	maxOrphanNum    = 2000

	orphanTTL                = 10 * time.Minute
	orphanExpireScanInterval = 3 * time.Minute

	// ErrTransactionNotExist is the pre-defined error message
	ErrTransactionNotExist = errors.New("transaction are not existed in the mempool")
	// ErrPoolIsFull indicates the pool is full
	ErrPoolIsFull = errors.New("transaction pool reach the max number")
	// ErrDustTx indicates transaction is dust tx
	ErrDustTx = errors.New("transaction is dust tx")
)

type TxMsgEvent struct{ TxMsg *TxPoolMsg }

// TxDesc store tx and related info for mining strategy
type TxDesc struct {
	Tx     *types.Tx `json:"transaction"`
	Added  time.Time `json:"-"`
	Height uint64    `json:"-"`
	Weight uint64    `json:"-"`
	Fee    uint64    `json:"-"`
}

// TxPoolMsg is use for notify pool changes
type TxPoolMsg struct {
	*TxDesc
	MsgType int
}

type orphanTx struct {
	*TxDesc
	expiration time.Time
}

// TxPool is use for store the unconfirmed transaction
type TxPool struct {
	lastUpdated     int64
	mtx             sync.RWMutex
	store           state.Store
	pool            map[bc.Hash]*TxDesc
	utxo            map[bc.Hash]*types.Tx
	orphans         map[bc.Hash]*orphanTx
	orphansByPrev   map[bc.Hash]map[bc.Hash]*orphanTx
	errCache        *lru.Cache
	eventDispatcher *event.Dispatcher
}

// NewTxPool init a new TxPool
func NewTxPool(store state.Store, dispatcher *event.Dispatcher) *TxPool {
	tp := &TxPool{
		lastUpdated:     time.Now().Unix(),
		store:           store,
		pool:            make(map[bc.Hash]*TxDesc),
		utxo:            make(map[bc.Hash]*types.Tx),
		orphans:         make(map[bc.Hash]*orphanTx),
		orphansByPrev:   make(map[bc.Hash]map[bc.Hash]*orphanTx),
		errCache:        lru.New(maxCachedErrTxs),
		eventDispatcher: dispatcher,
	}
	go tp.orphanExpireWorker()
	return tp
}

// AddErrCache add a failed transaction record to lru cache
func (tp *TxPool) AddErrCache(txHash *bc.Hash, err error) {
	tp.mtx.Lock()
	defer tp.mtx.Unlock()

	tp.errCache.Add(txHash, err)
}

// ExpireOrphan expire all the orphans that before the input time range
func (tp *TxPool) ExpireOrphan(now time.Time) {
	tp.mtx.Lock()
	defer tp.mtx.Unlock()

	for hash, orphan := range tp.orphans {
		if orphan.expiration.Before(now) {
			tp.removeOrphan(&hash)
		}
	}
}

// GetErrCache return the error of the transaction
func (tp *TxPool) GetErrCache(txHash *bc.Hash) error {
	tp.mtx.Lock()
	defer tp.mtx.Unlock()

	v, ok := tp.errCache.Get(txHash)
	if !ok {
		return nil
	}
	return v.(error)
}

// RemoveTransaction remove a transaction from the pool
func (tp *TxPool) RemoveTransaction(txHash *bc.Hash) {
	tp.mtx.Lock()
	defer tp.mtx.Unlock()

	txD, ok := tp.pool[*txHash]
	if !ok {
		return
	}

	for _, output := range txD.Tx.ResultIds {
		delete(tp.utxo, *output)
	}
	delete(tp.pool, *txHash)

	atomic.StoreInt64(&tp.lastUpdated, time.Now().Unix())
	tp.eventDispatcher.Post(TxMsgEvent{TxMsg: &TxPoolMsg{TxDesc: txD, MsgType: MsgRemoveTx}})
	log.WithFields(log.Fields{"module": logModule, "tx_id": txHash}).Debug("remove tx from mempool")
}

// GetTransaction return the TxDesc by hash
func (tp *TxPool) GetTransaction(txHash *bc.Hash) (*TxDesc, error) {
	tp.mtx.RLock()
	defer tp.mtx.RUnlock()

	if txD, ok := tp.pool[*txHash]; ok {
		return txD, nil
	}
	return nil, ErrTransactionNotExist
}

// GetTransactions return all the transactions in the pool
func (tp *TxPool) GetTransactions() []*TxDesc {
	tp.mtx.RLock()
	defer tp.mtx.RUnlock()

	txDs := make([]*TxDesc, len(tp.pool))
	i := 0
	for _, desc := range tp.pool {
		txDs[i] = desc
		i++
	}
	return txDs
}

// IsTransactionInPool check wheather a transaction in pool or not
func (tp *TxPool) IsTransactionInPool(txHash *bc.Hash) bool {
	tp.mtx.RLock()
	defer tp.mtx.RUnlock()

	_, ok := tp.pool[*txHash]
	return ok
}

// IsTransactionInErrCache check wheather a transaction in errCache or not
func (tp *TxPool) IsTransactionInErrCache(txHash *bc.Hash) bool {
	tp.mtx.RLock()
	defer tp.mtx.RUnlock()

	_, ok := tp.errCache.Get(txHash)
	return ok
}

// HaveTransaction IsTransactionInErrCache check is  transaction in errCache or pool
func (tp *TxPool) HaveTransaction(txHash *bc.Hash) bool {
	return tp.IsTransactionInPool(txHash) || tp.IsTransactionInErrCache(txHash)
}

func isTransactionNoBtmInput(tx *types.Tx) bool {
	for _, input := range tx.TxData.Inputs {
		if input.AssetID() == *consensus.BTMAssetID {
			return false
		}
	}
	return true
}

func isTransactionZeroOutput(tx *types.Tx) bool {
	for _, output := range tx.TxData.Outputs {
		if output.Amount == uint64(0) {
			return true
		}
	}
	return false
}

func isInvalidBCRPTx(tx *types.Tx) bool {
	for _, output := range tx.TxData.Outputs {
		if bcrp.IsBCRPScript(output.ControlProgram) {
			return true
		}
	}
	return false
}

func (tp *TxPool) IsDust(tx *types.Tx) bool {
	return isTransactionNoBtmInput(tx) || isTransactionZeroOutput(tx) || isInvalidBCRPTx(tx)
}

func (tp *TxPool) processTransaction(tx *types.Tx, height, fee uint64) (bool, error) {
	tp.mtx.Lock()
	defer tp.mtx.Unlock()

	txD := &TxDesc{
		Tx:     tx,
		Weight: tx.SerializedSize,
		Height: height,
		Fee:    fee,
	}
	requireParents, err := tp.checkOrphanUtxos(tx)
	if err != nil {
		return false, err
	}

	if len(requireParents) > 0 {
		return true, tp.addOrphan(txD, requireParents)
	}

	if err := tp.addTransaction(txD); err != nil {
		return false, err
	}

	tp.processOrphans(txD)
	return false, nil
}

// ProcessTransaction is the main entry for txpool handle new tx, ignore dust tx.
func (tp *TxPool) ProcessTransaction(tx *types.Tx, height, fee uint64) (bool, error) {
	if tp.IsDust(tx) {
		log.WithFields(log.Fields{"module": logModule, "tx_id": tx.ID.String()}).Warn("dust tx")
		return false, nil
	}
	return tp.processTransaction(tx, height, fee)
}

func (tp *TxPool) addOrphan(txD *TxDesc, requireParents []*bc.Hash) error {
	if len(tp.orphans) >= maxOrphanNum {
		return ErrPoolIsFull
	}

	orphan := &orphanTx{txD, time.Now().Add(orphanTTL)}
	tp.orphans[txD.Tx.ID] = orphan
	for _, hash := range requireParents {
		if _, ok := tp.orphansByPrev[*hash]; !ok {
			tp.orphansByPrev[*hash] = make(map[bc.Hash]*orphanTx)
		}
		tp.orphansByPrev[*hash][txD.Tx.ID] = orphan
	}
	return nil
}

func (tp *TxPool) addTransaction(txD *TxDesc) error {
	if len(tp.pool) >= maxNewTxNum {
		return ErrPoolIsFull
	}

	tx := txD.Tx
	txD.Added = time.Now()
	tp.pool[tx.ID] = txD
	for _, id := range tx.ResultIds {
		_, err := tx.OriginalOutput(*id)
		if err != nil {
			// error due to it's a retirement, utxo doesn't care this output type so skip it
			continue
		}

		tp.utxo[*id] = tx
	}

	atomic.StoreInt64(&tp.lastUpdated, time.Now().Unix())
	tp.eventDispatcher.Post(TxMsgEvent{TxMsg: &TxPoolMsg{TxDesc: txD, MsgType: MsgNewTx}})
	log.WithFields(log.Fields{"module": logModule, "tx_id": tx.ID.String()}).Debug("Add tx to mempool")
	return nil
}

func (tp *TxPool) checkOrphanUtxos(tx *types.Tx) ([]*bc.Hash, error) {
	view := state.NewUtxoViewpoint()
	if err := tp.store.GetTransactionsUtxo(view, []*bc.Tx{tx.Tx}); err != nil {
		return nil, err
	}

	hashes := []*bc.Hash{}
	for _, hash := range tx.SpentOutputIDs {
		if !view.CanSpend(&hash) && tp.utxo[hash] == nil {
			hashes = append(hashes, &hash)
		}
	}
	return hashes, nil
}

func (tp *TxPool) orphanExpireWorker() {
	ticker := time.NewTicker(orphanExpireScanInterval)
	defer ticker.Stop()

	for now := range ticker.C {
		tp.ExpireOrphan(now)
	}
}

func (tp *TxPool) processOrphans(txD *TxDesc) {
	processOrphans := []*orphanTx{}
	addRely := func(tx *types.Tx) {
		for _, outHash := range tx.ResultIds {
			orphans, ok := tp.orphansByPrev[*outHash]
			if !ok {
				continue
			}

			for _, orphan := range orphans {
				processOrphans = append(processOrphans, orphan)
			}
			delete(tp.orphansByPrev, *outHash)
		}
	}

	addRely(txD.Tx)
	for ; len(processOrphans) > 0; processOrphans = processOrphans[1:] {
		processOrphan := processOrphans[0]
		requireParents, err := tp.checkOrphanUtxos(processOrphan.Tx)
		if err != nil {
			log.WithFields(log.Fields{"module": logModule, "err": err}).Error("processOrphans got unexpect error")
			continue
		}

		if len(requireParents) == 0 {
			addRely(processOrphan.Tx)
			tp.removeOrphan(&processOrphan.Tx.ID)
			tp.addTransaction(processOrphan.TxDesc)
		}
	}
}

func (tp *TxPool) removeOrphan(hash *bc.Hash) {
	orphan, ok := tp.orphans[*hash]
	if !ok {
		return
	}

	for _, spend := range orphan.Tx.SpentOutputIDs {
		orphans, ok := tp.orphansByPrev[spend]
		if !ok {
			continue
		}

		if delete(orphans, *hash); len(orphans) == 0 {
			delete(tp.orphansByPrev, spend)
		}
	}
	delete(tp.orphans, *hash)
}
