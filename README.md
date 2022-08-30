# seamless-api-wrapper

Основной стек:
[Golang](https://golang.org/), [Postgres](https://www.postgresql.org/), [Docker](https://www.docker.com/)

Собрать приложение: **docker-compose build**

Запуск приложения:
**docker-compose up** 

Удаление контенеров: **docker-compose down**

Схема базы данных: **init.sql**

Конфигурация приложения: **config.toml**
```
[server]
address = ":8080"
read-timeout = "4s"
write-timeout = "5s"

[postgres]
host = "db"
port = 5432
user = "user"
password = "password"
db = "db"
pool-size = 100
```

Unit тест: **go test ./internal/rpc** 

Integration тест: **make all**

#### Замечания и дальнейшие доработки

1. Покрыть код тестами
2. Добавить кэширование
3. Использовать распределенную базу данных, например cockroachdb

