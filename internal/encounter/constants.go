package encounter

const (
	MaxClicksPerWarp       = 21
	MinClicksPerWarp       = 5
	EncounterBaseThreshold = 150
	DifficultyThresholdMod = 10
	AlienArtifactChance    = 4
	AlienArtifactDenom     = 20
	RareEncounterOdds      = 1000

	BountyDivisor  = 200
	BountyRounding = 25
	BountyMin      = 25
	BountyMax      = 2500

	BribeBaseDivisor   = 10
	BribeDiffFactor    = 5
	BribeDiffBase      = 4
	BribeRounding      = 100
	MinBribeCost       = 100
	MaxBribeCost       = 10000
	MinPrisonDays      = 30
	ArrestedRecordReset = -5
	IllegalGoodFine    = 500

	PirateLossMin        = 100
	PirateLossRange      = 400
	PirateSurrenderMin   = 200
	PirateSurrenderRange = 800
	CargoLossChance      = 50
	CargoLossDenom       = 100
)

func ClicksForDistance(distance float64, maxRange int) int {
	if maxRange <= 0 {
		return MinClicksPerWarp
	}
	ratio := distance / float64(maxRange)
	clicks := MinClicksPerWarp + int(ratio*float64(MaxClicksPerWarp-MinClicksPerWarp))
	if clicks < MinClicksPerWarp {
		clicks = MinClicksPerWarp
	}
	if clicks > MaxClicksPerWarp {
		clicks = MaxClicksPerWarp
	}
	return clicks
}
