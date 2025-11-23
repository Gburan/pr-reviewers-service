![Build Status](https://github.com/Gburan/pr-reviewers-service/actions/workflows/build.yml/badge.svg?job=build)
![Tests Status](https://github.com/Gburan/pr-reviewers-service/actions/workflows/test.yml/badge.svg?job=test)
![Lint Status](https://github.com/Gburan/pr-reviewers-service/actions/workflows/lint.yml/badge.svg?job=lint)
[![codecov](https://codecov.io/gh/Gburan/pr-reviewers-service/graph/badge.svg?token=MJRXEX58OW)](https://codecov.io/gh/Gburan/pr-reviewers-service)

# PR Reviewers Service

## 0. Базово

1. Быстрый старт - `go mod tidy`+`docker-compose up -d` или `make run-f`
2. Тк интеграционные тесты запускаются в одном докер-тест-контейнере, запуск тестов необходимо делать с
   ключом `-p 1` - `go test -p 1 ./...`
3. Изначально была мысль раскидать ручки с доступом по соответствующим ролям - по итогу сделал аутентификацию через
   конфиг файл -
   если хочется работать с jwt токеном - можно изменить значение в конфиге. `/dummyLogin` по дефолту возвращает роль
   админа, все ручки
   поддерживают как роль `USER`, так и `ADMIN`
4. Так и не понял, что имели в виду в доп.
   задании `Добавить простой эндпоинт статистики (например, количество назначений по пользователям и/или по PR)`.
   Реализовал простую ручку, которая возвращает айдишники ревьюеров с количеством вообще когда-либо всех назначенных на
   них ПРов для ревью
5. Сделал все доп.
   задания + {`кодогенерацию dto`,`юнит (их не просили xDDddDDdD), интеграционные, е2е тесты`,`CI`, `метрики`,
   `swagger`}

## 1. Описание

Сервис для работы с ревьюерами ПРов. Полный текст задания может быть найден [здесь](./task).

1. Метод `/dummyLogin`: Возвращает токен для авторизации `админа`.
2. Метод `/health`: Проверяет работоспособность сервиса и подключение к базе данных.
   Возвращает статус здоровья сервиса.
3. Метод `/pullRequest/create`: Создает ПР и автоматически назначает до 2 ревьюверов из команды автора.
   Принимает данные PR (author_id, pull_request_id, pull_request_name) и возвращает информацию о созданном PR.
4. Метод `/pullRequest/merge`: Мержит существующий Pull Request. Принимает идентификатор PR и возвращает результат
   операции мержа.
5. Метод `/pullRequest/reassign`: Заменяет одного ревьювера на другого из той же команды. Принимает идентификатор PR и
   идентификатор старого ревьювера, возвращает информацию о PR с новым ревьювером.
6. Метод `/stats/reviewers`: Получает статистику количества назначений для всех ревьюверов. Возвращает список ревьюверов
   с
   количеством PR, где они когда-либо были назначены, даже если ПР уже закрыт.
7. Метод `/team/add`: Создает новую команду с участниками (создает/обновляет пользователей). Принимает данные команды (
   название
   и список участников) и возвращает созданную команду.
8. Метод `/team/deactivateUsers`: Деактивирует нескольких пользователей в команде и обрабатывает перераспределение PR.
   Принимает
   название команды и список ID пользователей для деактивации, возвращает информацию о команде и затронутых PR.
9. Метод `/team/get`: Получает детальную информацию о команде с участниками по названию команды. Принимает название
   команды в
   параметрах запроса и возвращает информацию о команде.
10. Метод `/user/review`: Получает все Pull Request, назначенные пользователю для ревью. Принимает user_id в параметрах
    запроса
    и возвращает список PR.
11. Метод `/users/setIsActive`: Активирует или деактивирует пользователя. Принимает user_id и статус активности,
    возвращает
    обновленную информацию о пользователе.

## 2. Конфигурация

| Name                   | Type    | Default value                                                                        | Description                                               |
|------------------------|---------|--------------------------------------------------------------------------------------|-----------------------------------------------------------|
| REST_ADDRESS           | String  | `:8080`                                                                              | REST server address                                       |
| POSTGRES_CONN          | String  | `postgres://postgres:postgres@localhost:5432/pr-reviewers-service`                   | PostgreSQL connection string                              |
| REST_CONN_SETTINGS     | String  | `read_timeout: 5s, write_timeout: 5s, idle_timeout: 5m`                              | REST connection settings                                  |
| POSTGRES_POOL_SETTINGS | String  | `max_conns: 100, min_idle_conns: 20, max_conn_idle_time: 5m, max_conn_lifetime: 10m` | PostgreSQL connection pool settings                       |
| MIGRATIONS_DIR         | String  | `./migrations`                                                                       | Directory for database migrations                         |
| JWT_SECRET             | String  | `6a627a7fb025e2c5bed303316a3a1c801c1178bed303316a627a7fb67523a1c8`                   | Predefined JWT secret for authentication                  |
| LOGGING_OUTPUT         | String  | `"stdout"`                                                                           | Log output destination ("stdout", "stderr", or file path) |
| LOGGING_LEVEL          | String  | `"info"`                                                                             | Log level ("debug", "info", "warn", "error")              |
| MAX_PR_REVIEWERS       | Number  | `2`                                                                                  | Maximum number of reviewers per PR                        |
| AUTHORISATION_NEEDED   | Boolean | `false`                                                                              | Whether authorization is required                         |

## 3. Запуск

```
go mod tidy
go generate ./...
docker-compose up -d
```

или (в первый раз)

```
make run-f
```

или (с пересборкой контейнера)

```
make run-b
```

или

```
make run
```

## 4. Тестирование и кодогенерация

### 4.1 Генерация моков & ДТОшек для эндпоинтов

```
go generate ./...
```

or

```
make gen
```

### 4.2 Run tests

```
go clean -testcache
go test -p 1 ./...
```

or

```
make t-clean
```

### 4.3 Postman

`Для тестирования импортируйте` [файл](./test/postman/PrPeviewers.postman_collection.json)

...into the postman for manual testing.

## 5. Метрики

Метрики доступны [по адресу](http://localhost:3030/). [Login / Pass]: [admin / admin]

1)Перейдите в `Connections/Data sources`
-> `Add data source` -> `Prometheus` -> введите `http://prometheus:9090` в `Prometheus server URL` поле
-> `Save & Test`.

2)Перейдите в `Dashboard` -> `New` -> `Import` ->
импорт [файла](./metrics/grafana/pr_reviewers_service-1763853360845.json)

## 6. Swagger

Swagger доступен [по адресу](http://localhost:8080/swagger/)

Регенерация: `swag init -g internal/app/setup.go -o docs/rest --parseDependency --parseInternal`

## 7. Стресс-тестирование

### 7.3 Сценарии

Нагрузочное тестирование запускает два сценария параллельно. В первом сценарии постоянно создаются
новые команды и запрашивается информация о них. Во втором сценарии создаются ПРы и сразу же мержатся.

### 7.2 Результаты

| Metric                       | Value           |
|------------------------------|-----------------|
| Requests total               | 60,043          |
| Successful requests          | 60,041 (99.99%) |
| Failed requests              | 2 (0.00%)       |
| RPS (approx.)                | 998 req/s       |
| Avg HTTP request duration    | 4.27 ms         |
| Median HTTP request duration | 3.69 ms         |
| p90 HTTP request duration    | 5.5 ms          |
| p95 HTTP request duration    | 6.08 ms         |
| p99 HTTP request duration    | 13.82 ms        |
| Max HTTP request duration    | 266.16 ms       |
| VUs (min / max)              | 53 / 59         |
| Dropped iterations           | 29              |
| Iterations total             | 29,973          |

Docker окружение, в котором запускалось тестирование:

| Resource | Value   |
|----------|---------|
| CPUs     | 12      |
| RAM      | 31.3 GB |

### 7.3 Запуск стресс-теста

```
make stress-test
```

or

```
docker-compose up -d
docker-compose up k6-load-test
```
