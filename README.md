# blog-api-dp
Дипломная работа профессии «Go-разработчик 2026» 

# Запустить docker PostgreSQL
docker-compose up -d
## Проверить что БД работает на порту 5432: 'listening on IPv4 address "0.0.0.0", port 5432' (опционально)
docker-compose logs postgres
# Остановить  
docker-compose down  

# Очистить данные БД  
docker-compose down -v  

## запуск  
go run cmd/api/main.go