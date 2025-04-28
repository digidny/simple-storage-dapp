package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/digidny/simple-storage-dapp/backend/internal/storage"
	"github.com/joho/godotenv"
	"github.com/jumbochain/jumbochain-go/accounts/abi/bind"
	"github.com/jumbochain/jumbochain-go/common"
	"github.com/jumbochain/jumbochain-go/core/types"
	"github.com/jumbochain/jumbochain-go/crypto"
	"github.com/jumbochain/jumbochain-go/jumboclient"
)

// Ensure this matches the contract ABI.  Use `abigen` to generate.
//go:generate abigen --abi=../build/SimpleStorage.abi --pkg=storage --out=./internal/contract/storage/storage.go

// Contract binding package.
type Storage struct {
	address common.Address
	abi     string
}

func main() {
	// Load environment variables from .env file.
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// Connect to the Ethereum client.  Use the URL from the environment.
	rpcURL := os.Getenv("RPC_URL") // e.g., "http://localhost:8545"
	if rpcURL == "" {
		log.Fatal("ETH_URL environment variable not set")
	}
	client, err := jumboclient.DialContext(context.Background(), rpcURL)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 1. Deploy the contract (or use an existing address).
	//    For this example, we assume it's already deployed.  In a real
	//    application, you'd use the `DeployContract` function from
	//    the `bind` package.  We'll fetch the address from the env.
	contractAddressStr := os.Getenv("CONTRACT_ADDRESS")
	if contractAddressStr == "" {
		log.Fatal("CONTRACT_ADDRESS environment variable not set")
	}
	contractAddress := common.HexToAddress(contractAddressStr)
	fmt.Println("Contract Address:", contractAddress)

	// 2. Create an instance of the contract binding.
	instance, err := storage.NewSimpleStorage(contractAddress, client)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Get the initial value.
	initialValue, err := instance.Get(nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Initial value:", initialValue)

	// 4. Set a new value.
	auth, err := getTransactionAuthorizer(client) // Get auth for making tx
	if err != nil {
		log.Fatal(err)
	}

	newValue := big.NewInt(150)

	// 5. Estimate gas *before* sending the transaction.
	gas, err := instance.EstimateGas(&bind.CallOpts{From: auth.From, Context: context.Background()}, "set", newValue)
	if err != nil {
		log.Fatal("Error estimating gas:", err)
	}
	fmt.Println("Estimated gas:", gas)
	auth.GasLimit = gas + 20000 // add a buffer of 20000

	tx, err := instance.Set(auth, newValue)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Set transaction hash: %s\n", tx.Hash().Hex())

	// Wait for the transaction to be mined (optional, but useful for demonstration).
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("Transaction %s mining failed: %v", tx.Hash().Hex(), err)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		log.Fatalf("Transaction %s failed", tx.Hash().Hex())
	}
	fmt.Printf("Transaction mined in block %d\n", receipt.BlockNumber.Uint64())

	// 6. Get the updated value.
	updatedValue, err := instance.Get(nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updated value:", updatedValue)

	// 7. Call the add function
	addValue := big.NewInt(10)
	authAdd, err := getTransactionAuthorizer(client)
	if err != nil {
		log.Fatal(err)
	}
	gasAdd, err := instance.EstimateGas(&bind.CallOpts{From: authAdd.From, Context: context.Background()}, "add", addValue)
	if err != nil {
		log.Fatal("Error estimating gas for add:", err)
	}
	fmt.Println("Estimated gas for add:", gasAdd)
	authAdd.GasLimit = gasAdd + 20000

	txAdd, err := instance.Add(authAdd, addValue)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Add transaction hash: %s\n", txAdd.Hash().Hex())
	receiptAdd, err := bind.WaitMined(context.Background(), client, txAdd)
	if err != nil {
		log.Fatalf("Transaction %s mining failed: %v", txAdd.Hash().Hex(), err)
	}
	if receiptAdd.Status == types.ReceiptStatusFailed {
		log.Fatalf("Transaction %s failed", txAdd.Hash().Hex())
	}

	newValueAfterAdd, err := instance.Get(nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("New Value After Add:", newValueAfterAdd)
}

// getTransactionAuthorizer creates a `bind.TransactOptions` struct
// for signing and submitting transactions.  It reads the private key
// from the environment.
func getTransactionAuthorizer(client *jumboclient.Client) (*bind.TransactOptions, error) {
	privateKeyHex := os.Getenv("PRIVATE_KEY") // The sender's private key
	if privateKeyHex == "" {
		return nil, fmt.Errorf("PRIVATE_KEY environment variable not set")
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	// Get the sender's address from the private key.
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Get the nonce for the sender's address.
	nonce, err := client.PendingNonceAt(context.Background(), address)
	if err != nil {
		return nil, err
	}

	// Chain ID is needed for EIP-155 signing.  Get it from the client.
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	// Create a new `bind.TransactOptions` struct.  This struct holds
	// all the necessary information for signing and sending a transaction.
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // Amount to send (in wei).  Set to 0 for contract calls.
	auth.GasLimit = uint64(3000000) //  Maximum gas allowed for the transaction.
	// GasPrice is set automatically by ethclient in newer versions, but
	// you might need to set it manually for older versions or for more
	// control.  We'll leave it commented out for now and let geth handle it.
	// auth.GasPrice = gasPrice

	return auth, nil
}
