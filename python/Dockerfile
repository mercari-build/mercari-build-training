FROM python:3.12-alpine

RUN addgroup -S mercari && adduser -S trainee -G mercari
RUN apk add --no-cache sqlite

WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY main.py .
COPY db/items.sql /app/db/items.sql
COPY db/mercari.sqlite3 /app/db/mercari.sqlite3

# Create images directory and set permissions
RUN mkdir -p /app/images && chown -R trainee:mercari /app/images

# Fix permissions for database directory
RUN chown -R trainee:mercari /app/db

USER trainee
CMD ["/bin/sh", "-c", "sqlite3 /app/db/mercari.sqlite3 < /app/db/items.sql && uvicorn main:app --host 0.0.0.0 --port 9000"]



