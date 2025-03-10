FROM alpine

RUN apk add --no-cache python3 py3-pip && ln -sf python3 /usr/bin/python

RUN addgroup -S mercari && adduser -S trainee -G mercari
# RUN chown -R trainee:mercari /path/to/db

RUN apk add --no-cache python3 py3-pip

WORKDIR /app

COPY . /app

RUN python3 -m venv /app/venv

RUN /app/venv/bin/pip install --no-cache-dir -r /app/requirements.txt

RUN ln -sf /usr/bin/python3 /usr/bin/python

# USER trainee

EXPOSE 9000

CMD [/app/venv/bin/uvicorn", "main:app", "--host", "0.0.0.0", "--port", "9000"]
