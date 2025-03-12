FROM python:3.9-alpine@sha256:2bac43769ace90ebd3ad83e5392295e25dfc58e58543d3ab326c3330b505283d

RUN apk add --no-cache python3 py3-pip && \
    pip install --upgrade pip

# non-root user for better security
RUN addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app

# Dependencies
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

# application code
COPY main.py .
COPY db/items.sql ./db/

RUN chown -R appuser:appgroup /app

USER appuser

# Expose the application port
EXPOSE 9000

# Use array syntax for CMD
CMD ["python3", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "9000"]
