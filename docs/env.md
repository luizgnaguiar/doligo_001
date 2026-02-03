# Environment Variables

This document lists all environment variables used by the application, their descriptions, whether they are mandatory or optional, and their default values. The application uses `github.com/spf13/viper` to load configuration, allowing environment variables to override default settings.

**Note:** While many variables have default values and are technically optional for application startup, it is highly recommended to explicitly configure critical settings (especially database credentials and JWT_SECRET) in a production environment for security and stability.

## Application Configuration

| Variable Name | Description                       | Mandatory/Optional | Default Value        |
| :------------ | :-------------------------------- | :----------------- | :------------------- |
| `APP_ENV`     | Application environment profile.  | Optional           | `development`        |
| `PORT`        | Port for the application to listen on. | Optional           | `8080`               |

## Database Configuration

| Variable Name           | Description                                             | Mandatory/Optional | Default Value        |
| :---------------------- | :------------------------------------------------------ | :----------------- | :------------------- |
| `DB_TYPE`               | Type of the database (e.g., `postgres`, `mysql`).       | Optional           | `postgres`           |
| `DB_HOST`               | Database host address.                                  | Optional           | `localhost`          |
| `DB_PORT`               | Database port number.                                   | Optional           | `5432`               |
| `DB_USER`               | Username for database access.                           | Optional           | `user`               |
| `DB_PASSWORD`           | Password for database access.                           | Optional           | `password`           |
| `DB_NAME`               | Name of the database to connect to.                     | Optional           | `dolibarr`           |
| `DB_SSLMODE`            | SSL mode for database connection.                       | Optional           | `disable`            |
| `DB_MAX_OPEN_CONNS`     | Maximum number of open connections to the database.     | Optional           | `10`                 |
| `DB_MAX_IDLE_CONNS`     | Maximum number of idle connections in the pool.         | Optional           | `5`                  |
| `DB_CONN_MAX_LIFETIME`  | Maximum amount of time a connection may be reused.      | Optional           | `5m0s` (5 minutes)   |

## Logging Configuration

| Variable Name | Description                               | Mandatory/Optional | Default Value |
| :------------ | :---------------------------------------- | :----------------- | :------------ |
| `LOG_LEVEL`   | Logging level (e.g., `info`, `debug`, `warn`, `error`). | Optional           | `info`        |

## Internal Worker Configuration

| Variable Name                   | Description                                                | Mandatory/Optional | Default Value        |
| :------------------------------ | :--------------------------------------------------------- | :----------------- | :------------------- |
| `INTERNAL_WORKER_POOL_SIZE`     | Size of the internal task runner worker pool.              | Optional           | `5`                  |
| `INTERNAL_WORKER_SHUTDOWN_TIMEOUT` | Timeout for graceful shutdown of the internal worker. | Optional           | `15s` (15 seconds)   |

## Authentication Configuration

| Variable Name | Description                               | Mandatory/Optional | Default Value          |
| :------------ | :---------------------------------------- | :----------------- | :--------------------- |
| `JWT_SECRET`  | Secret key used for signing JWT tokens.   | Optional           | `super-secret-jwt-key` |
