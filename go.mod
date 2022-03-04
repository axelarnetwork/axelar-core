module github.com/axelarnetwork/axelar-core

go 1.17

require (
	github.com/armon/go-metrics v0.3.10
	github.com/axelarnetwork/tm-events v0.0.0-20220221000027-80a0d7f2077e
	github.com/axelarnetwork/utils v0.0.0-20220203232147-bf4a42f338e8
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/cosmos/cosmos-sdk v0.44.5
	github.com/cosmos/ibc-go v1.2.5
	github.com/ethereum/go-ethereum v1.10.14
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/matryer/moq v0.2.5
	github.com/miguelmota/go-ethereum-hdwallet v0.1.1
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.7
	github.com/regen-network/cosmos-proto v0.3.1
	github.com/rs/zerolog v1.26.1
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.10.1
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.15
	github.com/tendermint/tm-db v0.6.6
	golang.org/x/crypto v0.0.0-20211215165025-cf75a172585e
	golang.org/x/mod v0.5.1
	golang.org/x/text v0.3.7
	google.golang.org/genproto v0.0.0-20211223182754-3ac035c7e7cb
	google.golang.org/grpc v1.43.0
	google.golang.org/protobuf v1.27.1
)

require (
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/opencontainers/image-spec v1.0.1 => github.com/opencontainers/image-spec v1.0.2

replace github.com/opencontainers/runc v1.0.2 => github.com/opencontainers/runc v1.0.3

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // https://github.com/axelarnetwork/axelar-core/issues/36

// Fix upstream GHSA-h395-qcrw-5vmq vulnerability.
// TODO Remove it: https://github.com/cosmos/cosmos-sdk/issues/10409
replace github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.7.0

replace github.com/cosmos/cosmos-sdk v0.44.5 => github.com/axelarnetwork/cosmos-sdk v0.44.5-rosetta
