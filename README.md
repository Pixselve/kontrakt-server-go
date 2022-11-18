# Kontrakt server

## Usage with Docker

Find an example in [docker-compose.yml](docker-compose.yml).

| Environment variable | Description                                            |
|----------------------|--------------------------------------------------------|
| DATABASE_URL         | The URL to the postgresql database                     |
| JWT_KEY              | The Json Web Token secret                              |
| PORT                 | The port the app will listen to (inside the container) |
| USERNAME             | The default teacher account username                   |
| PASSWORD             | The default teacher account password                   |
