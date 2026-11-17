# Gotify iOS

App iOS do fork, equivalente funcional do módulo de presença do Android. Este pacote é um
Swift Package: abra `Package.swift` no Xcode (File → Open) para gerar o projeto e rodar os testes.

## O que existe

- `PresenceModels` / `PresenceEvaluator` — mesmo contrato e regras do avaliador Android.
- `PresenceDetectors` — geofences (`CoreLocation`) e dispositivo Bluetooth pareado (`CoreBluetooth`).
  SSID de Wi-Fi não entra: o iOS não permite ler a rede conectada sem entitlement privado.
- `GotifyWebSocket` — `/stream?presence=1`, decodificando frames `type=presence`.
- `AlarmController` — categorias de notificação com as ações dos estágios do alarme.
- `PresenceView` — status do time e override manual.

## Configuração no Xcode

1. Abra `Package.swift`, adicione um target de aplicativo iOS 17+ que dependa de `GotifyIOS`.
2. Em *Signing & Capabilities*, adicione:
   - **Background Modes**: Location updates e Uses Bluetooth LE accessories;
   - **Push Notifications** (APNs) se quiser acordar o processo morto — o servidor Gotify não
     tem relay de push, então isso exige um provider externo (fora desta entrega);
   - **Critical Alerts** (entitlement da Apple) para o som de alarme; sem ele o app usa
     notificação de prioridade alta com `UNNotificationCategory`.
3. Info.plist: `NSLocationWhenInUseUsageDescription` e `NSBluetoothAlwaysUsageDescription`.

## Testes

`swift test` (no Xcode: ⌘U). Os vetores de `PresenceEvaluatorTests` devem permanecer iguais
aos de `packages/android/app/src/test/.../PresenceEvaluatorTest.kt`.

## Fora de escopo

Regeneração do cliente gerado, relay APNs, leitura de SSID e edição de estágios do alarme no
telefone (feita na UI web do servidor).
