pragma solidity ^0.4.16;


contract testabi {


    address[] private _registers;
    string private _registerscode;
    uint8 private _decimals;
    uint16 private _count;
    uint32 private _number;
    uint64 private _block;
    uint256 _totalSupply;
    address private _own;

    event constructorEvent( address[] registers,
    string registerscode,
    uint8 decimals,
    uint16 count,
    uint32 number,
    uint64 blockN,
    uint256 totalSupply,
    address own);

    constructor(
        address[] registers,
    string registerscode,
    uint8 decimals,
    uint16 count,
    uint32 number,
    uint64 blockN,
    uint256 totalSupply,
    address own) public{
        _registers = registers;
    _registerscode=registerscode;
    _decimals=decimals;
    _count=count;
    _number=number;
    _block=blockN;
    _totalSupply=totalSupply;
    _own=own;
    emit constructorEvent(registers,registerscode,decimals,count,number,blockN,totalSupply,own);
    }

    function registers( address[] addrs) public returns(bool){
        _registers=addrs;
        return true;
    }

    function infon() public view returns(address,address[],uint8){
        return (_own,_registers,_decimals);
    }

    /**
     * @return the number of decimals of the token.
     */
    function decimals() public view returns (uint8) {
        return _decimals;
    }

    function totalSupply() public view returns (uint256) {
        return _totalSupply;
    }

    function registerscode() public view returns (string) {
        return _registerscode;
    }

    function number() public view returns ( uint32 a, uint64 b) {
        a=_number;
        b=_block;
        return;
    }

    function blockN() public view returns (uint64) {
        return _block;
    }

    function registers() public view returns (address[]) {
        return _registers;
    }
    function count() public view returns (uint16) {
        return _count;
    }

    function own() public view returns (address) {
        return _own;
    }

}