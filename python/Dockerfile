FROM python:3-alpine

# Install dependencies and remove unnecessary files
RUN apk add --no-cache --virtual .build-deps \
    python3-dev \
    libffi-dev \
    gcc \
    musl-dev \
    && apk add --no-cache py3-pip \
    && pip install --no-cache-dir --upgrade pip \
    && apk del .build-deps  # Remove build dependencies

RUN addgroup -S mercari && adduser -S trainee -G mercari
# RUN chown -R trainee:mercari /path/to/db

WORKDIR /app

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
RUN pip install --no-cache-dir -r requirements.txt \
    && rm -rf /root/.cache/pip  # Remove pip cache

COPY main.py ./
COPY db/items.sql ./db/items.sql

# Ensure correct ownership in a single layer
RUN chown -R trainee:mercari /app

USER trainee

EXPOSE 9000

CMD ["python3", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "9000"]
