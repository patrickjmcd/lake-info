
FROM python:3

WORKDIR /app
ADD lake-info .

RUN pip3 install -r requirements.txt

CMD ["python3", "main.py"]
