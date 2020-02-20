package tests

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/sero-cash/go-sero/common"
	"github.com/sero-cash/go-sero/tests/abigen"

	"github.com/sero-cash/go-sero/accounts/abi/bind"
	"github.com/sero-cash/go-sero/seroclient"
)

const keyStr = `
{
	"address": "3bG34svmC3pwCazHtsV7t1p5ihvaSBu2pA27CsnrsdJY5yCkf1ZqAKGArV6gHpeDHieiXScL5FKdgJTgHBiwe1uU",
	"tk": "3bG34svmC3pwCazHtsV7t1p5ihvaSBu2pA27CsnrsdJY1ENvjxfeC4vGQfigJ54nX4od4BM5tsT48WbnDpFFGAJn",
	"crypto": {
		"cipher": "aes-128-ctr",
		"ciphertext": "3138a97b25721f0829106aa652d3ea76483078be253bacb30449c0e6f970a097",
		"cipherparams": {
			"iv": "b29008189567795220a84bf7d8dfaf01"
		},
		"kdf": "scrypt",
		"kdfparams": {
			"dklen": 32,
			"n": 262144,
			"p": 1,
			"r": 8,
			"salt": "1a238ac93b466659368dbd89cb5558fbc09a8fde256c7e3daef5b778857e456c"
		},
		"mac": "a23300101025fd7ed170f69a283655688e0d05e073177f252fb9326fc626127b"
	},
	"id": "e5b057ac-2119-4cf5-9584-047828207eb3",
	"version": 1,
	"at": 0
}
`

func TestDeploy(t *testing.T) {
	conn, err := seroclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	//value := big.NewInt(0).Mul(big.NewInt(1000000000), big.NewInt(1000000000))
	auth, err := bind.NewTransactor(strings.NewReader(keyStr), "123456", nil)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}
	// Deploy a new awesome contract for the binding demo

	registers := []common.Address{common.Base58ToAddress("iQe7iX45yGZeQnxjdxgo7SZJfyWpWPaKmQJyR5UuPM2Aq1BJDUuFvh5bLtquiwuzdWCxmYXMW1LbhHZC4vDBgAhuRxZGtQr7zoTjkHH6Z1vMzUQNGjBAXE5nkufjjAzfWBZ")}

	registerscode := "registerscode"
	decimals := uint8(8)
	count := uint16(16)
	number := uint32(32)
	blockN := uint64(64)
	totalSupply := big.NewInt(1000000)
	own := common.Base58ToAddress("wg5cNrx3qeh6xea8nNQqrWEu5etbCauo7Ek75XhmWQnmQAyesBjVMbDwWLZdeUYfXj4eB7YAPBmqAyUcBgYuCFBRdHjuE8PQsaEPp5MTXpE7zBt6wcsEyzuguGd4NMw3i7H")

	_, tx, testAbi, err := abigen.DeployTestabi(auth, conn, registers, registerscode, decimals, count, number, blockN, totalSupply, own)
	if err != nil {
		log.Fatalf("Failed to deploy new token contract: %v", err)
	}
	fmt.Printf("Transaction waiting to be mined: 0x%x\n\n", tx.Hash())
	startTime := time.Now()
	fmt.Printf("TX start @:%s", time.Now())
	ctx := context.Background()
	addressAfterMined, err := bind.WaitDeployed(ctx, conn, tx)
	if err != nil {
		log.Fatalf("failed to deploy contact when mining :%v", err)
	}
	fmt.Printf("tx mining take time:%s\n", time.Now().Sub(startTime))
	log.Printf("mined address :%s", addressAfterMined.String())

	name, err := testAbi.Own(&bind.CallOpts{Pending: true})
	if err != nil {
		log.Fatalf("Failed to retrieve pending name: %v", err)
	}
	fmt.Println("Pending name:", name)
}

func TestCall(t *testing.T) {
	conn, err := seroclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	contractAddress := common.Base58ToAddress("25rwV92apLVTVcTyjhAKx2izEmEvpVv5kwoDBi1anWC2HtChJNVg3o4uHTaMxAPmYTUTRM1kLvNDdZLiFXgCJVsg")
	testAbi, err := abigen.NewTestabi(contractAddress, conn)
	if err != nil {
		log.Fatalf("Failed to NewTestabi : %v", err)
	}
	e, err := testAbi.Number(&bind.CallOpts{})
	if err != nil {
		log.Fatalf("Failed to retrieve pending name: %v", err)
	}
	fmt.Println(e.A, e.B)
}

func TestExcute(t *testing.T) {
	conn, err := seroclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	//value := big.NewInt(0).Mul(big.NewInt(1000000000), big.NewInt(1000000000))
	auth, err := bind.NewTransactor(strings.NewReader(keyStr), "123456", nil)
	if err != nil {
		log.Fatalf("Failed to create authorized transactor: %v", err)
	}
	// Deploy a new awesome contract for the binding demo

	registers := []common.Address{common.Base58ToAddress("2WmLwyPZ7e8E1pSrQavxXGtXbHyqQGuWAWnZdPGG5UQ7y1epSFJWNMdtmrfbYoHKviQ42AB872Ccp3EP3uzpWxCSxx9ZBn6a6E3fbGQEMT6xETV4xKEc8Nf5VMYRZYYXszHh")}

	contractAddress := common.Base58ToAddress("25rwV92apLVTVcTyjhAKx2izEmEvpVv5kwoDBi1anWC2HtChJNVg3o4uHTaMxAPmYTUTRM1kLvNDdZLiFXgCJVsg")
	testAbi, err := abigen.NewTestabi(contractAddress, conn)
	if err != nil {
		log.Fatalf("Failed to NewTestabi : %v", err)
	}
	e, err := testAbi.Registers(auth, registers)
	if err != nil {
		log.Fatalf("Failed to retrieve pending name: %v", err)
	}
	fmt.Println(e)
}
