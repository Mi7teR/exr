# exr

exr - is Exchange Rates informer service

## Web (templ + chi)

- Шаблоны находятся в `internal/web/*.templ`.
- Генерация компонентов:

```bash
# один раз: установить генератор
# fish
set -x PATH $GOPATH/bin $PATH

# запуск генерации из корня репозитория
(cd internal/web; go generate ./...)
```

- Запуск сервера:

```bash
EXR_SQLITE_DSN=exr.db go run ./cmd/app
```

При изменении .templ перезапускайте генерацию.
