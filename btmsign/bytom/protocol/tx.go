package protocol

import (
	log "github.com/sirupsen/logrus"

	"btmSign/bytom/consensus/bcrp"
	"btmSign/bytom/errors"
	"btmSign/bytom/protocol/bc"
	"btmSign/bytom/protocol/bc/types"
	"btmSign/bytom/protocol/state"
	"btmSign/bytom/protocol/validation"
)

// ErrBadTx is returned for transactions failing validation
var ErrBadTx = errors.New("invalid transaction")

// GetTransactionsUtxo return all the utxos that related to the txs' inputs
func (c *Chain) GetTransactionsUtxo(view *state.UtxoViewpoint, txs []*bc.Tx) error {
	return c.store.GetTransactionsUtxo(view, txs)
}

// ValidateTx validates the given transaction. A cache holds
// per-transaction validation results and is consulted before
// performing full validation.
func (c *Chain) ValidateTx(tx *types.Tx) (bool, error) {
	if ok := c.txPool.HaveTransaction(&tx.ID); ok {
		return false, c.txPool.GetErrCache(&tx.ID)
	}

	if c.txPool.IsDust(tx) {
		c.txPool.AddErrCache(&tx.ID, ErrDustTx)
		return false, ErrDustTx
	}

	bh := c.BestBlockHeader()
	gasStatus, err := validation.ValidateTx(tx.Tx, types.MapBlock(&types.Block{BlockHeader: *bh}), c.ProgramConverter)
	if err != nil {
		log.WithFields(log.Fields{"module": logModule, "tx_id": tx.Tx.ID.String(), "error": err}).Info("transaction status fail")
		c.txPool.AddErrCache(&tx.ID, err)
		return false, err
	}

	return c.txPool.ProcessTransaction(tx, bh.Height, gasStatus.BTMValue)
}

//ProgramConverter convert program. Only for BCRP now
func (c *Chain) ProgramConverter(prog []byte) ([]byte, error) {
	hash, err := bcrp.ParseContractHash(prog)
	if err != nil {
		return nil, err
	}

	return c.store.GetContract(hash)
}
