# Audyt wydajności — Monopoly NEAT Bot

**Data:** 2026-06-18
**Zakres:** silnik gry (`pkg/monopoly/`), trening (`pkg/neat/evaluator.go`), boty (`pkg/neat/player.go`, `bot.go`, `network_interface.go`)

---

## Struktura audytu

| Plik | Zawartość |
|------|-----------|
| [01_critical.md](01_critical.md) | Problemy krytyczne — największy wpływ na czas treningu |
| [02_high.md](02_high.md) | Problemy o wysokim priorytecie |
| [03_medium.md](03_medium.md) | Problemy o średnim priorytecie |
| [04_low.md](04_low.md) | Drobne optymalizacje niskiego riorytetu |

---

## Tabela zbiorcza

| # | Problem | Plik:Linia | Priorytet |
|---|---------|------------|-----------|
| 1 | Linearne wywołania sieci w `findMaxBuyPrice` / `findMinSellPrice` | `neat/player.go:287,377` | **Krytyczny** |
| 2 | `fmt.Sprintf` wykonywane przed sprawdzeniem loggera | `monopoly/game.go` (wszędzie) | **Krytyczny** |
| 3 | Podwójne wywołanie `NewNEATPlayerGroup` w `startGroup` | `neat/evaluator.go:269` | **Krytyczny** |
| 4 | Sześć oddzielnych iteracji przez properties w `standardActions` | `monopoly/game.go:467` | Wysoki |
| 5 | `loadAvailablePropertiesInput` przelicza 28 properties przy każdym wywołaniu sieci | `neat/network_interface.go:104` | Wysoki |
| 6 | `NewMonopolySensors` alokuje nowy slice przy każdym wywołaniu sieci | `neat/network_interface.go:56` | Wysoki |
| 7 | `getSetMaps` wywoływane dwukrotnie w jednym `GetStdAction` (SimplePlayerBot) | `neat/bot.go:89,129` | Wysoki |
| 8 | `getSetDetails` wywoływane dla tego samego setu dwukrotnie per decyzja | `neat/network_interface.go:99,122` | Średni |
| 9 | `map[int]float64` zamiast slice w `getBestDecision` | `neat/player.go:438` | Średni |
| 10 | `getState()` alokowany w `addMoney` i `setPosition` przy każdej transakcji | `monopoly/game.go:1255` | Średni |
| 11 | Pętla `for` zamiast `if` w `movePlayer` | `monopoly/game.go:353` | Niski |
| 12 | `new(SimplePlayerBot)` alokowany przy każdej grupie | `neat/evaluator.go:254` | Niski |
| 13 | `FindKeyProperties` → `getSetMaps` przy każdym kroku aukcji | `neat/bot.go:257` | Niski |
| 14 | Alokacja nowej mapy `usedActions` co turę zamiast czyszczenia | `monopoly/game.go:255` | Niski |
| 15 | `getActivePlayers` wywoływane kilkukrotnie bez zmiany stanu | `monopoly/game.go:259` | Niski |

---

## Gdzie szukać największego zwrotu

Problemy **#1** i **#2** razem odpowiadają prawdopodobnie za 60–80% overhead'u obliczeniowego jednej epoki treningu:

- **#1** — każda decyzja cenowa może wyzwolić do 200 forward-passów sieci zamiast ~8 (binary search)
- **#2** — setki milionów alokacji stringów przez GC przy 100 wątkach × 1000 gier, mimo że wynik nie jest nigdzie wypisywany

Naprawa tylko tych dwóch punktów powinna odczuwalnie skrócić czas epoki.
