# lucky-eth

`lucky-eth` is a Go-based project for working with Ethereum wallets. It includes tools for generating wallets, checking balances, and receiving notifications.

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