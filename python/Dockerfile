FROM python:3.11.5-alpine

RUN addgroup -S mercari && adduser -S trainee -G mercari

WORKDIR /python

COPY . .
RUN pip install --no-cache-dir -r requirements.txt
RUN pip install --upgrade pip setuptools wheel

RUN chown -R trainee:mercari .

USER trainee

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "9000"]