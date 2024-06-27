# System wirtualnego biura turystycznego

Uruchomienie systemu wymaga zainstalowanej platformy Docker (https://www.docker.com/) oraz
własnego konta w OpenAI API (https://openai.com/index/openai-api) wraz z kluczem.

Oprócz kodu aplikacji, w repozytorium znajdują się następujące pliki:
- .env - plik z konfiguracją
- embeddings.json
- pois.json
- categories.json
- docker-compose.yaml

W pliku `.env` należy uzupełnić następujące zmienne:
- do zmiennej `OPENAI_API_KEY` należy podać swój klucz do API OpenAI
- do zmiennej `EMBEDDINGS_LOADER_SOURCE_FILE_PATH` należy podać bezwzględną ścieżkę do pliku `embeddings.json`
- do zmiennej `POIS_FILE_PATH` należy podać bezwzględną ścieżkę do pliku `pois.json`
- do zmiennej `CATEGORIES_FILE_PATH` należy podać bezwzględną ścieżkę do pliku `categories.json`


Po uzupełnieniu pliku `.env`, wewnątrz repozytorium należy wykonać poniższą komendę:

```commandline
docker compose -f docker-compose.yaml up
```

Dwa kontenery, `mongo_loader` oraz `embeddings_loader` po kilku sekundach znajdą się w stanie Exited. Jest to normalne 
i pożądane zachowanie. Gdy kontener `flask_app` wystartuje, w przeglądarce należy wpisać adres `0.0.0.0:5000`. Pierwsze
wykonanie zapytanie w aplikacji zajmuje trochę więcej czasu. Każde kolejne jest szybsze.