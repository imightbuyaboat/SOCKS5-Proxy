# SOCKS5-Proxy

В данном репозитории представлена реализация безопасного **SOCKS5-прокси** с поддержкой **TCP** и **UDP**, написанная на языке **Go**.

## Структура проекта

- **SOCKS5-server** - локальный SOCKS5-прокси сервер, через который приложения могут передавать TCP/UDP трафик;
- **Relay-server** - удаленный relay-сервер, который принимает шифрованный трафик и передает его целевому хосту.

Поддерживается:

- Протокол TCP через SOCKS5 (`CONNECT`)
- Протокол UDP через SOCKS5 (`UDP_ASSOCIATE`)
- Сквозное шифрование между серверами с использованием алгоритма `ChaCha20-Poly1305`
- Обмен ключами между серверами и вычисление общего секрета для каждого соединения с использованием алгоритма `x25519`
- Handshake без аутентификации (`method: 0x00`)
- Адресация с использованием IPv4 (`atyp: 0x01`) и доменных имен (`atyp: 0x03`)
- Логирование с использованием `go.uber.org/zap`

## Конфигурация

Файл `config.json` в корне проекта содержит конфигурацию SOCKS5-сервера и Relay-сервера:

```json
{
    "listen_address": "0.0.0.0:1080",
    "remote_tcp_address": "0.0.0.0:1081",
    "remote_udp_address": "0.0.0.0:1082"
}
```

### SOCKS5-сервер

В `listen_address` указывается адрес локального хоста, на котором запускается SOCKS5-сервер (для запуска через Docker необходимо оставить `0.0.0.0` и по желанию изменить только порт).

В `remote_tcp_address` и `remote_udp_address` указывается IPv4 адрес удаленного хоста, на котором запущен Relay-сервер.

### Relay-сервер

Поле `listen_address` не используется.

В `remote_tcp_address` и `remote_udp_address` указывается IPv4 адрес удаленного хоста, на котором запущен Relay-сервер (для запуска через Docker необходимо оставить `0.0.0.0` в обоих полях и по желанию изменить только порты).

## Установка и запуск

```bash
git clone https://github.com/imightbuyaboat/SOCKS5-Proxy
cd SOCKS5-Proxy
```

SOCKS5-сервер и Relay-сервер собираются и запускаются из корня проекта с использованием соответствующих Dockerfile.

### SOCKS5-сервер

```bash
docker build -f socks5-server/Dockerfile -t socks5-server .
docker run --rm -d --network host --name socks5-server-container socks5-server
```

### Relay-сервер

```bash
docker build -f relay-server/Dockerfile -t relay-server .
docker run --rm -d --network host --name relay-server-container relay-server
```

## Тестирование

### TCP

На локальном хосте, на котором запущен SOCKS5-сервер, необходимо выполнить:

```bash
curl --socks5 127.0.0.1:1080 https://example.com
```

где `1080` - порт, который прослушивает SOCKS5-сервер.

### UDP

Для тестирования протокола UDP в корне проекта необходимо выполнить:

```cmd
go test -v socks5_tests/socks5_udp_test.go
```
