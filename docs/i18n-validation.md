# Validação de i18n (pt-BR / en-US) — Android

Data: 2026-11-24. Emulador `gotify-alarm.avd` (API 34, x86_64), APK debug do fork.

## Ambiente

- Servidor de testes: `https://gotify.svs.e7.dev` (não é o Grafana).
- Aplicativo de teste: **i18n-test** (app id 5), criado só para esta validação, com alarme
  habilitado (limiar 4) e stages cadastradas: `Estou ciente` (600 s) e
  `Estou no computador` (0 s, encerra).
- Usuário de teste: `i18n-admin` (local). Cliente Android logado via login local.
- Troca de idioma: `adb shell cmd locale set-app-locales com.github.gotify --locales pt-BR|en-US`,
 reiniciando o app. O padrão do app é `system` (detecta o idioma do sistema).

## Resultados

### pt-BR (`--locales pt-BR`)

| Recurso | Resultado observado |
| --- | --- |
| Painel / lista de mensagens | Renderiza sem erros; cartões com cantos arredondados e elevação |
| Alarme tocando | `Tocando` |
| Botão 1 | `Estou ciente` (veio das stages cadastradas) |
| Contagem regressiva | `Toca novamente em 9:55 se o próximo estágio não for confirmado` |
| Botão 2 | `Estou no computador` (label cadastrado) |
| Toque nos 2 botões | Incidentes 19 e 20 passaram a `closed` (confirmado via `GET /incident`) |
| Resolução via webhook do app de teste | `✅ teste i18n` atualizou a mensagem 284, `state=closed`, sem duplicar |

### en-US (`--locales en-US`)

| Recurso | Resultado observado |
| --- | --- |
| Painel / lista | Renderiza sem erros; datas relativas em inglês (`2 minutes ago`) |
| Markdown | Mensagem com extras `text/markdown` renderiza `negrito e link` (sem sintaxe crua) |
| Alarme tocando | `Ringing` |
| Botão 1 | `Estou ciente` (label cadastrado, independente do idioma do sistema — correto) |
| Contagem regressiva | `Rings again in 9:56 if the next stage is not confirmed` |
| Botão 2 | `Estou no computador` |
| Toque no botão 1 | Avançou para o estágio 2 com a contagem em inglês |

## Verificações técnicas

- `./gradlew :app:lintDebug` — sem erros de `MissingTranslation` (arrays e chaves de
  configuração também traduzidas/espelhadas em `values-pt-rBR`).
- `./gradlew :app:testDebugUnitTest :app:assembleDebug` — BUILD SUCCESSFUL.
- `LocaleHelper.apply()` roda no `onCreate` do `Application` e ao trocar a preferência
  `language`, que tem as opções *Follow system / Português (Brasil) / English (US)*.

## Limitações

- A troca de idioma no emulador foi feita por `cmd locale set-app-locales` (override por app),
  equivalente ao idioma do sistema para o app; não foi alterado o locale global do dispositivo.
- A tela de Settings não foi aberta via gesture do drawer no emulador (gesto de edge conflita
  com "voltar"); o seletor de idioma está declarado em `root_preferences.xml` e compilado.
