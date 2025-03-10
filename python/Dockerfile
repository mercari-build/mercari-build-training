FROM alpine

RUN apk add --no-cache python3 py3-pip

RUN addgroup -S mercari && adduser -S trainee -G mercari
# RUN chown -R trainee:mercari /path/to/db

WORKDIR /app

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY main.py .
COPY db/items.sql ./db/items.sql

CMD ["python3", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "9000"]
