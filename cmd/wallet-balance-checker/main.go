package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/basel-ax/lucky-eth/entity"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Blockchain struct to hold network information
type Blockchain struct {
	Name   string
	RpcURL string
}

var blockchains []Blockchain

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// --- Database Setup ---
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Migrate the schema
	db.AutoMigrate(&entity.WalletBalance{})

	// --- Telegram Bot Setup ---
	botToken := os.Getenv("TELEGRAM_APP_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_APP_BOT_TOKEN environment variable not set")
	}
	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	if chatIDStr == "" {
		log.Fatal("TELEGRAM_CHAT_ID environment variable not set")
	}
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Fatalf("Invalid TELEGRAM_CHAT_ID: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// --- Blockchain RPC Setup ---
	blockchains = []Blockchain{
		{Name: "Ethereum", RpcURL: os.Getenv("ETH_RPC_URL")},
		{Name: "Arbitrum", RpcURL: os.Getenv("ARBITRUM_RPC_URL")},
		{Name: "Base", RpcURL: os.Getenv("BASE_RPC_URL")},
		{Name: "BSC", RpcURL: os.Getenv("BSC_RPC_URL")},
	}

	for _, b := range blockchains {
		if b.RpcURL == "" {
			log.Fatalf("%s_RPC_URL environment variable not set", b.Name)
		}
	}

	// --- Main Logic ---
	log.Println("Starting wallet balance check...")
	var wallets []entity.WalletBalance
	// Find wallets that haven't been notified yet
	if err := db.Where("is_notified = ?", false).Find(&wallets).Error; err != nil {
		log.Fatalf("Failed to fetch wallets: %v", err)
	}

	log.Printf("Found %d wallets to check.", len(wallets))

	for i := range wallets {
		processWallet(db, bot, chatID, &wallets[i])
	}

	log.Println("Wallet balance check finished.")
}

// processWallet derives address, checks balances, and sends notification if needed
func processWallet(db *gorm.DB, bot *tgbotapi.BotAPI, chatID int64, wallet *entity.WalletBalance) {
	// 1. Derive address if it's not already set
	if wallet.Address == "" {
		address, err := deriveAddress(wallet.Mnemonic)
		if err != nil {
			log.Printf("Failed to derive address for mnemonic ID %d: %v", wallet.ID, err)
			return
		}
		wallet.Address = address
		if err := db.Save(wallet).Error; err != nil {
			log.Printf("Failed to save address for wallet %d: %v", wallet.ID, err)
			return // Continue to next wallet if save fails
		}
		log.Printf("Derived address %s for wallet ID %d", wallet.Address, wallet.ID)
	}

	// 2. Check balance on each blockchain
	for _, chain := range blockchains {
		balance, err := checkBalance(chain.RpcURL, wallet.Address)
		if err != nil {
			log.Printf("Error checking balance on %s for %s: %v", chain.Name, wallet.Address, err)
			continue // Try next blockchain
		}

		// If balance is found
		if balance.Cmp(big.NewInt(0)) > 0 {
			log.Printf("FOUND BALANCE on %s for address %s: %s", chain.Name, wallet.Address, balance.String())

			// Update wallet in DB
			now := time.Now()
			wallet.Balance = balance.String()
			wallet.BalanceUpdatedAt = &now
			wallet.IsNotified = true
			if err := db.Save(wallet).Error; err != nil {
				log.Printf("Failed to update wallet %d after finding balance: %v", wallet.ID, err)
				// Don't return, still try to send notification
			}

			// Send Telegram notification
			debankURL := fmt.Sprintf("https://debank.com/profile/%s", wallet.Address)
			messageText := fmt.Sprintf("💰 Found a wallet with a balance!\n\nChain: %s\nAddress: %s\n\nView on DeBank:\n%s", chain.Name, wallet.Address, debankURL)
			msg := tgbotapi.NewMessage(chatID, messageText)
			_, err := bot.Send(msg)
			if err != nil {
				log.Printf("Failed to send Telegram notification for wallet %s: %v", wallet.Address, err)
			} else {
				log.Printf("Successfully sent Telegram notification for wallet %s", wallet.Address)
			}

			// Once balance is found and notified, we are done with this wallet
			return
		}
		log.Printf("Zero balance on %s for %s", chain.Name, wallet.Address)
	}
}

// deriveAddress generates an Ethereum address from a BIP39 mnemonic.
func deriveAddress(mnemonic string) (string, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return "", fmt.Errorf("failed to create wallet from mnemonic: %w", err)
	}

	// Standard Ethereum derivation path
	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, false)
	if err != nil {
		return "", fmt.Errorf("failed to derive account: %w", err)
	}

	return account.Address.Hex(), nil
}

// checkBalance connects to a given RPC endpoint and fetches the balance of an address.
func checkBalance(rpcURL string, address string) (*big.Int, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to client %s: %w", rpcURL, err)
	}
	defer client.Close()

	account := common.HexToAddress(address)
	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}
