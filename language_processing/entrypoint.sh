#!/bin/sh
set -e
set -o errexit
set -o nounset

case $1 in

embeddings-loader)
  python3 /app/db_loaders/embeddings_loader.py --destination d --source_type "$SOURCE_TYPE" --source_file "$SOURCE_FILE_PATH"
  ;;

mongo-loader)
  python3 /app/db_loaders/mongo_loader.py --poi_file_path "$POIS_FILE_PATH" --categories_file_path "$CATEGORIES_FILE_PATH"
  ;;

start_server)
  cd /app/flask_app
  gunicorn -b :5000 routes:app --log-level INFO
  ;;

esac