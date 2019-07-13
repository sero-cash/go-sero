package app

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/sero-cash/go-czero-import/keys"
	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/accounts/keystore"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/common/address"
	"github.com/sero-cash/go-sero/crypto"
	"github.com/sero-cash/go-sero/light-wallet/common/logex"
	"github.com/sero-cash/go-sero/light-wallet/common/transport"
	"github.com/tyler-smith/go-bip39"
	"io/ioutil"
	"math/big"
	"net/http"
	"sort"
)

//keystore file upload
const maxUploadSize = 1 * 1024 * 2014 // 2 MB

type Service interface {
	NewAccountWithMnemonic(passphrase string) (map[string]string, error)
	UploadKeystoreHandler() http.HandlerFunc
	ImportAccountFromMnemonic(mnemonic, password string) (map[string]string, error)
	ImportAccountFromRawKey(privkey, password string) (map[string]string, error)
	ExportMnemonic(addressStr, password string) (string, error)
	AccountList() (accountResps)
	AccountDetail(pkStr string) accountResp
	AccountBalance(pkStr string) map[string]*big.Int
	TXNum(pkStr string) map[string]uint64
	TXList(pkStr string, request transport.PageRequest) (utxosResp, error)

	Transfer(from, to, currency, amount, gasPrice, pwd string) (hash string, err error)
}

func NewPrivateAccountAPI() Service {
	return &PrivateAccountAPI{
		SL: CurrentClient(),
	}
}

type PrivateAccountAPI struct {
	SL *SEROLight
}

func (s *PrivateAccountAPI) ExportMnemonic(addressStr, password string) (string, error) {
	return fetchKeystore(s.SL.accountManager).ExportMnemonic(accounts.Account{Address: address.Base58ToAccount(addressStr)}, password)
}

// fetchKeystore retrives the encrypted keystore from the account manager.
func fetchKeystore(am *accounts.Manager) *keystore.KeyStore {
	return am.Backends(keystore.KeyStoreType)[0].(*keystore.KeyStore)
}

func (s *PrivateAccountAPI) NewAccountWithMnemonic(passphrase string) (map[string]string, error) {
	blockNum := s.SL.CurrentBlock()

	mnemonic, acc, err := fetchKeystore(s.SL.accountManager).NewAccountWithMnemonic(passphrase, blockNum)
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	result["mnemonic"] = mnemonic
	result["address"] = acc.Address.Base58()
	return result, nil
}

func (s *PrivateAccountAPI) ImportAccountFromMnemonic(mnemonic, password string) (map[string]string, error) {
	_, err := bip39.MnemonicToByteArray(mnemonic)
	if err != nil {
		return nil, err
	}
	seed, err := bip39.EntropyFromMnemonic(mnemonic)
	if err != nil {
		return nil, err
	}
	if len(seed) != 32 {
		return nil, errors.New("EntropyFromMnemonic error seed not 256bits")
	}
	key, err := crypto.ToECDSA(seed[:32])
	if err != nil {
		return nil, err
	}
	acc, err := fetchKeystore(s.SL.accountManager).ImportECDSA(key, password)
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	result["address"] = acc.Address.Base58()
	return result, nil
}

func (s *PrivateAccountAPI) ImportAccountFromRawKey(privkey, password string) (map[string]string, error) {
	key, err := crypto.HexToECDSA(privkey)
	if err != nil {
		return nil, err
	}
	acc, err := fetchKeystore(s.SL.accountManager).ImportECDSA(key, password)
	if err != nil {
		return nil, err
	}
	result := map[string]string{}
	result["address"] = acc.Address.Base58()
	return result, nil
}

type accountResp struct {
	PK        string
	MainPKr   string
	Balance   map[string]*big.Int
	UtxoNums  map[string]uint64
	PkrBase58 []string
	index     int8
}

type accountResps []accountResp

func (acrs accountResps) Len() int {
	return len(acrs)
}
func (acrs accountResps) Less(i, j int) bool {
	return acrs[i].index < acrs[j].index
}
func (acrs accountResps) Swap(i, j int) {
	acrs[i], acrs[j] = acrs[j], acrs[i]
}

func (s *PrivateAccountAPI) AccountList() (accountListResps accountResps) {
	s.SL.accounts.Range(func(key, value interface{}) bool {
		pk := key.(keys.Uint512)
		account := value.(*Account)

		pkrs := []string{}

		pkrAndIndexs := s.SL.getPKrsForQueryByPk(pk)
		for _, pkrAndIndex := range pkrAndIndexs {
			pkrs = append(pkrs, base58.Encode(pkrAndIndex.pkr[:]))
		}
		balance := s.SL.GetBalances(pk)

		accountListResp := accountResp{PK: base58.Encode(pk[:]), MainPKr: base58.Encode(account.mainPkr[:]), Balance: balance, UtxoNums: account.utxoNums, PkrBase58: pkrs, index: account.index}
		accountListResps = append(accountListResps, accountListResp)
		return true
	})

	sort.Sort(accountListResps)

	return accountListResps
}

func (s *PrivateAccountAPI) AccountDetail(pkStr string) (account accountResp) {
	pk := address.Base58ToAccount(pkStr)
	if ac := s.SL.getAccountByPk(*pk.ToUint512()); ac != nil {
		pkrs := []string{}
		pkrAndIndexs := s.SL.getPKrsForQueryByPk(*pk.ToUint512())
		for _, pkrAndIndex := range pkrAndIndexs {
			pkrs = append(pkrs, base58.Encode(pkrAndIndex.pkr[:]))
		}
		balance := s.SL.GetBalances(*pk.ToUint512())
		account := accountResp{PK: base58.Encode(pk[:]), MainPKr: base58.Encode(ac.mainPkr[:]), Balance: balance, UtxoNums: ac.utxoNums, PkrBase58: pkrs}

		return account
	}
	return account
}

func (s *PrivateAccountAPI) AccountBalance(pkStr string) map[string]*big.Int {
	pk := address.Base58ToAccount(pkStr)
	return s.SL.GetBalances(*pk.ToUint512())
}

type utxoResp struct {
	Pkr    string
	Root   keys.Uint256
	TxHash keys.Uint256
	Nil    keys.Uint256
	Num    uint64
	Asset  assetResp
	IsZ    bool
	flag   int
}

type assetResp struct {
	Tkn tknResp
	Tkt tktResp
}

type tknResp struct {
	Currency string
	Value    big.Int
}

type tktResp struct {
	Category string
	Value    string
}

type utxosResp []utxoResp

func (u utxosResp) Len() int {
	return len(u)
}
func (u utxosResp) Less(i, j int) bool {
	return u[i].Num > u[j].Num
}
func (u utxosResp) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

func (s *PrivateAccountAPI) TXList(pkStr string, request transport.PageRequest) (utxos utxosResp, err error) {
	pk := address.Base58ToAccount(pkStr)

	if records, err := s.SL.GetRecordsByPk(pk.ToUint512(), uint64(request.PageSize*(request.PageNo-1)), uint64(request.PageNo*request.PageSize)); err != nil {
		return utxos, err
	} else {
		for _, record := range records {
			tkn := tknResp{}
			tkt := tktResp{}
			if record.Asset.Tkn != nil {
				currency := common.BytesToString(record.Asset.Tkn.Currency[:])
				tkn = tknResp{Currency: currency, Value: *record.Asset.Tkn.Value.ToIntRef()}
			}
			if record.Asset.Tkt != nil {
				tkt = tktResp{Category: string(record.Asset.Tkt.Category[:]), Value: string(record.Asset.Tkt.Value[:])}
			}
			utxo := utxoResp{
				Pkr:    base58.Encode(record.Pkr[:]),
				Root:   record.Root,
				TxHash: record.TxHash,
				Num:    record.Num,
				Asset:  assetResp{Tkn: tkn, Tkt: tkt},
				IsZ:    record.IsZ,
			}
			utxos = append(utxos, utxo)
		}

		sort.Sort(utxos)
	}
	return
}

func (s *PrivateAccountAPI) Transfer(from, to, currency, amountStr, gasPriceStr, password string) (hash string, err error) {

	amount, err := NewBigIntFromString(amountStr, 10);
	if err != nil {
		return hash, err
	} else {
		if amount.Sign() < 0 {
			return hash, fmt.Errorf("amount < 0")
		}
	}

	gasPrice, err := NewBigIntFromString(gasPriceStr, 10);
	if err != nil {
		return hash, err
	} else {
		if gasPrice.Sign() < 0 {
			return hash, fmt.Errorf("gasPrice < 0")
		}
	}
	if toBytes := base58.Decode(to); len(toBytes) != 96 {
		return hash, fmt.Errorf("Invalid colleaction address ")
	}
	h, err := s.SL.CommitTx(from, to, currency, password, amount, gasPrice)
	if err != nil {
		return hash, err
	}
	return string(h[:]), nil
}

func (s *PrivateAccountAPI) TXNum(pkStr string) map[string]uint64 {
	pk := address.Base58ToAccount(pkStr)
	return s.SL.GetUtxoNum(*pk.ToUint512())
}

func renderError(w http.ResponseWriter, errcode string, code int) {
	//w.WriteHeader(code)
	w.Write([]byte(errcode))
}

func (s *PrivateAccountAPI) UploadKeystoreHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")
		if r.Method == "OPTIONS" {
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			renderError(w, "FILE_TOO_BIG", http.StatusOK)
			return
		}
		file, _, err := r.FormFile("uploadFile")
		passphrase := r.FormValue("passphrase")
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusOK)
			return
		}
		defer file.Close()
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusOK)
			return
		}
		key, err := keystore.DecryptKey(fileBytes, passphrase)
		if err != nil {
			//renderError(w, "INVALID_FILE_TYPE", http.StatusOK)
			w.Write([]byte("INVALID_FILE_TYPE"))
			return
		}

		if err := ioutil.WriteFile(GetKeystorePath()+"/"+key.Address.String(), fileBytes, 0600); err != nil {
			renderError(w, "INVALID_FILE", http.StatusOK)
			return
		}

		logex.Infof("Import account successful. address=[%s]", key.Address)
		w.Write([]byte("SUCCESS"))
		return
	})
}
