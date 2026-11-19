package types

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
)

//go:generate moq -pkg mock -out ./mock/expected_keepers.go . UpgradeKeeper

// UpgradeKeeper provides access to the upgrade module's plan scheduling
type UpgradeKeeper interface {
	GetUpgradePlan(ctx context.Context) (plan upgradetypes.Plan, err error)
	ScheduleUpgrade(ctx context.Context, plan upgradetypes.Plan) error
}
