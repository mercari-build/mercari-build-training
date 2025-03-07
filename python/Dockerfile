FROM python:3.11.7

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

RUN groupadd -r mercari && useradd -r -g mercari trainee
# RUN chown -R trainee:mercari /path/to/db

USER trainee

CMD ["python", "-V"]
