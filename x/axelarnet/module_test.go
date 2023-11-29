package axelarnet_test

//
//func TestGetMigrationHandler(t *testing.T) {
//	var (
//		ctx       sdk.Context
//		appModule axelarnet.AppModule
//		k         keeper.Keeper
//		n         *mock.NexusMock
//		bankK     *mock.BankKeeperMock
//
//		ack       channeltypes.Acknowledgement
//		transfer  types.IBCTransfer
//		message   exported.GeneralMessage
//		transfers []types.IBCTransfer
//	)
//
//	const (
//		packetSeq = 1
//		channelID = "channel-0"
//	)
//
//	givenAnAppModule := Given("given a module", func() {
//		encCfg := appParams.MakeEncodingConfig()
//		subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey(types.StoreKey), sdk.NewKVStoreKey("tAxelarnetKey"), types.ModuleName)
//		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
//
//		channelK := &mock.ChannelKeeperMock{
//			GetNextSequenceSendFunc: func(sdk.Context, string, string) (uint64, bool) {
//				return packetSeq, true
//			},
//		}
//
//		k = keeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey(types.ModuleName), subspace, channelK, &mock.FeegrantKeeperMock{})
//		ibcK := keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{})
//
//		accountK := &mock.AccountKeeperMock{
//			GetModuleAddressFunc: func(string) sdk.AccAddress {
//				return rand.AccAddr()
//			},
//		}
//
//		bankK = &mock.BankKeeperMock{
//			SendCoinsFunc: func(sdk.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error {
//				return nil
//			},
//			SendCoinsFromAccountToModuleFunc: func(sdk.Context, sdk.AccAddress, string, sdk.Coins) error {
//				return nil
//			},
//			BurnCoinsFunc: func(sdk.Context, string, sdk.Coins) error { return nil },
//		}
//
//
//		n = &mock.NexusMock{}
//		appModule = axelarnet.NewAppModule(ibcK, n, bankK, accountK, log.TestingLogger())
//	})
//
//	fungibleTokenPacket := ibctransfertypes.NewFungibleTokenPacketData(rand.Denom(5, 10), "1", rand.AccAddr().String(), rand.AccAddr().String())
//
//	packet := channeltypes.NewPacket(fungibleTokenPacket.GetBytes(), packetSeq, ibctransfertypes.PortID, channelID, ibctransfertypes.PortID, channelID, clienttypes.NewHeight(0, 110), 0)
//
//	whenGetValidAckResult := When("get valid acknowledgement result", func() {
//		ack = channeltypes.NewResultAcknowledgement([]byte{byte(1)})
//	})
//
//	whenGetValidAckError := When("get valid acknowledgement error", func() {
//		ack = channeltypes.NewErrorAcknowledgement(fmt.Errorf("error"))
//	})
//
//	whenPendingTransfersExist := When("pending transfers exist", func() {
//		transfers = slices.Expand(
//			func(_ int) types.IBCTransfer { return testutils.RandomIBCTransfer() },
//			int(rand.I64Between(5, 50)),
//		)
//
//		slices.ForEach(transfers, func(t types.IBCTransfer) { funcs.MustNoErr(k.EnqueueIBCTransfer(ctx, t)) })
//	})
//
//	seqMapsToID := When("packet seq maps to transfer ID", func() {
//		transfer = testutils.RandomIBCTransfer()
//		transfer.ChannelID = channelID
//		funcs.MustNoErr(k.SetSeqIDMapping(ctx, transfer))
//		funcs.MustNoErr(k.EnqueueIBCTransfer(ctx, transfer))
//	})
//
//	seqMapsToMessageID := When("packet seq maps to message ID", func() {
//		message = nexustestutils.RandomMessage(exported.Processing)
//		funcs.MustNoErr(k.SetSeqMessageIDMapping(ctx, ibctransfertypes.PortID, channelID, packetSeq, message.ID))
//
//		n.GetMessageFunc = func(sdk.Context, string) (exported.GeneralMessage, bool) { return message, true }
//		n.IsAssetRegisteredFunc = func(sdk.Context, exported.Chain, string) bool { return true }
//		n.SetMessageFailedFunc = func(ctx sdk.Context, id string) error {
//			if id == message.ID {
//				message.Status = exported.Failed
//			}
//
//			return nil
//		}
//		n.GetChainByNativeAssetFunc = func(sdk.Context, string) (exported.Chain, bool) { return exported.Chain{}, false }
//	})
//
//	whenOnAck := When("on acknowledgement", func() {
//		err := appModule.OnAcknowledgementPacket(ctx, packet, ack.Acknowledgement(), nil)
//		assert.NoError(t, err)
//	})
//
//	whenOnTimeout := When("on timeout", func() {
//		err := appModule.OnTimeoutPacket(ctx, packet, nil)
//		assert.NoError(t, err)
//	})
//
//	shouldNotChangeTransferState := Then("should not change transfers status", func(t *testing.T) {
//		assert.True(t, slices.All(transfers, func(t types.IBCTransfer) bool {
//			return funcs.MustOk(k.GetTransfer(ctx, t.ID)).Status == types.TransferPending
//		}))
//	})
//
//	whenChainIsActivated := When("chain is activated", func() {
//		n.GetChainFunc = func(ctx sdk.Context, chain exported.ChainName) (exported.Chain, bool) { return exported.Chain{}, true }
//		n.IsChainActivatedFunc = func(ctx sdk.Context, chain exported.Chain) bool { return true }
//		n.RateLimitTransferFunc = func(ctx sdk.Context, chain exported.ChainName, asset sdk.Coin, direction exported.TransferDirection) error {
//			return nil
//		}
//	})
//
//	givenAnAppModule.
//		Branch(
//			whenGetValidAckResult.
//				When2(seqMapsToID).
//				When2(whenOnAck).
//				Then("should set transfer to complete", func(t *testing.T) {
//					transfer := funcs.MustOk(k.GetTransfer(ctx, transfer.ID))
//					assert.Equal(t, types.TransferCompleted, transfer.Status)
//				}),
//
//			whenGetValidAckError.
//				When2(whenChainIsActivated).
//				When2(seqMapsToID).
//				When2(whenOnAck).
//				Then("should set transfer to failed", func(t *testing.T) {
//					transfer := funcs.MustOk(k.GetTransfer(ctx, transfer.ID))
//					assert.Equal(t, types.TransferFailed, transfer.Status)
//				}),
//
//			whenPendingTransfersExist.
//				When("get invalid ack", func() {
//					err := appModule.OnAcknowledgementPacket(ctx, packet, rand.BytesBetween(1, 50), nil)
//					assert.Error(t, err)
//				}).
//				Then2(shouldNotChangeTransferState),
//
//			whenGetValidAckResult.
//				When2(whenPendingTransfersExist).
//				When("seq is not mapped to id", func() {}).
//				When2(whenOnAck).
//				Then2(shouldNotChangeTransferState),
//
//			seqMapsToID.
//				When2(whenChainIsActivated).
//				When2(whenOnTimeout).
//				Then("should set transfer to failed", func(t *testing.T) {
//					transfer := funcs.MustOk(k.GetTransfer(ctx, transfer.ID))
//					assert.Equal(t, types.TransferFailed, transfer.Status)
//				}),
//
//			whenPendingTransfersExist.
//				When("seq is not mapped to id", func() {}).
//				When2(whenChainIsActivated).
//				When2(whenOnTimeout).
//				Then2(shouldNotChangeTransferState),
//
//			whenGetValidAckError.
//				When2(whenChainIsActivated).
//				When2(seqMapsToMessageID).
//				When2(whenOnAck).
//				Then("should set message to failed", func(t *testing.T) {
//					assert.Equal(t, exported.Failed, message.Status)
//					assert.Len(t, bankK.SendCoinsFromAccountToModuleCalls(), 1)
//				}),
//		).Run(t)
//}
