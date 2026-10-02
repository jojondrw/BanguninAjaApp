FROM python:3.13-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY banguninaja_scoring ./banguninaja_scoring
COPY model ./model

ENV BUNDLE_DIR=/bundle
EXPOSE 8090

CMD ["uvicorn", "banguninaja_scoring.app:app", "--host", "0.0.0.0", "--port", "8090"]
