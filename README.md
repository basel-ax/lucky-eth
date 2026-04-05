# lucky-eth

`lucky-eth` is a Go-based project for working with Ethereum wallets. It includes tools for generating wallets, checking balances, and receiving notifications.

## LuckySix
This is a part of the project. All results you can get by 3 projects.    
 - [LuckySix](https://github.com/basel-ax/luckysix)
 - [LuckyEth](https://github.com/basel-ax/lucky-eth)
 - [LuckyCosmos](https://github.com/basel-ax/lucky-cosmos)

## Features

-   Generate wallets from 12-word mnemonics.
-   Check wallet balances across multiple Ethereum and L2 blockchains.
-   Receive Telegram notifications for wallets with non-zero balances.

## Configuration

The project is configured using environment variables. An `example.env` file is provided in the project root. To get started, copy it to a new file named `.env`:

```sh
cp example.env .env
```

Then, open `.env` and fill in the required values.

### Required Environment Variables

-   `DATABASE_URL`: The connection string for your PostgreSQL database.
-   `TELEGRAM_APP_BOT_TOKEN`: The token for your Telegram bot.
-   `TELEGRAM_CHAT_ID`: The ID of the Telegram chat where notifications will be sent.
-   `ETH_RPC_URL`: The RPC endpoint for the Ethereum mainnet.
-   `ARBITRUM_RPC_URL`: The RPC endpoint for the Arbitrum network.
-   `BASE_RPC_URL`: The RPC endpoint for the Base network.
-   `BSC_RPC_URL`: The RPC endpoint for the Binance Smart Chain.

## Commands

This project includes the following commands, which can be found in the `cmd` directory.

### `wallet-balance-checker`

This command checks the balances of Ethereum wallets stored in the database.

#### How to Run

1.  **Navigate to the project root directory.**
2.  **Build the command:**
    ```sh
    go build ./cmd/wallet-balance-checker
    ```
3.  **Run the command:**
    ```sh
    ./wallet-balance-checker
    ```

For more details, see the `README.md` file in the `cmd/wallet-balance-checker` directory.

## Cron Job Setup

The `wallet-balance-checker` command can be scheduled to run automatically using cron.

### Crontab Entry

Edit your crontab file:
```sh
crontab -e
```

Add the following entry to run the checker every 5 minutes in production mode:
```cron
*/5 * * * * /path/to/lucky-eth/wallet-balance-checker --prod >> /var/log/wallet-balance-checker.log 2>&1
```

Or to suppress all output:
```cron
*/5 * * * * /path/to/lucky-eth/wallet-balance-checker --prod > /dev/null 2>&1
```

### Run Every 35 Minutes

To run the checker every 35 minutes instead:
```cron
*/35 * * * * /path/to/lucky-eth/wallet-balance-checker --prod >> /var/log/wallet-balance-checker.log 2>&1
```

Or to suppress all output:
```cron
*/35 * * * * /path/to/lucky-eth/wallet-balance-checker --prod > /dev/null 2>&1
```

### How It Works

-   **Lock Mechanism**: The application uses a lock file (`/tmp/wallet-balance-checker.lock`) to prevent concurrent execution. If the previous instance is still running, the new instance will exit silently in production mode.
-   **Production Mode**: When using `--prod` flag:
    -   No console output is shown
    -   A Telegram summary message is sent after completion with:
        -   Number of addresses derived
        -   Number of balances updated
        -   Number of notifications sent
