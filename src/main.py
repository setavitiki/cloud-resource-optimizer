from fastapi import FastAPI, HTTPException
from prometheus_client import generate_latest, Counter
from aws_scanner import AWSResourceScanner
import uvicorn
import logging
import os

# Setup logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

app = FastAPI(title="Cloud Resource Optimizer", version="1.0.0")

# Prometheus metrics
scan_requests = Counter('scan_requests_total', 'Total scan requests')

# Initialize AWS scanner
aws_region = os.getenv('AWS_REGION', 'us-east-1')
scanner = AWSResourceScanner(region=aws_region)

@app.get("/")
async def root():
    return {
        "message": "Cloud Resource Optimizer API", 
        "version": "1.0.0",
        "region": aws_region
    }

@app.get("/health")
async def health():
    return {"status": "healthy", "timestamp": "2025-09-22T18:46:00Z"}

@app.get("/scan")
async def scan_resources():
    """Perform full AWS resource scan"""
    try:
        scan_requests.inc()
        results = scanner.full_scan()
        return results
    except Exception as e:
        logger.error(f"Scan failed: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Scan failed: {str(e)}")

@app.get("/scan/volumes")
async def scan_orphaned_volumes():
    """Scan for orphaned EBS volumes"""
    try:
        results = scanner.scan_orphaned_volumes()
        return {"orphaned_volumes": results}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Volume scan failed: {str(e)}")

@app.get("/scan/instances")
async def scan_idle_instances():
    """Scan for idle EC2 instances"""
    try:
        results = scanner.scan_idle_instances()
        return {"idle_instances": results}
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Instance scan failed: {str(e)}")

@app.get("/metrics")
async def metrics():
    return generate_latest()

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
