# Wallet Balance Checker

This command is a standalone application that checks the balances of Ethereum wallets stored in a database. It's designed to monitor wallets generated from 12-word mnemonics and notify a Telegram chat when a wallet with a non-zero balance is found.

## Features

-   Derives Ethereum addresses from 12-word mnemonics.
-   Checks wallet balances on multiple blockchains:
    -   Ethereum Mainnet
    -   Arbitrum
    -   Base
    -   Binance Smart Chain (BSC)
-   Sends a notification to a specified Telegram chat via a bot if a balance greater than zero is found on any of the supported networks.
-   Updates the database to mark wallets as notified to prevent duplicate alerts.

## Configuration

The application is configured using environment variables. You can create a `.env` file in the root of the `lucky-eth` project or set them in your environment.

### Required Environment Variables

-   `DATABASE_URL`: The connection string for your PostgreSQL database.
    -   Example: `postgres://user:password@localhost:5432/database_name`
-   `TELEGRAM_APP_BOT_TOKEN`: The token for your Telegram bot, obtained from BotFather.
-   `TELEGRAM_CHAT_ID`: The unique identifier for the target chat where notifications will be sent.
-   `ETH_RPC_URL`: The HTTP RPC endpoint for an Ethereum mainnet node (e.g., from Infura, Alchemy, or your own node).
-   `ARBITRUM_RPC_URL`: The HTTP RPC endpoint for an Arbitrum One node.
-   `BASE_RPC_URL`: The HTTP RPC endpoint for a Base node.
-   `BSC_RPC_URL`: The HTTP RPC endpoint for a Binance Smart Chain node.

An example `example.env` file is provided in the project root. You can copy it to `.env` and fill in your values.

## Database

The command requires a PostgreSQL database with a `wallet_balances` table. The application will automatically attempt to migrate the database schema for the `entity.WalletBalance` struct upon startup, creating the table if it doesn't exist.

The `WalletBalance` entity is expected to have at least a `mnemonic` column populated. The application will then derive the `address`, check the `balance`, and update the record accordingly.

## How to Run

1.  **Navigate to the project root directory:**
    ```sh
    cd /path/to/lucky-eth
    ```

2.  **Ensure all dependencies are downloaded:**
    ```sh
    go mod tidy
    ```

3.  **Build the application:**
    ```sh
    go build ./cmd/wallet-balance-checker
    ```

4.  **Run the compiled binary:**
    Make sure your `.env` file is in the same directory or that the environment variables are exported.
    ```sh
    ./wallet-balance-checker
    ```

The application will run, check all un-notified wallets in the database, send notifications if applicable, and then exit. You can set this up to run on a schedule using a cron job or a similar task scheduler.