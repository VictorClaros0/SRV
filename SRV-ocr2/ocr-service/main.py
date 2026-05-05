import os
import re
import math
from datetime import datetime
import cv2
import numpy as np
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from pdf2image import convert_from_path
from paddleocr import PaddleOCR
import logging

# Configuración de Logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

LOG_DIR = r"c:\Users\hugho\Desktop\SRV\data\logs"
os.makedirs(LOG_DIR, exist_ok=True)

app = FastAPI(title="Extreme Precision OCR Service")

# MOTOR NITRO
ocr_engine = PaddleOCR(
    use_angle_cls=True, 
    lang="es", 
    show_log=False,
    det_limit_side_len=960,
    use_mp=False
)

class OCRRequest(BaseModel):
    pdf_path: str

class OCRResponse(BaseModel):
    status: str
    data: dict = None
    message: str = None

def get_center(box):
    return (box[0][0] + box[1][0]) / 2, (box[0][1] + box[2][1]) / 2

def find_value_near(ocr_results, anchor_keywords, max_dist=500, direction="either", y_range=None, max_digits=None):
    if isinstance(anchor_keywords, str):
        anchor_keywords = [anchor_keywords]
    
    anchor_box = None
    for box, (text, score) in ocr_results:
        text_up = text.upper()
        nx_c, ny_c = get_center(box)
        if y_range and not (y_range[0] <= ny_c <= y_range[1]): continue
            
        if any(kw.upper() in text_up for kw in anchor_keywords):
            anchor_box = box
            break
    
    if not anchor_box: return None
    
    ax_c, ay_c = get_center(anchor_box)
    ax_r, ay_b = anchor_box[1][0], anchor_box[2][1]
    best_match, min_dist = None, float('inf')
    
    for box, (text, score) in ocr_results:
        num_str = re.sub(r'\D', '', text)
        if not num_str: continue
        if max_digits and len(num_str) > max_digits: continue
        
        nx_c, ny_c = get_center(box)
        nx_l, ny_t = box[0][0], box[0][1]
        
        if y_range and not (y_range[0] <= ny_c <= y_range[1]): continue

        is_to_right = (nx_l >= ax_r - 50) and (abs(ny_c - ay_c) < 50)
        is_below = (ny_t >= ay_b - 30) and (abs(nx_c - ax_c) < 300) and (ny_t - ay_b < max_dist)
        
        valid = False
        if direction == "right" and is_to_right: valid = True
        elif direction == "below" and is_below: valid = True
        elif direction == "either" and (is_to_right or is_below): valid = True
        
        if valid:
            dist = math.sqrt((nx_c - ax_c)**2 + (ny_c - ay_c)**2)
            if is_below and not is_to_right: dist += 100
            if dist < min_dist:
                min_dist, best_match = dist, int(num_str)
                    
    return best_match

def find_text_near(ocr_results, anchor_keywords, max_dist=500):
    if isinstance(anchor_keywords, str):
        anchor_keywords = [anchor_keywords]
    
    # Prioridad 1: Misma línea con ":"
    for box, (text, score) in ocr_results:
        text_up = text.upper()
        for kw in anchor_keywords:
            if kw.upper() in text_up and ":" in text:
                parts = text.split(":", 1)
                if len(parts) > 1 and parts[1].strip():
                    val = parts[1].strip()
                    # Si el valor es una de las etiquetas, ignorar
                    if any(k.upper() in val.upper() for k in ["PROVINCIA", "MUNICIPIO", "LOCALIDAD", "RECINTO"]):
                        continue
                    return val
    
    # Prioridad 2: Buscar a la derecha con distancia mínima
    anchor_box = None
    for box, (text, score) in ocr_results:
        if any(kw.upper() == re.sub(r'[^A-Z]', '', text.upper()) for kw in anchor_keywords):
            anchor_box = box
            break
    
    if not anchor_box: return ""
    ax_r, ay_c = anchor_box[1][0], get_center(anchor_box)[1]
    best_text, min_dist = "", float('inf')
    
    for box, (text, score) in ocr_results:
        nx_l, ny_c = box[0][0], get_center(box)[1]
        if nx_l >= ax_r - 30 and abs(ny_c - ay_c) < 35:
            dist = nx_l - ax_r
            if dist < min_dist:
                # Evitar capturar la siguiente etiqueta
                if any(k.upper() in text.upper() for k in ["MUNICIPIO", "LOCALIDAD", "RECINTO"]): continue
                min_dist, best_text = dist, text
                
    return best_text.strip(": ").strip()

@app.post("/api/ocr/process", response_model=OCRResponse)
async def process_pdf_endpoint(request: OCRRequest):
    try:
        images = convert_from_path(request.pdf_path, dpi=200, first_page=1, last_page=1)
        if not images: return {"status": "error", "message": "PDF vacío"}
        img = np.array(images[0])[:, :, ::-1].copy()

        result = ocr_engine.ocr(img, cls=True)
        ocr_lines = result[0]
        if not ocr_lines: return {"status": "error", "message": "No detectado"}

        # Verificar si existe la línea "PAPELETAS NO UTILIZADAS"
        found_unused = any("NO UTILIZADAS" in t.upper() for b, (t, s) in ocr_lines)
        if not found_unused:
            log_filename = f"ocr_error_{datetime.now().strftime('%Y%m%d_%H%M%S')}.log"
            log_path = os.path.join(LOG_DIR, log_filename)
            with open(log_path, "w", encoding="utf-8") as f:
                f.write(f"Timestamp: {datetime.now().isoformat()}\n")
                f.write(f"PDF Path: {request.pdf_path}\n")
                f.write("Error: La papeleta esta rota. No se encontro la linea 'PAPELETAS NO UTILIZADAS'.\n")
            logger.warning(f"Papeleta rota detectada: {request.pdf_path}")
            return {"status": "error", "message": "La papeleta esta rota. No se encontro la linea 'PAPELETAS NO UTILIZADAS'."}

        # --- SECCIONES Y ---
        y_ap = 0
        y_ci = 10000
        for b, (t, s) in ocr_lines:
            if "APERTURA DE MESA" in t.upper(): y_ap = b[0][1]
            if "CIERRE DE VOTACI" in t.upper(): y_ci = b[0][1]

        data = {
            "papeletasNoUsadas": find_value_near(ocr_lines, "NO UTILIZADAS", direction="below", max_dist=200) or 0,
            "p1": find_value_near(ocr_lines, "Daenerys", direction="right") or 0,
            "p2": find_value_near(ocr_lines, "Sansa", direction="right") or 0,
            "p3": find_value_near(ocr_lines, "Robert", direction="right") or 0,
            "p4": find_value_near(ocr_lines, "Tyrion", direction="right") or 0,
            "votosNulos": find_value_near(ocr_lines, "NULOS", direction="right") or 0,
            "votosBlanco": find_value_near(ocr_lines, "BLANCOS", direction="right") or 0,
            "votosValidos": find_value_near(ocr_lines, "VALIDOS", direction="right") or 0,
            
            # MESA: Búsqueda muy cercana hacia abajo
            "mesa": find_value_near(ocr_lines, "NUMERO DE MESA", direction="below", max_dist=150) or 0,
            "departamento": find_text_near(ocr_lines, "Departamento"),
            "provincia": find_text_near(ocr_lines, "Provincia"),
            "municipio": find_text_near(ocr_lines, "Municipio"),
            "recinto": find_text_near(ocr_lines, "Recinto"),
            
            "votantesHabilitados": find_value_near(ocr_lines, "en la mesa", direction="below", max_dist=200),
            "papeletasAnfora": find_value_near(ocr_lines, "EN ÁNFORA", direction="below", max_dist=100),
            
            # Tiempos: Max 2 dígitos para evitar "17 de agosto"
            "aperturaHora": find_value_near(ocr_lines, "Horas", direction="below", max_dist=80, y_range=(y_ap, y_ci), max_digits=2) or 0,
            "aperturaMinutos": find_value_near(ocr_lines, "Minutos", direction="below", max_dist=80, y_range=(y_ap, y_ci), max_digits=2) or 0,
            "cierreHora": find_value_near(ocr_lines, "Horas", direction="below", max_dist=80, y_range=(y_ci, 10000), max_digits=2) or 0,
            "cierreMinutos": find_value_near(ocr_lines, "Minutos", direction="below", max_dist=80, y_range=(y_ci, 10000), max_digits=2) or 0,
            
            "tipoCliente": "OCR"
        }
        return {"status": "success", "data": data}
    except Exception as e:
        return {"status": "error", "message": str(e)}

@app.get("/health")
def health(): return {"status": "ok"}
