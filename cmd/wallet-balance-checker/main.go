package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
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

// Custom logger that can be silenced in prod mode
type Logger struct {
	logger *log.Logger
	silent bool
}

func (l *Logger) Print(v ...interface{}) {
	if !l.silent {
		l.logger.Print(v...)
	}
}

func (l *Logger) Printf(format string, v ...interface{}) {
	if !l.silent {
		l.logger.Printf(format, v...)
	}
}

func (l *Logger) Fatal(v ...interface{}) {
	if !l.silent {
		l.logger.Fatal(v...)
	}
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	if !l.silent {
		l.logger.Fatalf(format, v...)
	}
	os.Exit(1)
}

func (l *Logger) Println(v ...interface{}) {
	if !l.silent {
		l.logger.Println(v...)
	}
}

var logger Logger

// RowCountStats holds statistics about processed records
type RowCountStats struct {
	TotalWallets   int
	AddressDerived int
	BalanceUpdated int
	Notifications  int
	Errors         int
}

// Lock file path for preventing concurrent execution
const lockFilePath = "/tmp/wallet-balance-checker.lock"

func acquireLock() (*os.File, error) {
	lockFile, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("another instance is already running")
		}
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	return lockFile, nil
}

func releaseLock(lockFile *os.File) {
	if lockFile != nil {
		lockFile.Close()
		os.Remove(lockFilePath)
	}
}

func main() {
	// Parse command line flags
	prodMode := flag.Bool("prod", false, "Run in production mode (silent console output, send telegram summary)")
	flag.Parse()

	// Initialize logger
	logger = Logger{logger: log.New(os.Stdout, "", log.LstdFlags), silent: *prodMode}

	// Acquire lock to prevent concurrent execution
	lockFile, err := acquireLock()
	if err != nil {
		if *prodMode {
			// In prod mode, just exit silently if already running
			os.Exit(0)
		}
		logger.Fatal("Failed to acquire lock: ", err)
	}
	defer releaseLock(lockFile)

	// Load environment variables from .env file
	// Try multiple locations since cron runs from / or home directory:
	// 1. Directory of the executable
	// 2. Current working directory
	// 3. Fallback to environment variables
	envLoaded := false

	// Try location of the executable
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		if err := godotenv.Load(filepath.Join(exeDir, ".env")); err == nil {
			envLoaded = true
		}
	}

	// If not loaded yet, try current working directory
	if !envLoaded {
		if err := godotenv.Load(); err == nil {
			envLoaded = true
		}
	}

	if !envLoaded {
		logger.Print("No .env file found, using environment variables")
	}

	// --- Database Setup ---
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		logger.Fatal("DATABASE_URL environment variable not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	// Migrate the schema
	db.AutoMigrate(&entity.WalletBalance{})

	// --- Telegram Bot Setup ---
	botToken := os.Getenv("TELEGRAM_APP_BOT_TOKEN")
	if botToken == "" {
		logger.Fatal("TELEGRAM_APP_BOT_TOKEN environment variable not set")
	}
	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	if chatIDStr == "" {
		logger.Fatal("TELEGRAM_CHAT_ID environment variable not set")
	}
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		logger.Fatalf("Invalid TELEGRAM_CHAT_ID: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		logger.Fatalf("Failed to create Telegram bot: %v", err)
	}
	bot.Debug = false
	logger.Printf("Authorized on account %s", bot.Self.UserName)

	// --- Blockchain RPC Setup ---
	ethRPC := os.Getenv("ETH_RPC_URL")
	if ethRPC == "" {
		logger.Println("ETH_RPC_URL not set, using public RPC: https://cloudflare-eth.com")
		ethRPC = "https://cloudflare-eth.com"
	}
	blockchains = []Blockchain{
		{Name: "Ethereum", RpcURL: ethRPC},
		{Name: "Arbitrum", RpcURL: os.Getenv("ARBITRUM_RPC_URL")},
		{Name: "Base", RpcURL: os.Getenv("BASE_RPC_URL")},
		{Name: "BSC", RpcURL: os.Getenv("BSC_RPC_URL")},
	}

	for _, b := range blockchains {
		if b.RpcURL == "" {
			logger.Fatalf("%s_RPC_URL environment variable not set", b.Name)
		}
	}

	// --- Main Logic ---
	logger.Println("Starting wallet balance check...")

	// Track statistics
	stats := &RowCountStats{}

	var wallets []entity.WalletBalance
	// Find wallets that haven't been notified yet
	if err := db.Where("is_notified = ?", false).Find(&wallets).Error; err != nil {
		logger.Fatalf("Failed to fetch wallets: %v", err)
	}

	stats.TotalWallets = len(wallets)
	logger.Printf("Found %d wallets to check.", len(wallets))

	for i := range wallets {
		processWallet(db, bot, chatID, stats, &wallets[i])
	}

	logger.Println("Wallet balance check finished.")

	// Send summary notification in prod mode
	if *prodMode {
		sendSummaryNotification(bot, chatID, stats)
	}
}

// sendSummaryNotification sends a summary of the processed wallets to Telegram
func sendSummaryNotification(bot *tgbotapi.BotAPI, chatID int64, stats *RowCountStats) {
	messageText := fmt.Sprintf("✅ Checker command completed!\n"+
		"Rows processed/updated: %d",
		stats.TotalWallets,
	)

	msg := tgbotapi.NewMessage(chatID, messageText)

	_, err := bot.Send(msg)
	if err != nil {
		logger.Printf("Failed to send summary notification: %v", err)
	}
}

// processWallet derives address, checks balances, and sends notification if needed
func processWallet(db *gorm.DB, bot *tgbotapi.BotAPI, chatID int64, stats *RowCountStats, wallet *entity.WalletBalance) {
	// 1. Derive address if it's not already set
	if wallet.Address == "" {
		address, err := deriveAddress(wallet.Mnemonic)
		if err != nil {
			logger.Printf("Failed to derive address for mnemonic ID %d: %v", wallet.ID, err)
			stats.Errors++
			return
		}
		wallet.Address = address
		if err := db.Save(wallet).Error; err != nil {
			logger.Printf("Failed to save address for wallet %d: %v", wallet.ID, err)
			stats.Errors++
			return // Continue to next wallet if save fails
		}
		stats.AddressDerived++
		logger.Printf("Derived address %s for wallet ID %d", wallet.Address, wallet.ID)
	}

	// 2. Check balance on each blockchain and update timestamp
	now := time.Now()
	wallet.BalanceUpdatedAt = &now
	if err := db.Save(wallet).Error; err != nil {
		logger.Printf("Failed to update balance timestamp for wallet %d: %v", wallet.ID, err)
		stats.Errors++
		// We can still proceed with balance checking even if the timestamp update fails
	} else {
		stats.BalanceUpdated++
	}

	for _, chain := range blockchains {
		balance, err := checkBalance(chain.RpcURL, wallet.Address)
		if err != nil {
			logger.Printf("Error checking balance on %s for %s: %v", chain.Name, wallet.Address, err)
			continue // Try next blockchain
		}

		// If balance is found
		if balance.Cmp(big.NewInt(0)) > 0 {
			logger.Printf("FOUND BALANCE on %s for address %s: %s", chain.Name, wallet.Address, balance.String())

			// Prepare notification
			debankURL := fmt.Sprintf("https://debank.com/profile/%s", wallet.Address)
			messageText := fmt.Sprintf("💰 Found a wallet with a balance!\n\nChain: %s\nAddress: %s\n\nView on DeBank:\n%s", chain.Name, wallet.Address, debankURL)
			msg := tgbotapi.NewMessage(chatID, messageText)

			// Send Telegram notification
			_, err := bot.Send(msg)
			if err != nil {
				logger.Printf("Failed to send Telegram notification for wallet %s: %v", wallet.Address, err)
				// If sending fails, we don't update the DB and will retry on the next run.
				return
			}

			stats.Notifications++
			logger.Printf("Successfully sent Telegram notification for wallet %s", wallet.Address)

			// Update wallet in DB only after successful notification
			wallet.Balance = balance.String()
			wallet.IsNotified = true
			if err := db.Save(wallet).Error; err != nil {
				logger.Printf("Failed to update wallet %d after sending notification: %v", wallet.ID, err)
				// Even if this fails, we don't want to re-notify. The log will have to suffice.
			}

			// Once balance is found and notified, we are done with this wallet
			return
		}
		logger.Printf("Zero balance on %s for %s", chain.Name, wallet.Address)
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
