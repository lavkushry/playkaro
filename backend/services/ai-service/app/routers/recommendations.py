from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import List
import pandas as pd
import random

router = APIRouter()

class RecommendationResponse(BaseModel):
    user_id: str
    recommended_games: List[str]
    reason: str

# Mock Data for Collaborative Filtering (User-Item Matrix)
# In production, this would come from a database/warehouse
mock_play_history = {
    "user_1": ["ludo_classic", "crash_aviator"],
    "user_2": ["ludo_classic", "carrom_pro"],
    "user_3": ["crash_aviator", "dice_classic"],
    "user_4": ["ludo_classic", "carrom_pro", "chess_master"],
}

game_metadata = {
    "ludo_classic": {"category": "SKILL"},
    "carrom_pro": {"category": "SKILL"},
    "crash_aviator": {"category": "CASINO"},
    "dice_classic": {"category": "CASINO"},
    "chess_master": {"category": "SKILL"},
}

@router.get("/recommendations/{user_id}", response_model=RecommendationResponse)
def get_recommendations(user_id: str):
    # 1. Get user history
    history = mock_play_history.get(user_id, [])

    if not history:
        # Cold start: Recommend popular games
        return RecommendationResponse(
            user_id=user_id,
            recommended_games=["ludo_classic", "crash_aviator"],
            reason="Popular Games"
        )

    # 2. Simple Content-Based Filtering
    # Find category user plays most
    categories = [game_metadata.get(g, {}).get("category") for g in history]
    favorite_category = max(set(categories), key=categories.count) if categories else "SKILL"

    # Recommend games in that category they haven't played
    recommendations = []
    for game, meta in game_metadata.items():
        if meta["category"] == favorite_category and game not in history:
            recommendations.append(game)

    # If no recommendations (played all in category), recommend from other categories
    if not recommendations:
        recommendations = [g for g in game_metadata if g not in history]

    return RecommendationResponse(
        user_id=user_id,
        recommended_games=recommendations[:3],
        reason=f"Because you like {favorite_category} games"
    )
