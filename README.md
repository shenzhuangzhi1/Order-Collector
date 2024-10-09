# Order Collector Docker Container

## Important Notes

1. **Database Connection**:
    - Ensure you connect to the PostgreSQL Docker container and use the `postgres.sql` file to create the database and tables required by the application.

2. **Docker Bridge Configuration**:
    - Configure Docker's bridge network to allow communication between the application container and the PostgreSQL container.
    - Use the `-e` flag to set the necessary environment variables when starting the application container.

## Example Command

```bash
docker run -e DB_HOST=<your_database_host> order-collector
```