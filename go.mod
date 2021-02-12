module github.com/axelarnetwork/axelar-core

go 1.13

require (
	github.com/axelarnetwork/tssd v0.0.0-20210115070842-5aaffe0b2178
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/cosmos/cosmos-sdk v0.39.1
	github.com/ethereum/go-ethereum v1.9.25
	github.com/gorilla/mux v1.7.4
	github.com/matryer/moq v0.2.0
	github.com/miguelmota/go-ethereum-hdwallet v0.0.0-20200123000308-a60dcd172b4c
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.33.7
	github.com/tendermint/tm-db v0.5.1
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	google.golang.org/grpc v1.32.0
)

replace github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // https://github.com/axelarnetwork/axelar-core/issues/36
