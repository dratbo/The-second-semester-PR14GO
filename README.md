<h1 align="center"> Привет! Я <a target="_blank"> Кармеев Артур из группы ЭФМО-01-25 </a> 
<img src="https://github.com/blackcater/blackcater/raw/main/images/Hi.gif" height="32"/></h1>
<h3 align="center"> Данная практика была непростой :hushed:  </h3>

<h3 align="center"> Практическая работа №14: Реализация очереди задач (producer–consumer) </h3>



Структура работы:

    └── pz14-rabbitmq/
        ├── .gitignore
        ├── go.mod
        ├── go.sum
        ├── README.md
        ├── .idea/
        │   ├── .gitignore
        │   ├── modules.xml
        │   ├── pz14-rabbitmq.iml
        │   └── workspace.xml
        ├── .git/
        ├── internal/
        │   ├── rabbitsetup/
        │   │   └── queues.go
        │   ├── jobs/
        │   │   ├── publish.go
        │   │   └── task_job.go
        │   └── amqpclient/
        │       ├── config.go
        │       └── connect.go
        ├── deploy/
        │   └── rabbit/
        │       └── docker-compose.yml
        └── services/
            ├── worker/
            │   ├── internal/
            │   │   ├── store/
            │   │   │   └── processed_store.go
            │   │   └── consumer/
            │   │       ├── consumer.go
            │   │       └── process.go
            │   └── cmd/
            │       └── worker/
            │           └── main.go
            └── tasks/
                ├── internal/
                │   ├── task/
                │   │   ├── model.go
                │   │   └── repo.go
                │   ├── service/
                │   │   └── task_service.go
                │   └── http/
                │       ├── handler.go
                │       └── jobs_handler.go
                └── cmd/
                    └── tasks/
                        └── main.go


## 1. Поднять RabbitMQ

<table cellpadding="10">
  <tr>
    <td><img width="974" height="517" alt="image" src="https://github.com/user-attachments/assets/74c13338-9b68-4afd-abcc-014919fbabc5" /></td>
  </tr>
</table>

<table cellpadding="10">
  <tr>
    <td><img width="974" height="517" alt="image" src="https://github.com/user-attachments/assets/d7551f09-183b-493f-bbc4-da98ad1bf8b2" /></td>
  </tr>
</table>



## 2. Запуск системы


Запуск worker

<table cellpadding="10">
  <tr>
    <td><img width="974" height="516" alt="image" src="https://github.com/user-attachments/assets/aaba3b96-1af8-4eb9-b43e-ed7bbe9d0349" /></td>
  </tr>
</table>

Запуск tasks

<table cellpadding="10">
  <tr>
    <td><img width="974" height="516" alt="image" src="https://github.com/user-attachments/assets/e7f7c7b9-d83a-4a68-a6ab-d91ff0d5d8bd" /></td>
  </tr>
</table>


## 3. Проверить постановку обычной задачи

<table cellpadding="10">
  <tr>
    <td><img width="974" height="517" alt="image" src="https://github.com/user-attachments/assets/6e2bb1c6-3bfa-451d-86f7-37519173df83" /></td>
  </tr>
</table>

<table cellpadding="10">
  <tr>
    <td><img width="974" height="517" alt="image" src="https://github.com/user-attachments/assets/b3389431-2d32-46a1-8d12-79747be55971" /></td>
  </tr>
</table>


## 4. Проверить retries и DLQ

1-ый запрос

<table cellpadding="10">
  <tr>
    <td><img width="974" height="515" alt="image" src="https://github.com/user-attachments/assets/88a8b9fb-4452-477f-a0f0-edd7d1df9300" /></td>
  </tr>
</table>

2-ой запрос

<table cellpadding="10">
  <tr>
    <td><img width="974" height="517" alt="image" src="https://github.com/user-attachments/assets/5e90dab0-3fb6-455e-be94-9852e31d17d1" /></td>
  </tr>
</table>

3-йи запрос

<table cellpadding="10">
  <tr>
    <td><img width="974" height="517" alt="image" src="https://github.com/user-attachments/assets/01b3768f-a8e5-4751-9145-46890ff395c1" /></td>
  </tr>
</table>

Отправка на каторгу (DLQ)

<table cellpadding="10">
  <tr>
    <td><img width="974" height="518" alt="image" src="https://github.com/user-attachments/assets/4a92b45c-7132-495b-9df3-ff64c34f678a" /></td>
  </tr>
</table>


## 5. Проверить DLQ через RabbitMQ Management UI

<table cellpadding="10">
  <tr>
    <td><img width="974" height="517" alt="image" src="https://github.com/user-attachments/assets/840f404e-accd-484e-b519-c2261072d33a" /></td>
  </tr>
</table>

<table cellpadding="10">
  <tr>
    <td><img width="974" height="518" alt="image" src="https://github.com/user-attachments/assets/b7756aef-84f2-4dbe-aaba-7863e1526915" /></td>
  </tr>
</table>


## 6. Контрольные вопросы :no_mouth:

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
