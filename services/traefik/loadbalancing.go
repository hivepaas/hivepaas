package traefik

type LBStrategy string

const (
	LBStrategyWeightedRoundRobin  LBStrategy = "wrr"
	LBStrategyPowerOfTwoChoices   LBStrategy = "p2c"
	LBStrategyHighestRandomWeight LBStrategy = "hrw"
	LBStrategyLeastTime           LBStrategy = "leasttime"
)

var (
	AllLBStrategies = []LBStrategy{LBStrategyWeightedRoundRobin, LBStrategyPowerOfTwoChoices,
		LBStrategyHighestRandomWeight, LBStrategyLeastTime}
)
