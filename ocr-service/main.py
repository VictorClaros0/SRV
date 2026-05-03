import os
import glob
from fastapi import FastAPI, BackgroundTasks
from pydantic import BaseModel
from celery import Celery

app = FastAPI(title="OCR Service API")

redis_url = os.environ.get("CELERY_BROKER_URL", "redis://redis:6379/0")
celery_app = Celery("ocr_tasks", broker=redis_url)

# Directorio donde están los PDFs (montado por Docker)
PDF_DIR = "/app/data/PdfOutput"

class ScanResponse(BaseModel):
    message: str
    enqueued_files: int

@app.post("/api/ocr/scan-all", response_model=ScanResponse)
def scan_all_pdfs():
    """
    Escanea la carpeta de salida de PDFs y encola cada archivo a Celery para su procesamiento.
    """
    if not os.path.exists(PDF_DIR):
        return {"message": f"El directorio {PDF_DIR} no existe.", "enqueued_files": 0}
        
    pdf_files = glob.glob(os.path.join(PDF_DIR, "*.pdf"))
    
    if not pdf_files:
        return {"message": "No se encontraron archivos PDF en el directorio.", "enqueued_files": 0}
        
    # Encolar tareas en Celery
    for pdf_path in pdf_files:
        celery_app.send_task("ocr.process_pdf", args=[pdf_path])
        
    return {
        "message": "Tareas de OCR encoladas exitosamente.",
        "enqueued_files": len(pdf_files)
    }

@app.get("/health")
def health_check():
    return {"status": "ok"}
