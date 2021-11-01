package types

// NewPool is the constructor of Pool
func NewPool(name string) Pool {
	return Pool{
		Name:    name,
		Rewards: []*Pool_Reward{},
	}
}
