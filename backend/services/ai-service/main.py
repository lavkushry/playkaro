from fastapi import FastAPI
from app.routers import recommendations, anticheat

app = FastAPI(title="PlayKaro AI Service", version="1.0.0")

@app.get("/health")
def health_check():
    return {"status": "ok", "service": "ai-service"}

app.include_router(recommendations.router, prefix="/v1/ai", tags=["recommendations"])
app.include_router(anticheat.router, prefix="/v1/ai", tags=["anticheat"])

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8084)
