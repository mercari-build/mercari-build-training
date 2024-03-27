FROM python:3.8-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        gcc \
        libc-dev \
    && rm -rf /var/lib/apt/lists/*


COPY python/requirements.txt requirements.txt
RUN pip install -r requirements.txt


WORKDIR /app

COPY db/ /app/db

COPY python/ /app/python

EXPOSE 9000

ENV DB_PATH=/app/db/main.sqlite3
ENV CATEGORY_DB_PATH=/app/db/category.sqlite3

CMD ["uvicorn", "python.main:app", "--host", "0.0.0.0", "--port", "9000"]