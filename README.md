# other-mocks
Сервис - с набором заглушек

## Использование Docker
Создание контейнера
```bash
docker build -t other-mocks .
```
Запуск контейнера
```bash
docker run --publish 8080:8080 other-mocks:latest
```

## Настройки сервиса:
1. Урл для получения версии - /api/system/version
2. Переменная среды GO_OTHER_MOCKS_PORT - порт на котором будет слушать приложение
3. Переменная среды GO_OTHER_MOCK_BRANCH - версия сервиса

## Реализованные моки
1. https://www.ooocis.ru/crv
