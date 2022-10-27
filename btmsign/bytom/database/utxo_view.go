package database

import (
	dbm "btmSign/bytom/database/leveldb"
	"btmSign/bytom/database/storage"
	"btmSign/bytom/errors"
	"btmSign/bytom/protocol/bc"
	"btmSign/bytom/protocol/state"
	"github.com/golang/protobuf/proto"
)

func CalcUtxoKey(hash *bc.Hash) []byte {
	return append(UtxoKeyPrefix, hash.Bytes()...)
}

func getTransactionsUtxo(db dbm.DB, view *state.UtxoViewpoint, txs []*bc.Tx) error {
	for _, tx := range txs {
		for _, prevout := range tx.SpentOutputIDs {
			if view.HasUtxo(&prevout) {
				continue
			}

			data := db.Get(CalcUtxoKey(&prevout))
			if data == nil {
				continue
			}

			var utxo storage.UtxoEntry
			if err := proto.Unmarshal(data, &utxo); err != nil {
				return errors.Wrap(err, "unmarshaling utxo entry")
			}

			view.Entries[prevout] = &utxo
		}
	}

	return nil
}

func getUtxo(db dbm.DB, hash *bc.Hash) (*storage.UtxoEntry, error) {
	var utxo storage.UtxoEntry
	data := db.Get(CalcUtxoKey(hash))
	if data == nil {
		return nil, errors.New("can't find utxo in db")
	}
	if err := proto.Unmarshal(data, &utxo); err != nil {
		return nil, errors.Wrap(err, "unmarshaling utxo entry")
	}
	return &utxo, nil
}

func saveUtxoView(batch dbm.Batch, view *state.UtxoViewpoint) error {
	for key, entry := range view.Entries {
		if entry.Spent && entry.Type != storage.CoinbaseUTXOType {
			batch.Delete(CalcUtxoKey(&key))
			continue
		}

		b, err := proto.Marshal(entry)
		if err != nil {
			return errors.Wrap(err, "marshaling utxo entry")
		}
		batch.Set(CalcUtxoKey(&key), b)
	}
	return nil
}

func SaveUtxoView(batch dbm.Batch, view *state.UtxoViewpoint) error {
	return saveUtxoView(batch, view)
}
