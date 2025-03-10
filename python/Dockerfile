FROM python:3.9-alpine
RUN apk add --no-cache python3 py3-pip

WORKDIR /app

# Install dependencies (layer 1)
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

# Copy rest of code (layer 2)
COPY main.py .
COPY db/items.sql ./db/items.sql

# Run backend
CMD ["python3", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "9000"]
