package config

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
)

var (
	_ keyring.Keyring = (*privKeyKeyring)(nil)
)

// privKeyKeyring partially implements the keyring.Keyring
// interface, the functionality which is implemented covers
// only the usage of keyring.Keyring in the client.Context
// and tx.Factory in transaction signing operations.
// Other calls which are out of the scope of Tx signing
// yield to panics.
type privKeyKeyring struct {
	addr    sdk.AccAddress
	pubKey  cryptotypes.PubKey
	privKey cryptotypes.PrivKey
}

func GetAuth(mnemonic string) (keyring.Keyring, sdk.ValAddress, sdk.AccAddress) {
	seed := bip39.NewSeed(mnemonic, "")
	master, ch := hd.ComputeMastersFromSeed(seed)
	path := "m/44'/118'/0'/0/0"

	priv, err := hd.DerivePrivateKeyForPath(master, ch, path)
	if err != nil {
		panic(err)
	}
	kr := newPrivKeyKeyring(hex.EncodeToString(priv))
	return kr, sdk.ValAddress(kr.addr), kr.addr
}

func newPrivKeyKeyring(hexKey string) *privKeyKeyring {
	b, err := hex.DecodeString(hexKey)
	if err != nil {
		panic(err)
	}
	key := &secp256k1.PrivKey{
		Key: b,
	}

	return &privKeyKeyring{
		addr:    sdk.AccAddress(key.PubKey().Address()),
		pubKey:  key.PubKey(),
		privKey: key,
	}
}

func (p privKeyKeyring) Key(uid string) (keyring.Info, error) {
	return p.KeyByAddress(p.addr)
}

func (p privKeyKeyring) KeyByAddress(address sdk.Address) (keyring.Info, error) {
	if !address.Equals(p.addr) {
		return nil, fmt.Errorf("key not found")
	}

	return &info{
		pubKey: p.pubKey,
		addr:   p.addr,
	}, nil
}

func (p privKeyKeyring) Sign(uid string, msg []byte) ([]byte, cryptotypes.PubKey, error) {
	return p.SignByAddress(p.addr, msg)
}

func (p privKeyKeyring) SignByAddress(address sdk.Address, msg []byte) ([]byte, cryptotypes.PubKey, error) {
	if !p.addr.Equals(address) {
		return nil, nil, fmt.Errorf("key not found")
	}

	signed, err := p.privKey.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return signed, p.pubKey, nil
}

var _ keyring.Info = (*info)(nil)

type info struct {
	pubKey cryptotypes.PubKey
	addr   sdk.AccAddress
}

func (i info) GetType() keyring.KeyType {
	return keyring.TypeLocal
}

func (i info) GetName() string {
	return ""
}

func (i info) GetPubKey() cryptotypes.PubKey {
	return i.pubKey
}

func (i info) GetAddress() sdk.AccAddress {
	return i.addr
}

// implementations that must never be called

func (p privKeyKeyring) List() ([]keyring.Info, error) {
	panic("must never be called")
}

func (p privKeyKeyring) SupportedAlgorithms() (keyring.SigningAlgoList, keyring.SigningAlgoList) {
	panic("must never be called")
}

func (i info) GetPath() (*hd.BIP44Params, error) {
	panic("must never be called")
}

func (i info) GetAlgo() hd.PubKeyType {
	panic("must never be called")
}

func (p privKeyKeyring) Delete(uid string) error {
	panic("must never be called")
}

func (p privKeyKeyring) DeleteByAddress(address sdk.Address) error {
	panic("must never be called")
}

func (p privKeyKeyring) NewMnemonic(uid string, language keyring.Language, hdPath, bip39Passphrase string, algo keyring.SignatureAlgo) (keyring.Info, string, error) {
	panic("must never be called")
}

func (p privKeyKeyring) NewAccount(uid, mnemonic, bip39Passphrase, hdPath string, algo keyring.SignatureAlgo) (keyring.Info, error) {
	panic("must never be called")
}

func (p privKeyKeyring) SaveLedgerKey(uid string, algo keyring.SignatureAlgo, hrp string, coinType, account, index uint32) (keyring.Info, error) {
	panic("must never be called")
}

func (p privKeyKeyring) SavePubKey(uid string, pubkey cryptotypes.PubKey, algo hd.PubKeyType) (keyring.Info, error) {
	panic("must never be called")
}

func (p privKeyKeyring) SaveMultisig(uid string, pubkey cryptotypes.PubKey) (keyring.Info, error) {
	panic("must never be called")
}

func (p privKeyKeyring) ImportPrivKey(uid, armor, passphrase string) error {
	panic("must never be called")
}

func (p privKeyKeyring) ImportPubKey(uid string, armor string) error {
	panic("must never be called")
}

func (p privKeyKeyring) ExportPubKeyArmor(uid string) (string, error) {
	panic("must never be called")
}

func (p privKeyKeyring) ExportPubKeyArmorByAddress(address sdk.Address) (string, error) {
	panic("must never be called")
}

func (p privKeyKeyring) ExportPrivKeyArmor(uid, encryptPassphrase string) (armor string, err error) {
	panic("must never be called")
}

func (p privKeyKeyring) ExportPrivKeyArmorByAddress(address sdk.Address, encryptPassphrase string) (armor string, err error) {
	panic("must never be called")
}
