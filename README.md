# gotfy-fork

Monorepo com o código da organização [Gotify](https://github.com/gotify), um pacote por diretório em `packages/`.

## Pacotes

| Diretório | Upstream | Descrição |
|-----------|----------|-----------|
| `packages/server` | [gotify/server](https://github.com/gotify/server) | Servidor WebSocket + web UI |
| `packages/android` | [gotify/android](https://github.com/gotify/android) | App Android |
| `packages/cli` | [gotify/cli](https://github.com/gotify/cli) | CLI para enviar mensagens |
| `packages/go-api-client` | [gotify/go-api-client](https://github.com/gotify/go-api-client) | Cliente Go gerado (Swagger) |
| `packages/plugin-api` | [gotify/plugin-api](https://github.com/gotify/plugin-api) | API de plugins do server |
| `packages/plugin-template` | [gotify/plugin-template](https://github.com/gotify/plugin-template) | Template de plugin |
| `packages/build` | [gotify/build](https://github.com/gotify/build) | Ambiente de build (server e plugins) |
| `packages/location` | [gotify/location](https://github.com/gotify/location) | Middleware Gin (hostname/scheme) |
| `packages/configor` | [gotify/configor](https://github.com/gotify/configor) | Lib de configuração Go (archived) |
| `packages/website` | [gotify/website](https://github.com/gotify/website) | Site (VitePress) |
| `packages/api-docs` | [gotify/api-docs](https://github.com/gotify/api-docs) | Swagger externo |
| `packages/contrib` | [gotify/contrib](https://github.com/gotify/contrib) | Contribuições da comunidade |
| `packages/logo` | [gotify/logo](https://github.com/gotify/logo) | Logo |
| `packages/server-aur-git` | [gotify/server-aur-git](https://github.com/gotify/server-aur-git) | Pacote AUR (build from source) |
| `packages/server-aur-bin` | [gotify/server-aur-bin](https://github.com/gotify/server-aur-bin) | Pacote AUR (binário GitHub) |

## Desenvolvimento

- **Website:** na raiz, `yarn` e `yarn website:dev` (workspace `gotify-website`).
- **Server / CLI / plugins:** siga o README de cada pacote em `packages/*` (Go, Docker, etc.).

Os clones iniciais foram feitos com `git clone --depth 1` a partir do upstream; o histórico Git de cada repositório original não está preservado neste monorepo.
