// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abigen

import (
	"math/big"
	"strings"

	sero "github.com/sero-cash/go-sero"
	"github.com/sero-cash/go-sero/accounts/abi"
	"github.com/sero-cash/go-sero/accounts/abi/bind"
	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/core/types"
	"github.com/sero-cash/go-sero/event"
)

// TestabiABI is the input ABI used to generate the binding from.
const TestabiABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"count\",\"outputs\":[{\"name\":\"\",\"type\":\"uint16\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"registerscode\",\"outputs\":[{\"name\":\"\",\"type\":\"string\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"infon\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"},{\"name\":\"\",\"type\":\"address[]\"},{\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"number\",\"outputs\":[{\"name\":\"a\",\"type\":\"uint32\"},{\"name\":\"b\",\"type\":\"uint64\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"own\",\"outputs\":[{\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"registers\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"blockN\",\"outputs\":[{\"name\":\"\",\"type\":\"uint64\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"addrs\",\"type\":\"address[]\"}],\"name\":\"registers\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"registers\",\"type\":\"address[]\"},{\"name\":\"registerscode\",\"type\":\"string\"},{\"name\":\"decimals\",\"type\":\"uint8\"},{\"name\":\"count\",\"type\":\"uint16\"},{\"name\":\"number\",\"type\":\"uint32\"},{\"name\":\"blockN\",\"type\":\"uint64\"},{\"name\":\"totalSupply\",\"type\":\"uint256\"},{\"name\":\"own\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"registers\",\"type\":\"address[]\"},{\"indexed\":false,\"name\":\"registerscode\",\"type\":\"string\"},{\"indexed\":false,\"name\":\"decimals\",\"type\":\"uint8\"},{\"indexed\":false,\"name\":\"count\",\"type\":\"uint16\"},{\"indexed\":false,\"name\":\"number\",\"type\":\"uint32\"},{\"indexed\":false,\"name\":\"blockN\",\"type\":\"uint64\"},{\"indexed\":false,\"name\":\"totalSupply\",\"type\":\"uint256\"},{\"indexed\":false,\"name\":\"own\",\"type\":\"address\"}],\"name\":\"constructorEvent\",\"type\":\"event\"}]"

// TestabiBin is the compiled bytecode used for deploying new contracts.
const TestabiBin = `0x608060405234801561001057600080fd5b50604051610a3e380380610a3e8339810160409081528151602080840151928401516060850151608086015160a087015160c088015160e089015196890180519099989098019794969395929491939092909161007391600091908b0190610290565b5086516100879060019060208a01906102f5565b5085600260006101000a81548160ff021916908360ff16021790555084600260016101000a81548161ffff021916908361ffff16021790555083600260036101000a81548163ffffffff021916908363ffffffff16021790555082600260076101000a8154816001604060020a0302191690836001604060020a031602179055508160038190555080600460006101000a815481600160a060020a030219169083600160a060020a031602179055507f5d4c6f231ce175a28c0d89b77cb4a74c6a5f61efcc537d422a66c2220c8f8c0388888888888888886040518080602001806020018960ff1660ff1681526020018861ffff1661ffff1681526020018763ffffffff1663ffffffff168152602001866001604060020a03166001604060020a0316815260200185815260200184600160a060020a0316600160a060020a0316815260200183810383528b818151815260200191508051906020019060200280838360005b838110156102055781810151838201526020016101ed565b5050505090500183810382528a818151815260200191508051906020019080838360005b83811015610241578181015183820152602001610229565b50505050905090810190601f16801561026e5780820380516001836020036101000a031916815260200191505b509a505050505050505050505060405180910390a150505050505050506103b0565b8280548282559060005260206000209081019282156102e5579160200282015b828111156102e55782518254600160a060020a031916600160a060020a039091161782556020909201916001909101906102b0565b506102f192915061036f565b5090565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f1061033657805160ff1916838001178555610363565b82800160010185558215610363579182015b82811115610363578251825591602001919060010190610348565b506102f1929150610396565b61039391905b808211156102f1578054600160a060020a0319168155600101610375565b90565b61039391905b808211156102f1576000815560010161039c565b61067f806103bf6000396000f3006080604052600436106100a35763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166306661abd81146100a857806318160ddd146100d45780631d44b1c2146100fb578063313ce5671461018557806359593c70146101b05780638381f58a1461023357806399cdee0e14610271578063a0d3d084146102a2578063a52dc2e714610307578063a818131a14610339575b600080fd5b3480156100b457600080fd5b506100bd6103a2565b6040805161ffff9092168252519081900360200190f35b3480156100e057600080fd5b506100e96103b2565b60408051918252519081900360200190f35b34801561010757600080fd5b506101106103b8565b6040805160208082528351818301528351919283929083019185019080838360005b8381101561014a578181015183820152602001610132565b50505050905090810190601f1680156101775780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b34801561019157600080fd5b5061019a61044d565b6040805160ff9092168252519081900360200190f35b3480156101bc57600080fd5b506101c5610456565b60408051600160a060020a038516815260ff8316918101919091526060602080830182815285519284019290925284516080840191868101910280838360005b8381101561021d578181015183820152602001610205565b5050505090500194505050505060405180910390f35b34801561023f57600080fd5b506102486104dd565b6040805163ffffffff909316835267ffffffffffffffff90911660208301528051918290030190f35b34801561027d57600080fd5b50610286610507565b60408051600160a060020a039092168252519081900360200190f35b3480156102ae57600080fd5b506102b7610516565b60408051602080825283518183015283519192839290830191858101910280838360005b838110156102f35781810151838201526020016102db565b505050509050019250505060405180910390f35b34801561031357600080fd5b5061031c610577565b6040805167ffffffffffffffff9092168252519081900360200190f35b34801561034557600080fd5b506040805160206004803580820135838102808601850190965280855261038e953695939460249493850192918291850190849080828437509497506105929650505050505050565b604080519115158252519081900360200190f35b600254610100900461ffff165b90565b60035490565b60018054604080516020601f600260001961010087891615020190951694909404938401819004810282018101909252828152606093909290918301828280156104435780601f1061041857610100808354040283529160200191610443565b820191906000526020600020905b81548152906001019060200180831161042657829003601f168201915b5050505050905090565b60025460ff1690565b600454600254600080546040805160208084028201810190925282815292946060948694600160a060020a0390921693859360ff90921692918491908301828280156104cb57602002820191906000526020600020905b8154600160a060020a031681526001909101906020018083116104ad575b50505050509150925092509250909192565b6002546301000000810463ffffffff169167010000000000000090910467ffffffffffffffff1690565b600454600160a060020a031690565b6060600080548060200260200160405190810160405280929190818152602001828054801561044357602002820191906000526020600020905b8154600160a060020a03168152600190910190602001808311610550575050505050905090565b600254670100000000000000900467ffffffffffffffff1690565b80516000906105a790829060208501906105b0565b50600192915050565b828054828255906000526020600020908101928215610612579160200282015b82811115610612578251825473ffffffffffffffffffffffffffffffffffffffff1916600160a060020a039091161782556020909201916001909101906105d0565b5061061e929150610622565b5090565b6103af91905b8082111561061e57805473ffffffffffffffffffffffffffffffffffffffff191681556001016106285600a165627a7a72305820ca5b03515993c293931836dd754566ba471e6435977179567ef9c5e54c721d320029`

// DeployTestabi deploys a new Ethereum contract, binding an instance of Testabi to it.
func DeployTestabi(auth *bind.TransactOpts, backend bind.ContractBackend, registers []common.Address, registerscode string, decimals uint8, count uint16, number uint32, blockN uint64, totalSupply *big.Int, own common.Address) (common.Address, *types.Transaction, *Testabi, error) {
	parsed, err := abi.JSON(strings.NewReader(TestabiABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(TestabiBin), backend, registers, registerscode, decimals, count, number, blockN, totalSupply, own)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Testabi{TestabiCaller: TestabiCaller{contract: contract}, TestabiTransactor: TestabiTransactor{contract: contract}, TestabiFilterer: TestabiFilterer{contract: contract}}, nil
}

// Testabi is an auto generated Go binding around an Ethereum contract.
type Testabi struct {
	TestabiCaller     // Read-only binding to the contract
	TestabiTransactor // Write-only binding to the contract
	TestabiFilterer   // Log filterer for contract events
}

// TestabiCaller is an auto generated read-only Go binding around an Ethereum contract.
type TestabiCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestabiTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TestabiTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestabiFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TestabiFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestabiSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TestabiSession struct {
	Contract     *Testabi          // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TestabiCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TestabiCallerSession struct {
	Contract *TestabiCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts  // Call options to use throughout this session
}

// TestabiTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TestabiTransactorSession struct {
	Contract     *TestabiTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// TestabiRaw is an auto generated low-level Go binding around an Ethereum contract.
type TestabiRaw struct {
	Contract *Testabi // Generic contract binding to access the raw methods on
}

// TestabiCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TestabiCallerRaw struct {
	Contract *TestabiCaller // Generic read-only contract binding to access the raw methods on
}

// TestabiTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TestabiTransactorRaw struct {
	Contract *TestabiTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTestabi creates a new instance of Testabi, bound to a specific deployed contract.
func NewTestabi(address common.Address, backend bind.ContractBackend) (*Testabi, error) {
	contract, err := bindTestabi(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Testabi{TestabiCaller: TestabiCaller{contract: contract}, TestabiTransactor: TestabiTransactor{contract: contract}, TestabiFilterer: TestabiFilterer{contract: contract}}, nil
}

// NewTestabiCaller creates a new read-only instance of Testabi, bound to a specific deployed contract.
func NewTestabiCaller(address common.Address, caller bind.ContractCaller) (*TestabiCaller, error) {
	contract, err := bindTestabi(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TestabiCaller{contract: contract}, nil
}

// NewTestabiTransactor creates a new write-only instance of Testabi, bound to a specific deployed contract.
func NewTestabiTransactor(address common.Address, transactor bind.ContractTransactor) (*TestabiTransactor, error) {
	contract, err := bindTestabi(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TestabiTransactor{contract: contract}, nil
}

// NewTestabiFilterer creates a new log filterer instance of Testabi, bound to a specific deployed contract.
func NewTestabiFilterer(address common.Address, filterer bind.ContractFilterer) (*TestabiFilterer, error) {
	contract, err := bindTestabi(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TestabiFilterer{contract: contract}, nil
}

// bindTestabi binds a generic wrapper to an already deployed contract.
func bindTestabi(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(TestabiABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Testabi *TestabiRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Testabi.Contract.TestabiCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Testabi *TestabiRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Testabi.Contract.TestabiTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Testabi *TestabiRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Testabi.Contract.TestabiTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Testabi *TestabiCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Testabi.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Testabi *TestabiTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Testabi.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Testabi *TestabiTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Testabi.Contract.contract.Transact(opts, method, params...)
}

// BlockN is a free data retrieval call binding the contract method 0xa52dc2e7.
//
// Solidity: function blockN() constant returns(uint64)
func (_Testabi *TestabiCaller) BlockN(opts *bind.CallOpts) (uint64, error) {
	var (
		ret0 = new(uint64)
	)
	out := ret0
	err := _Testabi.contract.Call(opts, out, "blockN")
	return *ret0, err
}

// BlockN is a free data retrieval call binding the contract method 0xa52dc2e7.
//
// Solidity: function blockN() constant returns(uint64)
func (_Testabi *TestabiSession) BlockN() (uint64, error) {
	return _Testabi.Contract.BlockN(&_Testabi.CallOpts)
}

// BlockN is a free data retrieval call binding the contract method 0xa52dc2e7.
//
// Solidity: function blockN() constant returns(uint64)
func (_Testabi *TestabiCallerSession) BlockN() (uint64, error) {
	return _Testabi.Contract.BlockN(&_Testabi.CallOpts)
}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() constant returns(uint16)
func (_Testabi *TestabiCaller) Count(opts *bind.CallOpts) (uint16, error) {
	var (
		ret0 = new(uint16)
	)
	out := ret0
	err := _Testabi.contract.Call(opts, out, "count")
	return *ret0, err
}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() constant returns(uint16)
func (_Testabi *TestabiSession) Count() (uint16, error) {
	return _Testabi.Contract.Count(&_Testabi.CallOpts)
}

// Count is a free data retrieval call binding the contract method 0x06661abd.
//
// Solidity: function count() constant returns(uint16)
func (_Testabi *TestabiCallerSession) Count() (uint16, error) {
	return _Testabi.Contract.Count(&_Testabi.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint8)
func (_Testabi *TestabiCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var (
		ret0 = new(uint8)
	)
	out := ret0
	err := _Testabi.contract.Call(opts, out, "decimals")
	return *ret0, err
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint8)
func (_Testabi *TestabiSession) Decimals() (uint8, error) {
	return _Testabi.Contract.Decimals(&_Testabi.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() constant returns(uint8)
func (_Testabi *TestabiCallerSession) Decimals() (uint8, error) {
	return _Testabi.Contract.Decimals(&_Testabi.CallOpts)
}

// Infon is a free data retrieval call binding the contract method 0x59593c70.
//
// Solidity: function infon() constant returns(address, address[], uint8)
func (_Testabi *TestabiCaller) Infon(opts *bind.CallOpts) (common.Address, []common.Address, uint8, error) {
	var (
		ret0 = new(common.Address)
		ret1 = new([]common.Address)
		ret2 = new(uint8)
	)
	out := &[]interface{}{
		ret0,
		ret1,
		ret2,
	}
	err := _Testabi.contract.Call(opts, out, "infon")
	return *ret0, *ret1, *ret2, err
}

// Infon is a free data retrieval call binding the contract method 0x59593c70.
//
// Solidity: function infon() constant returns(address, address[], uint8)
func (_Testabi *TestabiSession) Infon() (common.Address, []common.Address, uint8, error) {
	return _Testabi.Contract.Infon(&_Testabi.CallOpts)
}

// Infon is a free data retrieval call binding the contract method 0x59593c70.
//
// Solidity: function infon() constant returns(address, address[], uint8)
func (_Testabi *TestabiCallerSession) Infon() (common.Address, []common.Address, uint8, error) {
	return _Testabi.Contract.Infon(&_Testabi.CallOpts)
}

// Number is a free data retrieval call binding the contract method 0x8381f58a.
//
// Solidity: function number() constant returns(a uint32, b uint64)
func (_Testabi *TestabiCaller) Number(opts *bind.CallOpts) (struct {
	A uint32
	B uint64
}, error) {
	ret := new(struct {
		A uint32
		B uint64
	})
	out := ret
	err := _Testabi.contract.Call(opts, out, "number")
	return *ret, err
}

// Number is a free data retrieval call binding the contract method 0x8381f58a.
//
// Solidity: function number() constant returns(a uint32, b uint64)
func (_Testabi *TestabiSession) Number() (struct {
	A uint32
	B uint64
}, error) {
	return _Testabi.Contract.Number(&_Testabi.CallOpts)
}

// Number is a free data retrieval call binding the contract method 0x8381f58a.
//
// Solidity: function number() constant returns(a uint32, b uint64)
func (_Testabi *TestabiCallerSession) Number() (struct {
	A uint32
	B uint64
}, error) {
	return _Testabi.Contract.Number(&_Testabi.CallOpts)
}

// Own is a free data retrieval call binding the contract method 0x99cdee0e.
//
// Solidity: function own() constant returns(address)
func (_Testabi *TestabiCaller) Own(opts *bind.CallOpts) (common.Address, error) {
	var (
		ret0 = new(common.Address)
	)
	out := ret0
	err := _Testabi.contract.Call(opts, out, "own")
	return *ret0, err
}

// Own is a free data retrieval call binding the contract method 0x99cdee0e.
//
// Solidity: function own() constant returns(address)
func (_Testabi *TestabiSession) Own() (common.Address, error) {
	return _Testabi.Contract.Own(&_Testabi.CallOpts)
}

// Own is a free data retrieval call binding the contract method 0x99cdee0e.
//
// Solidity: function own() constant returns(address)
func (_Testabi *TestabiCallerSession) Own() (common.Address, error) {
	return _Testabi.Contract.Own(&_Testabi.CallOpts)
}

// Registerscode is a free data retrieval call binding the contract method 0x1d44b1c2.
//
// Solidity: function registerscode() constant returns(string)
func (_Testabi *TestabiCaller) Registerscode(opts *bind.CallOpts) (string, error) {
	var (
		ret0 = new(string)
	)
	out := ret0
	err := _Testabi.contract.Call(opts, out, "registerscode")
	return *ret0, err
}

// Registerscode is a free data retrieval call binding the contract method 0x1d44b1c2.
//
// Solidity: function registerscode() constant returns(string)
func (_Testabi *TestabiSession) Registerscode() (string, error) {
	return _Testabi.Contract.Registerscode(&_Testabi.CallOpts)
}

// Registerscode is a free data retrieval call binding the contract method 0x1d44b1c2.
//
// Solidity: function registerscode() constant returns(string)
func (_Testabi *TestabiCallerSession) Registerscode() (string, error) {
	return _Testabi.Contract.Registerscode(&_Testabi.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_Testabi *TestabiCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _Testabi.contract.Call(opts, out, "totalSupply")
	return *ret0, err
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_Testabi *TestabiSession) TotalSupply() (*big.Int, error) {
	return _Testabi.Contract.TotalSupply(&_Testabi.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() constant returns(uint256)
func (_Testabi *TestabiCallerSession) TotalSupply() (*big.Int, error) {
	return _Testabi.Contract.TotalSupply(&_Testabi.CallOpts)
}

// Registers is a paid mutator transaction binding the contract method 0xa818131a.
//
// Solidity: function registers(addrs address[]) returns(bool)
func (_Testabi *TestabiTransactor) Registers(opts *bind.TransactOpts, addrs []common.Address) (*types.Transaction, error) {
	return _Testabi.contract.Transact(opts, "registers", addrs)
}

// Registers is a paid mutator transaction binding the contract method 0xa818131a.
//
// Solidity: function registers(addrs address[]) returns(bool)
func (_Testabi *TestabiSession) Registers(addrs []common.Address) (*types.Transaction, error) {
	return _Testabi.Contract.Registers(&_Testabi.TransactOpts, addrs)
}

// Registers is a paid mutator transaction binding the contract method 0xa818131a.
//
// Solidity: function registers(addrs address[]) returns(bool)
func (_Testabi *TestabiTransactorSession) Registers(addrs []common.Address) (*types.Transaction, error) {
	return _Testabi.Contract.Registers(&_Testabi.TransactOpts, addrs)
}

// TestabiConstructorEventIterator is returned from FilterConstructorEvent and is used to iterate over the raw logs and unpacked data for ConstructorEvent events raised by the Testabi contract.
type TestabiConstructorEventIterator struct {
	Event *TestabiConstructorEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log    // Log channel receiving the found contract events
	sub  sero.Subscription // Subscription for errors, completion and termination
	done bool              // Whether the subscription completed delivering logs
	fail error             // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *TestabiConstructorEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TestabiConstructorEvent)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(TestabiConstructorEvent)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *TestabiConstructorEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TestabiConstructorEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TestabiConstructorEvent represents a ConstructorEvent event raised by the Testabi contract.
type TestabiConstructorEvent struct {
	Registers     []common.Address
	Registerscode string
	Decimals      uint8
	Count         uint16
	Number        uint32
	BlockN        uint64
	TotalSupply   *big.Int
	Own           common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterConstructorEvent is a free log retrieval operation binding the contract event 0x5d4c6f231ce175a28c0d89b77cb4a74c6a5f61efcc537d422a66c2220c8f8c03.
//
// Solidity: e constructorEvent(registers address[], registerscode string, decimals uint8, count uint16, number uint32, blockN uint64, totalSupply uint256, own address)
func (_Testabi *TestabiFilterer) FilterConstructorEvent(opts *bind.FilterOpts) (*TestabiConstructorEventIterator, error) {

	logs, sub, err := _Testabi.contract.FilterLogs(opts, "constructorEvent")
	if err != nil {
		return nil, err
	}
	return &TestabiConstructorEventIterator{contract: _Testabi.contract, event: "constructorEvent", logs: logs, sub: sub}, nil
}

// WatchConstructorEvent is a free log subscription operation binding the contract event 0x5d4c6f231ce175a28c0d89b77cb4a74c6a5f61efcc537d422a66c2220c8f8c03.
//
// Solidity: e constructorEvent(registers address[], registerscode string, decimals uint8, count uint16, number uint32, blockN uint64, totalSupply uint256, own address)
func (_Testabi *TestabiFilterer) WatchConstructorEvent(opts *bind.WatchOpts, sink chan<- *TestabiConstructorEvent) (event.Subscription, error) {

	logs, sub, err := _Testabi.contract.WatchLogs(opts, "constructorEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TestabiConstructorEvent)
				if err := _Testabi.contract.UnpackLog(event, "constructorEvent", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}
