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
-   `TELEGRAM_TOPIC_ID`: (Optional) The ID of a specific topic in a forum/group. Leave empty to send messages to the general chat.
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

## Command Line Flags

The application supports the following command line flags:

| Flag | Type | Description |
|------|------|-------------|
| `--prod` | boolean | Run in production mode. When set, console output is suppressed and a summary notification is sent to Telegram after completion. |

### Examples

**Development mode (default):**
```sh
./wallet-balance-checker
```
This will show all log messages in the console.

**Production mode:**
```sh
./wallet-balance-checker --prod
```
This will run silently without console output and send a summary message to Telegram when complete.

## Cron Job Setup

To run the wallet balance checker automatically on a schedule, you can set up a cron job.

### Crontab Entry

Edit your crontab file:
```sh
crontab -e
```

Add the following entry to run the checker every 5 minutes:
```cron
*/5 * * * * /path/to/lucky-eth/wallet-balance-checker --prod >> /var/log/wallet-balance-checker.log 2>&1
```

Or to suppress all output:
```cron
*/5 * * * * /path/to/lucky-eth/wallet-balance-checker --prod > /dev/null 2>&1
```

### How It Works

-   **Lock Mechanism**: The application uses a lock file (`/tmp/wallet-balance-checker.lock`) to prevent concurrent execution. If the previous instance is still running, the new instance will exit silently in production mode.
-   **Production Mode**: When using `--prod` flag:
    -   No console output is shown
    -   A Telegram summary message is sent after completion with:
        -   Number of addresses derived
        -   Number of balances updated
        -   Number of notifications sent
        -   Total rows processed

### Log Rotation

To prevent log files from growing too large, consider setting up log rotation. Add to `/etc/logrotate.d/wallet-balance-checker`:

```
/var/log/wallet-balance-checker.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 root root
}
```