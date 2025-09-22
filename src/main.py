from fastapi import FastAPI
from prometheus_client import generate_latest, Counter, Histogram
import uvicorn

app = FastAPI(title="Cloud Resource Optimizer", version="1.0.0")

# Prometheus metrics
scan_requests = Counter('scan_requests_total', 'Total scan requests')
scan_duration = Histogram('scan_duration_seconds', 'Scan duration')

@app.get("/")
async def root():
    return {"message": "Cloud Resource Optimizer API", "version": "1.0.0"}

@app.get("/health")
async def health():
    return {"status": "healthy"}

@app.get("/metrics")
async def metrics():
    return generate_latest()

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
