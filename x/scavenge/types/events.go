package types

// scavenge module event types
const (
	EventTypeCreateScavenge = "CreateScavenge"
	EventTypeCommitSolution = "CommitSolution"
	EventTypeSolveScavenge  = "SolveScavenge"
	EventTypeTransferTokens = "TransferTokens"

	AttributeDescription           = "description"
	AttributeSolution              = "solution"
	AttributeSolutionHash          = "solutionHash"
	AttributeReward                = "reward"
	AttributeScavenger             = "scavenger"
	AttributeSolutionScavengerHash = "solutionScavengerHash"
	AttributeRecipient             = "recipient"
	AttributeAmount                = "amount"

	AttributeValueCategory = ModuleName
)
