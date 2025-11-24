from fastapi import APIRouter
from pydantic import BaseModel
from typing import Dict, Any

router = APIRouter()

class GameSessionData(BaseModel):
    user_id: str
    game_id: str
    metrics: Dict[str, Any]

class AnalysisResponse(BaseModel):
    risk_score: int
    flagged: bool
    reason: str

@router.post("/analyze/session", response_model=AnalysisResponse)
def analyze_session(data: GameSessionData):
    # Heuristic-based Anti-Cheat (Mocking ML Model)

    risk_score = 0
    reasons = []

    # Check 1: Reaction Time (Ludo/Carrom)
    if "avg_reaction_time_ms" in data.metrics:
        rt = data.metrics["avg_reaction_time_ms"]
        if rt < 100: # Superhuman
            risk_score += 80
            reasons.append("Inhuman reaction time (<100ms)")
        elif rt < 200:
            risk_score += 40
            reasons.append("Suspiciously fast reaction")

    # Check 2: Win Rate (if provided)
    if "win_rate" in data.metrics:
        wr = data.metrics["win_rate"]
        if wr > 0.9: # 90% win rate
            risk_score += 50
            reasons.append("Abnormal win rate")

    # Check 3: Click Consistency (Standard Deviation)
    if "click_std_dev" in data.metrics:
        std = data.metrics["click_std_dev"]
        if std < 5: # Robotic precision
            risk_score += 90
            reasons.append("Robotic click consistency")

    flagged = risk_score > 70

    return AnalysisResponse(
        risk_score=min(risk_score, 100),
        flagged=flagged,
        reason="; ".join(reasons) if reasons else "Normal behavior"
    )
