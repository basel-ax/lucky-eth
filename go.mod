module github.com/basel-ax/lucky-eth

go 1.25.3

require (
	github.com/ethereum/go-ethereum v1.10.26
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/joho/godotenv v1.5.1
	github.com/miguelmota/go-ethereum-hdwallet v0.1.1
	gorm.io/driver/postgres v1.5.0
	gorm.io/gorm v1.31.1
)

// The version of go-ethereum-hdwallet we are using depends on an older version
// of the btcsuite library, where the `btcec` package was part of the main `btcd`
// module. Newer versions of `btcd` have moved this package. The `replace`
// directive below pins `btcd` to a version that is compatible with our
// dependencies and prevents `go mod tidy` from trying to upgrade it to a
// version with breaking changes.
replace github.com/btcsuite/btcd => github.com/btcsuite/btcd v0.22.1

require (
	github.com/btcsuite/btcd v0.23.0 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.0 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.1 // indirect
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce // indirect
	github.com/deckarep/golang-set v1.8.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/holiman/uint256 v1.2.2-0.20230321075855-87b91420868c // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.3.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.0 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/crypto v0.9.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	gopkg.in/natefinch/npipe.v2 v2.0.0-20160621034901-c1b8fa8bdcce // indirect
)
