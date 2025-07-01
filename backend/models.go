package main

import "time"

type Score struct {
	PlayerName      string `json:"player_name"`
	LegacyScore     *int   `json:"legacy_score,omitempty"`
	BaseCards       *int   `json:"base_cards,omitempty"`
	ExtraVP         *int   `json:"extra_vp,omitempty"`
	BasicEvents     *int   `json:"basic_events,omitempty"`
	SpecialEvents   *int   `json:"special_events,omitempty"`
	ProsperityCards *int   `json:"prosperity_cards,omitempty"`
	Visitors        *int   `json:"visitors,omitempty"`
	Journey         *int   `json:"journey,omitempty"`
	GarlandAward    *int   `json:"garland_award,omitempty"`
}

type Game struct {
	ID       int       `json:"id"`
	GameDate time.Time `json:"game_date"`
	Scores   []Score   `json:"scores"`
}