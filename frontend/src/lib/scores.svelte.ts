export interface ScoreForPlayer {
    id: string
    player_name: string,
    base_cards: number,
    extra_vp: number,
    basic_events: number,
    special_events: number,
    prosperity_cards: number,
    visitors: number,
    journey: number,
    garland_award: number,
    legacy_score: number,
    total_score?: number
}

export interface GameData {
    game_date: string;
    scores: ScoreForPlayer[];
}

function getRandomPlayerName() {
    const names = ["Joseph", "James", "Niamh"]
    const randomIndex = Math.floor(Math.random() * names.length);
    return names[randomIndex];
}

export function createPlayer(): ScoreForPlayer {
    return {
        id: crypto.randomUUID(),
        player_name: getRandomPlayerName(),
        base_cards: 0,
        extra_vp: 0,
        basic_events: 0,
        special_events: 0,
        prosperity_cards: 0,
        visitors: 0,
        journey: 0,
        garland_award: 0,
        legacy_score: 0,
    }
}

export function deriveTotalScore(score: ScoreForPlayer) {
    if (score.legacy_score) return score.legacy_score
    return (
        score.base_cards +
        score.extra_vp +
        score.basic_events +
        score.special_events +
        score.prosperity_cards +
        score.visitors +
        score.journey +
        score.garland_award
    )
}