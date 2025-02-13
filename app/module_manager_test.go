package app_test

import (
	"testing"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/golang/mock/gomock"

	"github.com/axelarnetwork/axelar-core/app"
)

func TestFilteredModuleManager_RegisterServices(t *testing.T) {
	encodingConfig := app.MakeEncodingConfig()
	configurator := module.NewConfigurator(encodingConfig.Codec, bam.NewMsgServiceRouter(), bam.NewGRPCQueryRouter())
	mockCtrl := gomock.NewController(t)

	mockAppModule1 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)
	mockAppModule2 := mock.NewMockAppModuleWithAllExtensions(mockCtrl)

	mockAppModule1.EXPECT().Name().Times(3).Return("module1")
	mockAppModule2.EXPECT().Name().Times(3).Return("module2")

	mockAppModule1.EXPECT().RegisterServices(configurator).Times(1)
	mockAppModule2.EXPECT().RegisterServices(configurator).Times(0)

	mm := app.NewFilteredModuleManager([]module.AppModule{mockAppModule1, mockAppModule2}, []string{"module2"})
	mm.RegisterServices(configurator)

}
