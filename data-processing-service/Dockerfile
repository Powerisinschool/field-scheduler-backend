FROM python:3.11-slim

# Install system dependencies for OpenCV, Tesseract, and pdf2image
RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    poppler-utils \
    libgl1 \
    libglib2.0-0 \
    gdal-bin \
    libgdal-dev \
    g++ \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .
# Upgrade pip first, then install requirements
RUN pip install --upgrade pip && \
    pip install --no-cache-dir -r requirements.txt

COPY . .

EXPOSE 8000

# Render assigns a PORT environment variable dynamically. 
# We set a default of 8000 so local testing still works.
ENV PORT=8000

# Execute uvicorn using the environment variable
CMD uvicorn main:app --host 0.0.0.0 --port $PORT
