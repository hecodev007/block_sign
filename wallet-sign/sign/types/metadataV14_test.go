package types_test

import (
	"testing"

	"github.com/group-coldwallet/wallet-sign/sign/types"
	. "github.com/group-coldwallet/wallet-sign/sign/types"
	"github.com/stretchr/testify/assert"
)

// Verify that (Decode . Encode) outputs the input.
func TestMetadataV14EncodeDecodeRoundtrip(t *testing.T) {
	// Decode the metadata
	var metadata Metadata
	err := DecodeFromHexString(MetadataV14Data, &metadata)
	assert.EqualValues(t, metadata.Version, 14)
	assert.NoError(t, err)

	// Now encode it
	encoded, err := EncodeToHexString(metadata)
	assert.NoError(t, err)

	// Verify the encoded metadata equals the original one
	assert.Equal(t, MetadataV14Data, encoded)

	// Verify that decoding the encoded metadata
	// equals the decoded original metadata
	var decodedMetadata Metadata
	err = DecodeFromHexString(encoded, &decodedMetadata)
	assert.NoError(t, err)
	assert.EqualValues(t, metadata, decodedMetadata)
}

/* Test Metadata interface functions for v14 */

func TestMetadataV14_TestFindCallIndexWithUnknownFunction(t *testing.T) {
	var metadata Metadata
	err := DecodeFromHexString(MetadataV14Data, &metadata)
	assert.EqualValues(t, metadata.Version, 14)
	assert.NoError(t, err)

	_, err = metadata.FindCallIndex("Module2_14.unknownFunction")
	assert.Error(t, err)
}

// Verify that we can find the index of a valid call
func TestMetadataV14FindCallIndex(t *testing.T) {
	var meta Metadata
	err := DecodeFromHexString(MetadataV14Data, &meta)
	assert.NoError(t, err)
	index, err := meta.FindCallIndex("Balances.transfer")
	assert.NoError(t, err)
	assert.Equal(t, index, CallIndex{SectionIndex: 5, MethodIndex: 0})
}

// Verify that we get an error when querying for an invalid
// call with FindCallIndex.
func TestMetadataV14FindCallIndexNonExistent(t *testing.T) {
	var meta Metadata
	err := DecodeFromHexString(MetadataV14Data, &meta)
	assert.NoError(t, err)
	_, err = meta.FindCallIndex("Doesnt.Exist")
	assert.Error(t, err)
}

// Verify that we obtain the right modName, varName pair for a given Event id
func TestMetadataV14FindEventNamesForEventID(t *testing.T) {
	var meta Metadata
	err := DecodeFromHexString(MetadataV14Data, &meta)
	assert.NoError(t, err)

	modName, varName, err := meta.FindEventNamesForEventID(EventID{5, 2})
	assert.NoError(t, err)
	assert.Equal(t, modName, NewText("Balances"))
	assert.Equal(t, varName, NewText("Transfer"))
}

// Verify that we get an error when passing an invalid module ID
func TestMetadataV14FindEventNamesInvalidModuleID(t *testing.T) {
	var meta Metadata
	err := DecodeFromHexString(MetadataV14Data, &meta)
	assert.NoError(t, err)

	_, _, err = meta.FindEventNamesForEventID(EventID{100, 2})
	assert.Error(t, err)
}

// Verify that we get an error when passing an invalid event ID
func TestMetadataV14FindEventNamesInvalidEventID(t *testing.T) {
	var meta Metadata
	err := DecodeFromHexString(MetadataV14Data, &meta)
	assert.NoError(t, err)

	_, _, err = meta.FindEventNamesForEventID(EventID{5, 42})
	assert.Error(t, err)
}

func TestMetadataV14FindStorageEntryMetadata(t *testing.T) {
	var meta Metadata
	err := DecodeFromHexString(MetadataV14Data, &meta)
	assert.NoError(t, err)

	_, err = meta.FindStorageEntryMetadata("System", "Account")
	assert.NoError(t, err)
}

// Verify FindStorageEntryMetadata returns an err when
// the given module can't be found.
func TestMetadataV14FindStorageEntryMetadataInvalidModule(t *testing.T) {
	var meta Metadata
	err := DecodeFromHexString(MetadataV14Data, &meta)
	assert.NoError(t, err)

	_, err = meta.FindStorageEntryMetadata("SystemZ", "Account")
	assert.Error(t, err)
}

// Verify FindStorageEntryMetadata returns an err when
// it doesn't find a storage within an existing module.
func TestMetadataV14FindStorageEntryMetadataInvalidStorage(t *testing.T) {
	var meta Metadata
	err := DecodeFromHexString(MetadataV14Data, &meta)
	assert.NoError(t, err)

	_, err = meta.FindStorageEntryMetadata("System", "Accountz")
	assert.Error(t, err)
}

func TestMetadataV14ExistsModuleMetadata(t *testing.T) {
	var meta Metadata
	err := DecodeFromHexString(MetadataV14Data, &meta)
	if err != nil {
		t.Fatal(err)
	}
	res := meta.ExistsModuleMetadata("System")
	assert.True(t, res)
}

/* Unit tests covering decoding/encoding of nested Metadata v14 types */

func TestMetadataV14PalletEmpty(t *testing.T) {
	var pallet = PalletMetadataV14{
		Name:       NewText("System"),
		HasStorage: false,
		HasCalls:   false,
		HasEvents:  false,
		Constants:  nil,
		HasErrors:  false,
		Index:      42,
	}

	encoded, err := EncodeToBytes(pallet)
	assert.NoError(t, err)

	var encodedPallets PalletMetadataV14
	err = DecodeFromBytes(encoded, &encodedPallets)
	assert.NoError(t, err)

	// Verify they are the same value
	assert.EqualValues(t, encodedPallets, pallet)
}

func TestMetadataV14PalletFilled(t *testing.T) {
	var pallet = PalletMetadataV14{
		Name:       NewText("System"),
		HasStorage: true,
		Storage: StorageMetadataV14{
			Prefix: "Pre-fix",
			Items: []StorageEntryMetadataV14{
				{
					Name:     "StorageName",
					Modifier: StorageFunctionModifierV0{IsOptional: true},
					Type: StorageEntryTypeV14{
						IsPlainType: false,
						IsMap:       true,
						AsMap: MapTypeV14{
							Hashers: []StorageHasherV10{
								{IsBlake2_128: true}, {IsBlake2_256: true},
							},
							Key:   NewSi1LookupTypeIDFromUInt(3),
							Value: NewSi1LookupTypeIDFromUInt(4),
						},
					},
				},
				{
					Name: "Account",
					Modifier: types.StorageFunctionModifierV0{
						IsOptional: false,
						IsDefault:  true,
						IsRequired: false,
					},
					Type: types.StorageEntryTypeV14{
						IsPlainType: false,
						IsMap:       true,
						AsMap: types.MapTypeV14{
							Hashers: []types.StorageHasherV10{
								{
									IsBlake2_128:       false,
									IsBlake2_256:       false,
									IsBlake2_128Concat: true,
									IsTwox128:          false,
									IsTwox256:          false,
									IsTwox64Concat:     false,
									IsIdentity:         false,
								},
							},
						},
					},
					Fallback: types.Bytes{
						0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
					},
					Documentation: []types.Text{" The full account information for a particular account ID."},
				},
			},
		},
		HasCalls:  true,
		Calls:     FunctionMetadataV14{Type: NewSi1LookupTypeIDFromUInt(24)},
		HasEvents: true,
		Events:    EventMetadataV14{Type: NewSi1LookupTypeIDFromUInt(72)},
		Constants: []ConstantMetadataV14{
			{
				Name:  NewText("Yellow"),
				Type:  NewSi1LookupTypeIDFromUInt(83),
				Value: []byte("Valuez"),
				Docs:  []Text{"README", "Contribute"},
			},
		},
		HasErrors: true,
		Errors:    ErrorMetadataV14{Type: NewSi1LookupTypeIDFromUInt(57)},
		Index:     42,
	}

	encoded, err := EncodeToBytes(pallet)
	assert.NoError(t, err)

	var encodedPallets PalletMetadataV14
	err = DecodeFromBytes(encoded, &encodedPallets)
	assert.NoError(t, err)

	// Verify they are the same
	assert.Equal(t, encodedPallets, pallet)
}

func TestSi1TypeDecodeEncode(t *testing.T) {
	type Si1Type struct {
		Path   Si1Path
		Params []Si1TypeParameter
		Def    Si1TypeDef
		Docs   []Text
	}

	// Replicate the first Si1Type we get from rpc json, marsh it, and aside encode it, and decode it
	var ti = Si1Type{
		Path: []Text{"sp_core", "crypto", "AccountId32"},
		Def: Si1TypeDef{
			IsComposite: true,
			Composite: Si1TypeDefComposite{
				Fields: []Si1Field{
					{
						Type:        NewSi1LookupTypeIDFromUInt(1),
						HasTypeName: true,
						TypeName:    NewText("[u8; 32]"),
					},
				},
			},
		},
	}

	// Verify that (decode . encode) equals the original value
	encoded, err := EncodeToHexString(ti)
	assert.NoError(t, err)

	var decoded Si1Type
	DecodeFromHexString(encoded, &decoded)

	assert.Equal(t, ti, decoded)
}
