package model

// RewardRecommendation holds Prolific's recommended reward-per-hour rates
// for a given workspace, currency, and (optionally) a set of screener
// filter IDs. MinRewardPerHour and RecommendedRewardPerHour are both in the
// hundredth subunit of Currency (e.g. cents for USD, pence for GBP).
type RewardRecommendation struct {
	Currency                 string `json:"currency"`
	MinRewardPerHour         int    `json:"min_reward_per_hour"`
	RecommendedRewardPerHour int    `json:"recommended_reward_per_hour"`
}
