package app_test

import (
	"testing"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/tests/mocks"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
)

func TestFilteredModuleManager_RegisterServices(t *testing.T) {
	encodingConfig := app.MakeEncodingConfig()
	configurator := module.NewConfigurator(encodingConfig.Codec, bam.NewMsgServiceRouter(), bam.NewGRPCQueryRouter())
	mockCtrl := gomock.NewController(t)

	mockAppModule1 := mocks.NewMockAppModule(mockCtrl)
	mockAppModule2 := mocks.NewMockAppModule(mockCtrl)

	mockAppModule1.EXPECT().Name().Times(3).Return("module1")
	mockAppModule2.EXPECT().Name().Times(3).Return("module2")

	mockAppModule1.EXPECT().RegisterServices(configurator).Times(1)
	mockAppModule2.EXPECT().RegisterServices(configurator).Times(0)

	mm := app.NewFilteredModuleManager([]module.AppModule{mockAppModule1, mockAppModule2}, []string{"module2"})
	mm.RegisterServices(configurator)

}

func TestFilteredModuleManager_RegisterRoutes(t *testing.T) {
	encodingConfig := app.MakeEncodingConfig()
	mockCtrl := gomock.NewController(t)

	mockAppModule1 := mocks.NewMockAppModule(mockCtrl)
	mockAppModule2 := mocks.NewMockAppModule(mockCtrl)

	mockAppModule1.EXPECT().Name().Times(3).Return("module1")
	mockAppModule2.EXPECT().Name().Times(3).Return("module2")

	mm := app.NewFilteredModuleManager([]module.AppModule{mockAppModule1, mockAppModule2}, []string{"module2"})

	router := bam.NewRouter()
	queryRouter := bam.NewQueryRouter()
	noopHandler := sdk.Handler(func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) { return nil, nil })

	mockAppModule1.EXPECT().QuerierRoute().Times(1)
	mockAppModule1.EXPECT().Route().Times(1).Return(sdk.NewRoute("route1", noopHandler))
	mockAppModule2.EXPECT().Route().Times(0)

	mm.RegisterRoutes(router, queryRouter, encodingConfig.Amino)
	assert.Nil(t, router.Route(sdk.Context{}, "route2"))
}
