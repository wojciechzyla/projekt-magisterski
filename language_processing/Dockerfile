FROM python:3.11.9-alpine
WORKDIR /app

COPY requirements.txt entrypoint.sh ./
COPY flask_app ./flask_app
COPY db_loaders ./db_loaders
RUN pip install -r ./requirements.txt
RUN chmod +x entrypoint.sh
RUN python3 -m nltk.downloader punkt

ENTRYPOINT ["./entrypoint.sh"]