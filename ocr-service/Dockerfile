# Usamos una imagen de Python base
FROM python:3.11-slim

# Evitar prompts interactivos durante la instalación
ENV DEBIAN_FRONTEND=noninteractive

# Instalar dependencias del sistema necesarias para OpenCV, Poppler (PDF) y dependencias base
RUN apt-get update && apt-get install -y \
    libgl1 \
    libglib2.0-0 \
    libgomp1 \
    poppler-utils \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copiar requirements primero para aprovechar el caché de Docker
COPY requirements.txt .

# Instalar librerías de Python. 
# Si no hay GPU disponible en el contenedor, PaddlePaddle usará la CPU automáticamente.
RUN pip install --no-cache-dir --timeout=120 -r requirements.txt

# Copiar el código fuente
COPY . .

# Comando por defecto (será sobreescrito en el docker-compose dependiendo del servicio)
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
