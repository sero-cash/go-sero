package keystore

import (
	"fmt"
	"testing"

	"github.com/tyler-smith/go-bip39"

	"github.com/sero-cash/go-sero/common/hexutil"
	"github.com/sero-cash/go-sero/crypto"
)

func TestNewKeyFromString(t *testing.T) {

}

func TestNewMaster(t *testing.T) {
	seed, _ := hexutil.Decode("0x747f302d9c916698912d5f70be53a6cf53bc495803a5523d3a7c3afa2afba94ec3803f838b3e1929ab5481f9da35441372283690fdcf27372c38f40ba134fe03")
	hdKey, _ := NewMaster(seed)
	fmt.Println(hdKey.PrivateHDkey())
	fmt.Println(hdKey.PublicHDkey())

	if hdKey.PrivateHDkey() != "xprv9s21ZrQH143K4KqQx9Zrf1eN8EaPQVFxM2Ast8mdHn7GKiDWzNEyNdduJhWXToy8MpkGcKjxeFWd8oBSvsz4PCYamxR7TX49pSpp3bmHVAY" {
		t.Error("not success PrivateHDkey")
	}
	if hdKey.PublicHDkey() != "xpub661MyMwAqRbcGout4B6s29b6gGQsowyoiF6UgXBEr7eFCWYfXuZDvRxP9zEh1Kwq3TLqDQMbkbaRpSnoC28oWvjLeshoQz1StZ9YHM1EpcJ" {
		t.Error("not success PublicHDkey")
	}

}

func TestHDKey_DriverPath(t *testing.T) {

	seed, _ := hexutil.Decode("0x747f302d9c916698912d5f70be53a6cf53bc495803a5523d3a7c3afa2afba94ec3803f838b3e1929ab5481f9da35441372283690fdcf27372c38f40ba134fe03")
	hdKey, _ := NewMaster(seed)
	ck, _ := hdKey.DriverPath("m")
	if ck.PrivateHDkey() != "xprv9s21ZrQH143K4KqQx9Zrf1eN8EaPQVFxM2Ast8mdHn7GKiDWzNEyNdduJhWXToy8MpkGcKjxeFWd8oBSvsz4PCYamxR7TX49pSpp3bmHVAY" {
		t.Error("not success PrivateHDkey")
	}
	if ck.PublicHDkey() != "xpub661MyMwAqRbcGout4B6s29b6gGQsowyoiF6UgXBEr7eFCWYfXuZDvRxP9zEh1Kwq3TLqDQMbkbaRpSnoC28oWvjLeshoQz1StZ9YHM1EpcJ" {
		t.Error("not success PublicHDkey")
	}
	ck, _ = hdKey.DriverPath("m/44'/0'/0/0")
	fmt.Println(ck.PrivateHDkey())
	if ck.PrivateHDkey() != "xprvA1ErCzsuXhpB8iDTsbmgpkA2P8ggu97hMZbAXTZCdGYeaUrDhyR8fEw47BNEgLExsWCVzFYuGyeDZJLiFJ9kwBzGojQ6NB718tjVJrVBSrG" {
		t.Error("not DriverPath m/44'/0'/0/1")
	}
	ck, _ = hdKey.DriverChild(1)
	fmt.Println(ck.PrivateHDkey())
	if ck.PrivateHDkey() != "xprv9vYSvrg3eR5FaKbQE4Ao2vHdyvfFL27aWMyH6X818mKWMsqqQZAN6HmRqYDGDPLArzaqbLExRsxFwtx2B2X2QKkC9uoKsiBNi22tLPKZHNS" {
		t.Error("not success DriverChild")
	}
}

func TestWallet(t *testing.T) {
	seed, _ := hexutil.Decode("0x747f302d9c916698912d5f70be53a6cf53bc495803a5523d3a7c3afa2afba94ec3803f838b3e1929ab5481f9da35441372283690fdcf27372c38f40ba134fe03")
	hdKey, _ := NewMaster(seed)
	fmt.Println(hexutil.Encode(hdKey.key))
	if hexutil.Encode(hdKey.key) != "0x26cc9417b89cd77c4acdbe2e3cd286070a015d8e380f9cd1244ae103b7d89d81" {
		t.Error("not success wallet")
	}
	key, err := crypto.ToECDSA(hdKey.key)
	if err != nil {
		t.Error("not success wallet pub")
	}

	PublicKey := crypto.FromECDSAPub(&key.PublicKey)
	fmt.Println(hexutil.Encode(PublicKey[1:]))
	if hexutil.Encode(PublicKey[1:]) != "0x0639797f6cc72aea0f3d309730844a9e67d9f1866e55845c5f7e0ab48402973defa5cb69df462bcc6d73c31e1c663c225650e80ef14a507b203f2a12aea55bc1" {
		t.Error("not success wallet pub")
	}
}

func TestCreateBip39HDKey(t *testing.T) {
	m, hdkey, _ := CreateBip39HDKey()
	fmt.Println(m)
	fmt.Println(hdkey.PrivateHDkey())
	fmt.Println(hexutil.Encode(hdkey.chainCode))

}

func TestNewFromMnemonic(t *testing.T) {
	hdkey, _ := NewFromMnemonic("elite winner autumn cash judge once paddle lawn wrap age shoe summer awesome together motion mammal neck decrease steel excess awake random cram song")
	if hdkey.PrivateHDkey() != "xprv9s21ZrQH143K4Y7kqh8YDnt7xoxPLwppfZ7pKHBPTASt9NMwFgkNWWvj43a3pDPJe6kQSTB7C5RhYXnxA6xRQtBaJue4fxZSp6bKgygp2cd" {
		t.Error("not success NewFromMnemonic")
	}
	//0xf8f4f6fff6dc83ebc27108445895fabb7536f4c968c1b2525af8e8fef1a25db4
	fmt.Println(hexutil.Encode(hdkey.chainCode))

}
func Test_mnemonic(t *testing.T) {
	entropy, err := bip39.NewEntropy(256)

	fmt.Println(hexutil.Encode(entropy))

	if err != nil {
		t.Error(err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		t.Error(err)
	}

	seed, err := bip39.EntropyFromMnemonic(mnemonic)

	if err != nil {
		t.Error(err)
	}
	fmt.Println(hexutil.Encode(seed))

}
