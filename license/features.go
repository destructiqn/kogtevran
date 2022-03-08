package license

type Feature uint64

const (
	FeatureAntiKnockback Feature = 0b1
	FeatureKillAura      Feature = 0b10
	FeatureNoFall        Feature = 0b100
	FeatureFlight        Feature = 0b1000
	FeatureLongJump      Feature = 0b10000
	FeatureUnlimitedCPS  Feature = 0b100000
	FeatureTPAura        Feature = 0b1000000
	FeaturePlayerESP     Feature = 0b10000000
)
