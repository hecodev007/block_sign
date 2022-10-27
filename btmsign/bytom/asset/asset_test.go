package asset

import (
	"context"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"btmSign/bytom/consensus"
	"btmSign/bytom/crypto/ed25519/chainkd"
	"btmSign/bytom/database"
	dbm "btmSign/bytom/database/leveldb"
	"btmSign/bytom/event"
	"btmSign/bytom/protocol"
	"btmSign/bytom/testutil"
)

func TestDefineAssetWithLowercase(t *testing.T) {
	reg := mockNewRegistry(t)
	alias := "lower"
	asset, err := reg.Define([]chainkd.XPub{testutil.TestXPub}, 1, nil, 0, alias, nil)
	if err != nil {
		t.Fatal(err)
	}
	if *asset.Alias != strings.ToUpper(alias) {
		t.Fatal("created asset alias should be uppercase")
	}
}

func TestDefineAssetWithSpaceTrimed(t *testing.T) {
	reg := mockNewRegistry(t)
	alias := " WITH SPACE "
	asset, err := reg.Define([]chainkd.XPub{testutil.TestXPub}, 1, nil, 0, alias, nil)
	if err != nil {
		t.Fatal(err)
	}
	if *asset.Alias != strings.TrimSpace(alias) {
		t.Fatal("created asset alias should be uppercase")
	}
}

func TestDefineAsset(t *testing.T) {
	ctx := context.Background()
	reg := mockNewRegistry(t)
	asset, err := reg.Define([]chainkd.XPub{testutil.TestXPub}, 1, nil, 0, "asset-alias", nil)
	if err != nil {
		testutil.FatalErr(t, err)
	}

	limitAsset, err := reg.Define([]chainkd.XPub{testutil.TestXPub}, 1, nil, 100, "limit-asset", nil)
	if err != nil {
		testutil.FatalErr(t, err)
	}

	assets := []*Asset{asset, limitAsset}
	for _, ass := range assets {
		found, err := reg.FindByID(ctx, &ass.AssetID)
		if err != nil {
			t.Errorf("unexpected error %v", err)
		}

		if !testutil.DeepEqual(ass, found) {
			t.Errorf("expected asset %v to be recorded as %v", ass, found)
		}
	}
}

func TestDefineBtmAsset(t *testing.T) {
	reg := mockNewRegistry(t)
	_, err := reg.Define([]chainkd.XPub{testutil.TestXPub}, 1, nil, 0, consensus.BTMAlias, nil)
	if err == nil {
		testutil.FatalErr(t, err)
	}
}

func TestFindAssetByID(t *testing.T) {
	ctx := context.Background()
	reg := mockNewRegistry(t)
	keys := []chainkd.XPub{testutil.TestXPub}
	asset, err := reg.Define(keys, 1, nil, 0, "TESTASSET", nil)
	if err != nil {
		testutil.FatalErr(t, err)

	}
	found, err := reg.FindByID(ctx, &asset.AssetID)
	if err != nil {
		testutil.FatalErr(t, err)
	}

	if !testutil.DeepEqual(asset, found) {
		t.Errorf("expected %v and %v to match", asset, found)
	}
}

func TestUpdateAssetAlias(t *testing.T) {
	reg := mockNewRegistry(t)

	oldAlias := "OLD_ALIAS"
	newAlias := "NEW_ALIAS"

	asset, err := reg.Define([]chainkd.XPub{testutil.TestXPub}, 1, nil, 0, oldAlias, nil)
	if err != nil {
		testutil.FatalErr(t, err)
	}

	if reg.UpdateAssetAlias(asset.AssetID.String(), newAlias) != nil {
		testutil.FatalErr(t, err)
	}

	asset1, err := reg.FindByAlias(newAlias)
	if err != nil {
		testutil.FatalErr(t, err)
	}

	gotAlias := *asset1.Alias
	if !reflect.DeepEqual(gotAlias, newAlias) {
		t.Fatalf("alias:\ngot:  %v\nwant: %v", gotAlias, newAlias)
	}
}

type SortByAssetsAlias []*Asset

func (a SortByAssetsAlias) Len() int { return len(a) }
func (a SortByAssetsAlias) Less(i, j int) bool {
	return strings.Compare(*a[i].Alias, *a[j].Alias) <= 0
}
func (a SortByAssetsAlias) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func TestListAssets(t *testing.T) {
	reg := mockNewRegistry(t)

	firstAlias := "FIRST_ALIAS"
	secondAlias := "SECOND_ALIAS"

	firstAsset, err := reg.Define([]chainkd.XPub{testutil.TestXPub}, 1, nil, 0, firstAlias, nil)
	if err != nil {
		testutil.FatalErr(t, err)
	}

	secondAsset, err := reg.Define([]chainkd.XPub{testutil.TestXPub}, 1, nil, 0, secondAlias, nil)
	if err != nil {
		testutil.FatalErr(t, err)
	}

	wantAssets := []*Asset{DefaultNativeAsset, firstAsset, secondAsset}

	gotAssets, err := reg.ListAssets("")
	if err != nil {
		testutil.FatalErr(t, err)
	}
	sort.Sort(SortByAssetsAlias(wantAssets))
	sort.Sort(SortByAssetsAlias(gotAssets))
	if !testutil.DeepEqual(gotAssets, wantAssets) {
		t.Fatalf("got:\ngot:  %v\nwant: %v", gotAssets, wantAssets)
	}
}

func mockChain(testDB dbm.DB) (*protocol.Chain, error) {
	store := database.NewStore(testDB)
	dispatcher := event.NewDispatcher()
	txPool := protocol.NewTxPool(store, dispatcher)
	chain, err := protocol.NewChain(store, txPool, dispatcher)
	if err != nil {
		return nil, err
	}
	return chain, nil
}

func mockNewRegistry(t *testing.T) *Registry {
	dirPath, err := ioutil.TempDir(".", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirPath)

	testDB := dbm.NewDB("testdb", "leveldb", "temp")
	defer os.RemoveAll("temp")

	chain, err := mockChain(testDB)
	if err != nil {
		t.Fatal(err)
	}

	return NewRegistry(testDB, chain)
}
