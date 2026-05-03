import os
import re
import math
import requests
import logging
from celery import Celery
from pdf2image import convert_from_path
import numpy as np

# Configurar Celery
redis_url = os.environ.get("CELERY_BROKER_URL", "redis://redis:6379/0")
app = Celery("ocr_tasks", broker=redis_url)

app.conf.update(
    task_serializer="json",
    accept_content=["json"],
    result_serializer="json",
    timezone="UTC",
    enable_utc=True,
    worker_prefetch_multiplier=1 # Para distribuir equitativamente si hay PDFs más pesados
)

# Configurar Logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Instanciar PaddleOCR de forma global en el worker para no cargarlo en cada tarea
from paddleocr import PaddleOCR
ocr_engine = PaddleOCR(use_angle_cls=True, lang='es', use_gpu=False)

# URL de la API de Go
API_BASE_URL = os.environ.get("API_BASE_URL", "http://api:8080/api/v1")

# Caché local para el mapeo Codigo de Mesa -> IDMesa (MongoDB ObjectID)
MESA_MAP_CACHE = {}

def find_value_near(ocr_results, anchor_text, max_dist=500):
    """
    Busca el número más cercano a un texto ancla.
    """
    anchor_box = None
    # Limpieza más agresiva para el ancla
    anchor_clean = re.sub(r'[^A-Z]', '', anchor_text.upper())
    
    for box, (text, score) in ocr_results:
        text_clean = re.sub(r'[^A-Z]', '', text.upper())
        if anchor_clean in text_clean or text_clean in anchor_clean:
            if len(text_clean) > 3 or text_clean == anchor_clean:
                anchor_box = box
                break
            
    if not anchor_box:
        return None
        
    ax_center = (anchor_box[0][0] + anchor_box[1][0]) / 2
    ay_center = (anchor_box[0][1] + anchor_box[2][1]) / 2
    ax_right = anchor_box[1][0]
    ay_bottom = anchor_box[2][1]
    
    best_match = None
    min_dist = float('inf')
    
    for box, (text, score) in ocr_results:
        num_str = re.sub(r'\D', '', text)
        if not num_str: continue
            
        nx_center = (box[0][0] + box[1][0]) / 2
        ny_center = (box[0][1] + box[2][1]) / 2
        
        # Alineación razonable para capturar valores a la derecha
        is_to_right = (box[0][0] >= ax_right - 10) and (abs(ny_center - ay_center) < 30)
        # Limitar búsqueda vertical a la vecindad inmediata (máx 150px abajo)
        is_below = (box[0][1] >= ay_bottom - 10) and (abs(nx_center - ax_center) < 300) and (box[0][1] - ay_bottom < 150)
        
        if is_to_right or is_below:
            dist = math.sqrt((nx_center - ax_center)**2 + (ny_center - ay_center)**2)
            # Penalización para priorizar la lectura horizontal (misma fila)
            if is_below: dist += 100
            
            if dist < min_dist and dist < max_dist:
                min_dist = dist
                best_match = int(num_str)
                    
    return best_match

def find_text_near(ocr_results, anchor_text, max_dist=300):
    """
    Busca texto a la derecha con alineación horizontal flexible.
    """
    anchor_box = None
    anchor_clean = re.sub(r'[^A-Z]', '', anchor_text.upper())
    
    for box, (text, score) in ocr_results:
        text_clean = re.sub(r'[^A-Z]', '', text.upper())
        if anchor_clean in text_clean:
            anchor_box = box
            break
    
    if not anchor_box: return ""
        
    ax_right = anchor_box[1][0]
    ay_center = (anchor_box[0][1] + anchor_box[2][1]) / 2
    
    best_text = ""
    min_dist = float('inf')
    
    for box, (text, score) in ocr_results:
        if re.sub(r'[^A-Z]', '', text.upper()) == anchor_clean: continue
        
        nx_left = box[0][0]
        ny_center = (box[0][1] + box[2][1]) / 2
        
        # Margen de 25px para compensar rotaciones leves del escaneo
        if nx_left >= ax_right - 10 and abs(ny_center - ay_center) < 25:
            dist = nx_left - ax_right
            if dist < min_dist:
                min_dist = dist
                best_text = text
                
    return best_text.strip(": ").strip()

@app.task(name="ocr.process_pdf")
def process_pdf(pdf_path):
    logger.info(f"Iniciando procesamiento de: {pdf_path}")
    
    filename = os.path.basename(pdf_path)
    match = re.search(r"acta_(\d+)\.pdf", filename)
    if not match:
        logger.error(f"Nombre de archivo inválido: {filename}")
        return {"status": "error", "message": "Nombre de archivo inválido"}
        
    codigo_mesa = match.group(1)
    
    # 2. Convertir PDF a Imagen
    try:
        images = convert_from_path(pdf_path, first_page=1, last_page=1)
        if not images:
            logger.error(f"El PDF {filename} no tiene imágenes.")
            return {"status": "error", "message": "PDF vacío"}
        img = np.array(images[0])
        img = img[:, :, ::-1].copy()
    except Exception as e:
        logger.error(f"Error al convertir PDF ({filename}): {e}")
        return {"status": "error", "message": str(e)}

    # 3. Realizar OCR
    try:
        result = ocr_engine.ocr(img, cls=True)
        ocr_lines = result[0]
        if not ocr_lines:
            logger.warning(f"No se detectó texto en {filename}")
            return {"status": "error", "message": "No se detectó texto"}
        
        detected_texts = [line[1][0] for line in ocr_lines]
        logger.info(f"OCR en {filename}: {detected_texts}")
    except Exception as e:
        logger.error(f"Error durante OCR ({filename}): {e}")
        return {"status": "error", "message": str(e)}

    # 4. Extraer los campos usando referencias espaciales
    # Candidatos
    p1 = find_value_near(ocr_lines, "Daenerys Targaryen")
    p2 = find_value_near(ocr_lines, "Sansa Stark")
    p3 = find_value_near(ocr_lines, "Robert Baratheon")
    p4 = find_value_near(ocr_lines, "Tyrion Lannister")
    
    # Totales de votos
    votos_validos = find_value_near(ocr_lines, "VOTOS VALIDOS")
    votos_blancos = find_value_near(ocr_lines, "VOTOS BLANCOS")
    votos_nulos = find_value_near(ocr_lines, "VOTOS NULOS")
    
    # Metadatos de ubicación
    departamento = find_text_near(ocr_lines, "Departamento")
    provincia = find_text_near(ocr_lines, "Provincia")
    municipio = find_text_near(ocr_lines, "Municipio")
    recinto = find_text_near(ocr_lines, "Recinto")
    
    # Datos adicionales de la mesa (Anclas completas de la imagen)
    habilitados = find_value_near(ocr_lines, "ELECTORES HABILITADOS EN LA MESA")
    papeletas_anfora = find_value_near(ocr_lines, "CANTIDAD TOTAL DE PAPELETAS EN ANFORA")
    papeletas_no_usadas = find_value_near(ocr_lines, "CANTIDAD TOTAL DE PAPELETAS NO UTILIZADAS")
    
    # Tiempos
    ap_hora = find_value_near(ocr_lines, "A horas")
    ap_min = find_value_near(ocr_lines, "Minutos")
    cierre_hora = find_value_near(ocr_lines, "concluy a horas")
    cierre_min = 0 # Valor por defecto
    
    # Priorizar el número de mesa extraído del código del acta
    try:
        mesa_num = int(codigo_mesa[-3:])
    except:
        pass

    # 5. Comprobar si hay campos críticos ilegibles
    campos_criticos = {
        "p1": p1, "p2": p2, "p3": p3, "p4": p4,
        "votosNulos": votos_nulos, "votosBlanco": votos_blancos, "votosValidos": votos_validos
    }
    
    campos_ilegibles = [k for k, v in campos_criticos.items() if v is None]
    
    if campos_ilegibles:
        log_dir = "/app/data/logs"
        os.makedirs(log_dir, exist_ok=True)
        log_file = os.path.join(log_dir, "unreadable_records.log")
        mensaje = f"Mesa: {codigo_mesa} | Archivo: {filename} | Ilegibles: {', '.join(campos_ilegibles)}\n"
        with open(log_file, "a", encoding="utf-8") as f:
            f.write(mensaje)
        return {"status": "error", "message": f"Campos ilegibles: {', '.join(campos_ilegibles)}"}

    # 6. Armar el Payload con los datos extraídos
    payload = {
        "papeletasNoUsadas": papeletas_no_usadas or 0,
        "p1": p1,
        "p2": p2,
        "p3": p3,
        "p4": p4,
        "votosNulos": votos_nulos,
        "votosBlanco": votos_blancos,
        "votosValidos": votos_validos,
        "codigoMesa": codigo_mesa,
        "mesa": mesa_num,
        "departamento": departamento,
        "provincia": provincia,
        "municipio": municipio,
        "recinto": recinto,
        "votantesHabilitados": habilitados or 0,
        "papeletasAnfora": papeletas_anfora or 0,
        "aperturaHora": ap_hora or 0,
        "aperturaMinutos": ap_min or 0,
        "cierreHora": cierre_hora or 0,
        "cierreMinutos": cierre_min or 0,
        "tipoCliente": "OCR"
    }
    
    logger.info(f"Datos extraídos para {codigo_mesa}: {payload}")

    # 6. Enviar a la API de Go
    try:
        # En el proyecto Go existe POST /api/v1/actas
        response = requests.post(f"{API_BASE_URL}/actas", json=payload)
        response.raise_for_status()
        logger.info(f"Acta guardada correctamente para la mesa {codigo_mesa}")
        return {"status": "success", "codigo": codigo_mesa}
    except requests.exceptions.RequestException as e:
        logger.error(f"Error al enviar datos a la API de Go para la mesa {codigo_mesa}: {e}")
        # Si la API devuelve el error en JSON, intentamos leerlo
        if e.response is not None:
             logger.error(f"Respuesta API: {e.response.text}")
        return {"status": "error", "message": f"API error: {str(e)}"}
