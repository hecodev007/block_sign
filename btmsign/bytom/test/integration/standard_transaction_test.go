package integration

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"btmSign/bytom/account"
	"btmSign/bytom/blockchain/pseudohsm"
	"btmSign/bytom/blockchain/signers"
	"btmSign/bytom/crypto/ed25519/chainkd"
	dbm "btmSign/bytom/database/leveldb"
	"btmSign/bytom/protocol/bc/types"
	"btmSign/bytom/protocol/validation"
	"btmSign/bytom/test"
)

func TestP2PKH(t *testing.T) {
	dirPath, err := ioutil.TempDir(".", "TestP2PKH")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirPath)

	testDB := dbm.NewDB("testdb", "leveldb", "temp")
	defer os.RemoveAll("temp")

	chain, _, _, err := test.MockChain(testDB)
	if err != nil {
		t.Fatal(err)
	}
	accountManager := account.NewManager(testDB, chain)
	hsm, err := pseudohsm.New(dirPath)
	if err != nil {
		t.Fatal(err)
	}
	xpub, _, err := hsm.XCreate("TestP2PKH", "password", "en")
	if err != nil {
		t.Fatal(err)
	}
	testAccount, err := accountManager.Create([]chainkd.XPub{xpub.XPub}, 1, "testAccount", signers.BIP0044)
	if err != nil {
		t.Fatal(err)
	}

	controlProg, err := accountManager.CreateAddress(testAccount.ID, false)
	if err != nil {
		t.Fatal(err)
	}
	utxo := test.MockUTXO(controlProg)
	tpl, tx, err := test.MockTx(utxo, testAccount)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := test.MockSign(tpl, hsm, "password"); err != nil {
		t.Fatal(err)
	}

	tx.SerializedSize = 1
	converter := func(prog []byte) ([]byte, error) { return nil, nil }
	if _, err = validation.ValidateTx(types.MapTx(tx), test.MockBlock(), converter); err != nil {
		t.Fatal(err)
	}

}

func TestBip0032P2PKH(t *testing.T) {
	dirPath, err := ioutil.TempDir(".", "TestP2PKH")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirPath)

	testDB := dbm.NewDB("testdb", "leveldb", "temp")
	defer os.RemoveAll("temp")

	chain, _, _, err := test.MockChain(testDB)
	if err != nil {
		t.Fatal(err)
	}

	accountManager := account.NewManager(testDB, chain)
	hsm, err := pseudohsm.New(dirPath)
	if err != nil {
		t.Fatal(err)
	}

	xpub, _, err := hsm.XCreate("TestP2PKH", "password", "en")
	if err != nil {
		t.Fatal(err)
	}

	testAccount, err := accountManager.Create([]chainkd.XPub{xpub.XPub}, 1, "testAccount1", signers.BIP0044)
	if err != nil {
		t.Fatal(err)
	}

	signer, err := signers.Create("account", []chainkd.XPub{xpub.XPub}, 1, 2, signers.BIP0032)
	if err != nil {
		t.Fatal(err)
	}

	testAccount.Signer = signer
	rawAccount, err := json.Marshal(testAccount)
	if err != nil {
		t.Fatal(err)
	}

	testDB.Set(account.Key(testAccount.ID), rawAccount)
	controlProg, err := accountManager.CreateAddress(testAccount.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	utxo := test.MockUTXO(controlProg)
	tpl, tx, err := test.MockTx(utxo, testAccount)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := test.MockSign(tpl, hsm, "password"); err != nil {
		t.Fatal(err)
	}

	tx.SerializedSize = 1
	converter := func(prog []byte) ([]byte, error) { return nil, nil }
	if _, err = validation.ValidateTx(types.MapTx(tx), test.MockBlock(), converter); err != nil {
		t.Fatal(err)
	}
}

func TestP2SH(t *testing.T) {
	dirPath, err := ioutil.TempDir(".", "TestP2SH")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirPath)

	testDB := dbm.NewDB("testdb", "leveldb", "temp")
	defer os.RemoveAll("temp")

	chain, _, _, err := test.MockChain(testDB)
	if err != nil {
		t.Fatal(err)
	}

	accountManager := account.NewManager(testDB, chain)
	hsm, err := pseudohsm.New(dirPath)
	if err != nil {
		t.Fatal(err)
	}

	xpub1, _, err := hsm.XCreate("TestP2SH1", "password", "en")
	if err != nil {
		t.Fatal(err)
	}

	xpub2, _, err := hsm.XCreate("TestP2SH2", "password", "en")
	if err != nil {
		t.Fatal(err)
	}

	testAccount, err := accountManager.Create([]chainkd.XPub{xpub1.XPub, xpub2.XPub}, 2, "testAccount", signers.BIP0044)
	if err != nil {
		t.Fatal(err)
	}

	controlProg, err := accountManager.CreateAddress(testAccount.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	utxo := test.MockUTXO(controlProg)
	tpl, tx, err := test.MockTx(utxo, testAccount)
	if err != nil {
		t.Fatal(err)
	}

	// the number of sign transaction is equal to the count of xpub for account
	if _, err := test.MockSign(tpl, hsm, "password"); err != nil {
		t.Fatal(err)
	}
	if _, err := test.MockSign(tpl, hsm, "password"); err != nil {
		t.Fatal(err)
	}

	tx.SerializedSize = 1
	converter := func(prog []byte) ([]byte, error) { return nil, nil }
	if _, err = validation.ValidateTx(types.MapTx(tx), test.MockBlock(), converter); err != nil {
		t.Fatal(err)
	}
}

func TestBip0032P2SH(t *testing.T) {
	dirPath, err := ioutil.TempDir(".", "TestP2SH")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirPath)

	testDB := dbm.NewDB("testdb", "leveldb", "temp")
	defer os.RemoveAll("temp")

	chain, _, _, err := test.MockChain(testDB)
	if err != nil {
		t.Fatal(err)
	}

	accountManager := account.NewManager(testDB, chain)
	hsm, err := pseudohsm.New(dirPath)
	if err != nil {
		t.Fatal(err)
	}

	xpub1, _, err := hsm.XCreate("TestP2SH1", "password", "en")
	if err != nil {
		t.Fatal(err)
	}

	xpub2, _, err := hsm.XCreate("TestP2SH2", "password", "en")
	if err != nil {
		t.Fatal(err)
	}

	testAccount, err := accountManager.Create([]chainkd.XPub{xpub1.XPub, xpub2.XPub}, 2, "testAccount", signers.BIP0044)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := signers.Create("account", []chainkd.XPub{xpub1.XPub, xpub2.XPub}, 2, 2, signers.BIP0032)
	if err != nil {
		t.Fatal(err)
	}

	testAccount.Signer = signer
	rawAccount, err := json.Marshal(testAccount)
	if err != nil {
		t.Fatal(err)
	}

	testDB.Set(account.Key(testAccount.ID), rawAccount)

	controlProg, err := accountManager.CreateAddress(testAccount.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	utxo := test.MockUTXO(controlProg)
	tpl, tx, err := test.MockTx(utxo, testAccount)
	if err != nil {
		t.Fatal(err)
	}

	// the number of sign transaction is equal to the count of xpub for account
	if _, err := test.MockSign(tpl, hsm, "password"); err != nil {
		t.Fatal(err)
	}
	if _, err := test.MockSign(tpl, hsm, "password"); err != nil {
		t.Fatal(err)
	}

	tx.SerializedSize = 1
	converter := func(prog []byte) ([]byte, error) { return nil, nil }
	if _, err = validation.ValidateTx(types.MapTx(tx), test.MockBlock(), converter); err != nil {
		t.Fatal(err)
	}
}

func TestMutilNodeSign(t *testing.T) {
	dirPath, err := ioutil.TempDir(".", "TestMutilNodeSign")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirPath)

	testDB := dbm.NewDB("testdb", "leveldb", "temp")
	defer os.RemoveAll("temp")

	chain, _, _, err := test.MockChain(testDB)
	if err != nil {
		t.Fatal(err)
	}

	accountManager := account.NewManager(testDB, chain)
	hsm, err := pseudohsm.New(dirPath)
	if err != nil {
		t.Fatal(err)
	}

	xpub1, _, err := hsm.XCreate("TestMutilNodeSign1", "password1", "en")
	if err != nil {
		t.Fatal(err)
	}

	xpub2, _, err := hsm.XCreate("TestMutilNodeSign2", "password2", "en")
	if err != nil {
		t.Fatal(err)
	}

	testAccount, err := accountManager.Create([]chainkd.XPub{xpub1.XPub, xpub2.XPub}, 2, "testAccount", signers.BIP0044)
	if err != nil {
		t.Fatal(err)
	}

	controlProg, err := accountManager.CreateAddress(testAccount.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	utxo := test.MockUTXO(controlProg)
	tpl, tx, err := test.MockTx(utxo, testAccount)
	if err != nil {
		t.Fatal(err)
	}

	if finishSign, err := test.MockSign(tpl, hsm, "password"); err != nil {
		t.Fatal(err)
	} else if finishSign == true {
		t.Fatal("sign progress is finish, but either xpub1 nor xpub2 is signed")
	}

	if finishSign, err := test.MockSign(tpl, hsm, "password1"); err != nil {
		t.Fatal(err)
	} else if finishSign == true {
		t.Fatal("sign progress is finish, but xpub2 is not signed")
	}

	if finishSign, err := test.MockSign(tpl, hsm, "password2"); err != nil {
		t.Fatal(err)
	} else if finishSign == false {
		t.Fatal("sign progress is not finish,  but both xpub1 and xpub2 is signed")
	}

	tx.SerializedSize = 1
	converter := func(prog []byte) ([]byte, error) { return nil, nil }
	if _, err = validation.ValidateTx(types.MapTx(tx), test.MockBlock(), converter); err != nil {
		t.Fatal(err)
	}
}

func TestBip0032MutilNodeSign(t *testing.T) {
	dirPath, err := ioutil.TempDir(".", "TestMutilNodeSign")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirPath)

	testDB := dbm.NewDB("testdb", "leveldb", "temp")
	defer os.RemoveAll("temp")

	chain, _, _, err := test.MockChain(testDB)
	if err != nil {
		t.Fatal(err)
	}

	accountManager := account.NewManager(testDB, chain)
	hsm, err := pseudohsm.New(dirPath)
	if err != nil {
		t.Fatal(err)
	}

	xpub1, _, err := hsm.XCreate("TestMutilNodeSign1", "password1", "en")
	if err != nil {
		t.Fatal(err)
	}

	xpub2, _, err := hsm.XCreate("TestMutilNodeSign2", "password2", "en")
	if err != nil {
		t.Fatal(err)
	}

	testAccount, err := accountManager.Create([]chainkd.XPub{xpub1.XPub, xpub2.XPub}, 2, "testAccount", signers.BIP0044)
	if err != nil {
		t.Fatal(err)
	}

	signer, err := signers.Create("account", []chainkd.XPub{xpub1.XPub, xpub2.XPub}, 2, 2, signers.BIP0032)
	if err != nil {
		t.Fatal(err)
	}

	testAccount.Signer = signer
	rawAccount, err := json.Marshal(testAccount)
	if err != nil {
		t.Fatal(err)
	}

	testDB.Set(account.Key(testAccount.ID), rawAccount)

	controlProg, err := accountManager.CreateAddress(testAccount.ID, false)
	if err != nil {
		t.Fatal(err)
	}

	utxo := test.MockUTXO(controlProg)
	tpl, tx, err := test.MockTx(utxo, testAccount)
	if err != nil {
		t.Fatal(err)
	}

	if finishSign, err := test.MockSign(tpl, hsm, "password"); err != nil {
		t.Fatal(err)
	} else if finishSign == true {
		t.Fatal("sign progress is finish, but either xpub1 nor xpub2 is signed")
	}

	if finishSign, err := test.MockSign(tpl, hsm, "password1"); err != nil {
		t.Fatal(err)
	} else if finishSign == true {
		t.Fatal("sign progress is finish, but xpub2 is not signed")
	}

	if finishSign, err := test.MockSign(tpl, hsm, "password2"); err != nil {
		t.Fatal(err)
	} else if finishSign == false {
		t.Fatal("sign progress is not finish,  but both xpub1 and xpub2 is signed")
	}

	tx.SerializedSize = 1
	converter := func(prog []byte) ([]byte, error) { return nil, nil }
	if _, err = validation.ValidateTx(types.MapTx(tx), test.MockBlock(), converter); err != nil {
		t.Fatal(err)
	}
}
