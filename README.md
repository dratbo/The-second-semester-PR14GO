# Практическая работа №14 — очередь задач (producer–consumer)

Продолжение ПЗ 13. Вместо события task.created — задача (job) в очереди task_jobs с retries и DLQ.

## Структура проекта

```
pz14-rabbitmq/
  deploy/rabbit/
    docker-compose.yml          # RabbitMQ (pz14-rabbitmq)

  internal/
    amqpclient/                 # подключение к RabbitMQ
    jobs/                       # формат TaskJob
    rabbitsetup/                # объявление task_jobs и task_jobs_dlq

  services/
    tasks/
      cmd/tasks/main.go         # HTTP API + producer jobs (шаг 4)
      internal/
        http/                   # REST handlers
        jobs/                   # публикация job в очередь
        service/                # бизнес-логика
        task/                   # модель Task, in-memory repo

    worker/
      cmd/worker/main.go        # consumer
      internal/
        consumer/               # обработка task_jobs
        store/                  # идемпотентность (message_id)
```

## Очереди

- task_jobs — основная очередь задач
- task_jobs_dlq — проблемные сообщения после исчерпания попыток

## Запуск (после реализации шагов)

RabbitMQ:
```
cd deploy/rabbit
docker compose up -d
```

Tasks:
```
go run ./services/tasks/cmd/tasks
```

Worker:
```
go run ./services/worker/cmd/worker
```

Management UI: http://localhost:15672 (guest / guest)

## REST API tasks (из ПЗ 13)

GET /v1/tasks, GET /v1/tasks/{id}, POST /v1/tasks, PATCH, DELETE, GET /health

Endpoint для jobs (шаг 4):
POST /v1/jobs/process-task

## Шаг 2. Формат сообщения задачи

Структура в internal/jobs/task_job.go

Обязательные поля JSON:

| Поле | Описание |
|------|----------|
| job | тип задачи, у нас process_task |
| task_id | id бизнес-объекта, например t_001 |
| attempt | номер попытки, при постановке = 1 |
| message_id | уникальный id сообщения (uuid) |

Пример сообщения в очереди task_jobs:

```json
{
  "job": "process_task",
  "task_id": "t_001",
  "attempt": 1,
  "message_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

Создание в коде:

```go
job := jobs.NewProcessTaskJob("t_001", uuid.New().String())
```

Отличие от ПЗ 13: там было событие (event, ts), здесь — задача на выполнение работы с attempt и message_id для retries и идемпотентности.

## Шаг 3. Объявление очередей

Пакет internal/rabbitsetup/queues.go

Очереди:
- task_jobs — основная, durable=true
- task_jobs_dlq — DLQ, durable=true

У task_jobs заданы x-dead-letter-exchange (default) и x-dead-letter-routing-key → task_jobs_dlq. Основной перевод в DLQ — вручную в worker после 3 неудачных попыток.

Если при старте ошибка PRECONDITION_FAILED — удалить старые очереди в UI (Queues → Delete) или пересоздать контейнер:
```
docker compose down
docker compose up -d
```

DeclareQueues вызывается при старте tasks и worker.

Проверка:

```
cd deploy/rabbit
docker compose up -d

go run ./services/worker/cmd/worker
```

В логе: `queues declared: task_jobs, task_jobs_dlq`

В http://localhost:15672 → Queues — обе очереди должны появиться.

## Шаг 4. Постановка задачи в очередь

POST /v1/jobs/process-task

Тело запроса:
```json
{"task_id": "t_001"}
```

Ответ (202 Accepted):
```json
{"status": "accepted", "task_id": "t_001"}
```

Сервис проверяет task_id, формирует TaskJob (attempt=1, message_id=uuid), публикует в task_jobs.

Запуск:
```
$env:RABBIT_URL="amqp://guest:guest@localhost:5672/"
go run ./services/tasks/cmd/tasks
```

Проверка:
```
[System.IO.File]::WriteAllText("job.json", '{"task_id":"t_001"}')
curl.exe -i -X POST http://localhost:8082/v1/jobs/process-task -H "Content-Type: application/json" -d "@job.json"
```

В RabbitMQ UI в очереди task_jobs появится сообщение.

## Шаг 5. Worker (consumer)

services/worker/internal/consumer — читает task_jobs, обрабатывает job, ack/retry/DLQ.

Порядок обработки:
1. получить сообщение из task_jobs
2. проверить message_id (идемпотентность, store)
3. processTask — имитация работы 2 сек
4. успех → MarkDone + ack
5. ошибка → attempt++, retry в task_jobs (макс. 3 попытки) или в task_jobs_dlq

t_fail — всегда ошибка (для проверки retries и DLQ).

Запуск:
```
$env:RABBIT_URL="amqp://guest:guest@localhost:5672/"
go run ./services/worker/cmd/worker
```

Порядок: RabbitMQ → worker → tasks → POST /v1/jobs/process-task

Успех (t_001):
```
curl.exe -i -X POST http://localhost:8082/v1/jobs/process-task -H "Content-Type: application/json" -d "@job.json"
```
В логе worker: processing → done → ack

Ошибка (t_fail):
```
[System.IO.File]::WriteAllText("job-fail.json", '{"task_id":"t_fail"}')
curl.exe -i -X POST http://localhost:8082/v1/jobs/process-task -H "Content-Type: application/json" -d "@job-fail.json"
```
В логе: 3 попытки, затем moved to dlq. Сообщение в task_jobs_dlq (UI).

## Шаг 12. Retries и DLQ (t_fail)

В repo есть задача t_fail — worker всегда падает на ней.

PowerShell:
```
[System.IO.File]::WriteAllText("job-fail.json", '{"task_id":"t_fail"}')
curl.exe -i -X POST http://localhost:8082/v1/jobs/process-task -H "Content-Type: application/json" -d "@job-fail.json"
```

Ожидаемый ответ tasks: HTTP 202 Accepted.

В логе worker (каждая попытка ~2 сек):
```
processing task_id=t_fail attempt=1
process error ...
retry scheduled attempt=2
processing task_id=t_fail attempt=2
...
moved to dlq task_id=t_fail
```

Если в worker тишина — проверь:
1. worker запущен до curl
2. tasks перезапущен после добавления t_fail
3. curl вернул 202, а не 404 task not found
4. RabbitMQ запущен

## Контрольные вопросы

**1. Чем задача в очереди отличается от простого события?**

Событие (ПЗ 13) — короткое уведомление «что-то произошло» (task.created). Задача (job) — работа на выполнение: может длиться долго, падать с ошибкой, требовать повторов. У job есть attempt, message_id, логика retries и DLQ.

**2. Зачем нужны retries?**

Повторные попытки нужны при временных ошибках: сеть, занятый ресурс, кратковременный сбой. Вместо немедленного отказа задача получает ещё шанс обработаться.

**3. Почему нельзя бесконечно возвращать ошибочное сообщение в основную очередь?**

Безнадёжно битое сообщение будет крутиться в цикле, занимать worker, блокировать обработку нормальных задач и забивать очередь. Нужен лимит попыток и вывод в DLQ.

**4. Что такое DLQ и зачем она используется?**

DLQ (Dead Letter Queue) — очередь проблемных сообщений. Сюда попадают задачи после исчерпания попыток. Сообщения не теряются, основная очередь не блокируется, можно разобрать ошибку вручную.

**5. Почему в системах очередей возможна повторная доставка одного и того же сообщения?**

Доставка at-least-once: если worker обработал сообщение, но не отправил ack до сбоя, RabbitMQ считает его неподтверждённым и может отдать снова.

**6. Что такое идемпотентность обработчика?**

Повторная обработка того же сообщения не меняет результат повторно. Второй раз с тем же message_id не выполняет работу заново — только ack.

**7. Зачем нужен message_id?**

Уникальный id сообщения, не путать с task_id. По message_id отличают дубликаты при повторной доставке и ведут учёт обработанных сообщений.

**8. Почему хранение обработанных message_id даже в памяти полезно для учебного примера?**

Достаточно, чтобы показать принцип идемпотентности без БД. map[string]bool наглядно: повторное сообщение с тем же id пропускается. В проде — Redis или БД.

**9. Что произойдёт, если worker выполнит обработку, но не успеет отправить ack?**

RabbitMQ не получит подтверждение. Сообщение останется неподтверждённым и может быть доставлено другому consumer или тому же после перезапуска — без идемпотентности работа выполнится дважды.

**10. Почему модель producer–consumer удобна для тяжёлых фоновых задач?**

HTTP-сервис быстро принимает запрос (202 Accepted) и ставит job в очередь. Тяжёлую работу делает отдельный worker асинхронно. Можно масштабировать workers независимо, клиент не ждёт минуты обработки.
