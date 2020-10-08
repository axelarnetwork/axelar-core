module github.com/axelarnetwork/axelar-core

go 1.13

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/cosmos/cosmos-sdk v0.39.1
	github.com/golang/protobuf v1.4.0
	github.com/gorilla/mux v1.7.4
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.33.7
	github.com/tendermint/tm-db v0.5.1
)

replace github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4