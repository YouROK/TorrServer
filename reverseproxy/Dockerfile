FROM python:3.11-slim
WORKDIR /app

EXPOSE 8080

RUN pip install --no-cache-dir -r requirements.txt

COPY . .

CMD ["python", "proxy.py", "http://backend:8000"]
