# VK SubPub gRPC Server

Это приложение реализует шину событий по принципу Publisher-Subscriber, запускает gRPC-сервер для подписки и публикации сообщений по ключам. 
Все данные хранятся в памяти (in-memory). Конфигурация задаётся через YAML-файл.

Были использованы:
- Логирование
- Graceful Shutdown

## Структура проекта

```text
VK/
├── cmd/
│   └── server/
│       ├── main.go
│       └── config.yml
├── internal/
│   ├── gateways/
│   │   ├── generated/
│   │   │   ├── subpub.pb.go
│   │   │   └── subpub_grpc.pb.go
│   │   └── grpc/
│   │       ├── pubsub.go
│   │       ├── server.go
│   │       └── subpub.proto
│   └── repository/
│       └── inmemory/
│           ├── subpub.go
│           ├── subpub_test.go
│           └── subscriber.go
├── internal/subpub/
│   └── usecase.go
├── go.mod
├── README.md
```

## Конфигурация

Файл: `cmd/server/config.yml` -- конфиг приложения, куда можно прописать порт/хост.

Пример содержимого:

```yaml
server:
  host: localhost
  port: 8080
```

### Запуск сервера (из корня проекта)

```bash

go run cmd/server/main.go
```

Для локального тестирования устанавливаем *grpcurl*
```bash

go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

Метод Subscribe:
```bash

grpcurl -plaintext -d "{\"key\":\"some_key\"}" localhost:8080 PubSub/Subscribe
```

Метод Publish:
```bash 

grpcurl -plaintext -d "{\"key\":\"some_key\",\"data\":\"hello world\"}" localhost:8080 PubSub/Publish
```