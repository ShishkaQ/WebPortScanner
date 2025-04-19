# WebPortScanner
## Учебный проекты для исследования возможностей GO и сетивых технологий

## Сканирование портов 1-500 с сохранением в файл
```bash
go run main.go -host=example.com -start=1 -end=500 -output=results.txt
```

## Быстрое сканирование с 200 горутин и таймаутом 200ms
```bash
go run main.go -workers=200 -timeout=200ms
```
