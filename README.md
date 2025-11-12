# disco-bouncer

A bouncer to ask users for their unique code upon entering a Discord server, so there isn't Panic! at the Disco(rd).

## run it

You can run the server as a Docker container. Here is an example docker-compose configuration:

```yml
services:
  discobouncer:
    image: ghcr.io/kylrth/disco-bouncer:latest
    restart: unless-stopped
    depends_on:
      - postgres
    volumes:
      - ./data/discobouncer:/data
    ports:
      - 3000:80
    environment:
      DATABASE_URL: postgres://discobouncer:SuperSecretPassword@postgres/discobouncer?sslmode=disable
      DISCORD_TOKEN: <token retrieved when creating Discord bot>
  postgres:
    image: postgres:18
    user: 1000:1000
    restart: unless-stopped
    volumes:
      - ./data/postgres:/var/lib/postgresql
    environment:
      POSTGRES_PASSWORD: SuperSecretPassword
      POSTGRES_USER: discobouncer
```

Put this in `docker-compose.yml`, create the folders `mkdir -p ./data/{discobouncer,postgres}`, and start it with `docker-compose up -d`.

If you want to run the server without turning on the Discord bot, set `DISCORD_TOKEN: disable`. The API for editing users will still work, but the Discord bot will not.

## using the client

Before using the client, you (or the server admin) need to create a new admin account:

```sh
docker-compose exec discobouncer /bouncer admin setpass testing ThisIsATest
```

Download the client from the [releases page](https://github.com/kylrth/disco-bouncer/releases), and connect using the username and password you set:

```sh
BOUNCER_USER=testing BOUNCER_PASS=ThisIsATest ./client upload -s http://localhost:3000
```

For more information about how to use the client, run `./client -h`.
