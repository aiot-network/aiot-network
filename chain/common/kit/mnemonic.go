package kit

import (
	"encoding/hex"
	"fmt"
	"github.com/aiot-network/aiotchain/chain/common/kit/hd"
	"github.com/aiot-network/aiotchain/tools/crypto/base58"
	"github.com/aiot-network/aiotchain/tools/crypto/bip32"
	"github.com/aiot-network/aiotchain/tools/crypto/bip39"
	"github.com/aiot-network/aiotchain/tools/crypto/ecc/secp256k1"
	"github.com/aiot-network/aiotchain/tools/crypto/seed"
)

func Entropy() (string, error) {
	s, err := seed.GenerateSeed(seed.DefaultSeedBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(s), nil
}

func Mnemonic(entropyStr string) (string, error) {
	entropy, err := hex.DecodeString(entropyStr)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

func MnemonicToEc(mnemonic string) (*secp256k1.PrivateKey, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return nil, err
	}
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}
	return secp256k1.PrivKeyFromString(hex.EncodeToString(masterKey.Key))
}

func MnemonicToEcString(mnemonic string) (string, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return "", err
	}
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(masterKey.Key), nil
}

func MnemonicToSeed(mnemonic string) string {
	seed := bip39.NewSeed(mnemonic, "")

	return hex.EncodeToString(seed)
}

func HdDerive(network string, seed string, index uint32) (string, error) {
	bytes, err := hex.DecodeString(seed)
	if err != nil {
		return "", err
	}

	mKey, err := bip32.NewMasterKey2(bytes, bip32.DefaultBip32Version)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("m/44'/0'/0'/0/%d", index)
	derivePath, err := hd.ParseDerivationPath(path)
	if err != nil {
		return "", err
	}

	for _, i := range derivePath {
		tmk, err := mKey.NewChildKey(i)
		if err != nil {
			return "", err
		}
		mKey = tmk
	}

	return mKey.String(), nil
}

func HdPrivateToPublic(private string, network string) (string, error) {
	data := base58.Decode(private)
	masterKey, err := bip32.Deserialize2(data, bip32.DefaultBip32Version)
	if err != nil {
		return "", err
	}
	if !masterKey.IsPrivate {
		return "", fmt.Errorf("%s is not a HD (BIP32) private key", private)
	}
	pubKey := masterKey.PublicKey()
	return pubKey.String(), nil
}

func HdToEc(hdPrivateOrPublic string, network string) (string, error) {
	data := base58.Decode(hdPrivateOrPublic)
	key, err := bip32.Deserialize2(data, bip32.DefaultBip32Version)
	if err != nil {
		return "", err
	}
	if key.IsPrivate {
		return fmt.Sprintf("%x", key.Key[:]), nil
	} else {
		return fmt.Sprintf("%x", key.PublicKey().Key[:]), nil
	}
}

func HdDeriveAddress(network string, entropy string, index uint32) (string, error) {
	hdPri, err := HdDerive(network, entropy, index)
	if err != nil {
		return "", err
	}
	hdPub, err := HdPrivateToPublic(hdPri, network)
	if err != nil {
		return "", err
	}
	pub, err := HdToEc(hdPub, network)
	if err != nil {
		return "", err
	}
	return GenerateAddress(network, pub)
}
