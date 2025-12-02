# Guidelines for AI Coding Agents

## Restricted files
Files in the list contain sensitive data, they MUST NOT be read
- .env

## Project Overview

This project is a small Go application that get wallet address by 12 words (mnemonic) in Ethereum blockchain and stores it in a PostgreSQL database. Then check balance on this wallet and send notification in Telegram, if balance > 0. Project uses GORM for database interactions and supports configuration via environment variables.
