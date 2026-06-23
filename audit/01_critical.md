# Problemy krytyczne

---

## P1 — Linearne wywołania sieci w `findMaxBuyPrice` / `findMinSellPrice`

**Pliki:** `pkg/neat/player.go:287–304`, `377–388`
**Komponent:** Bot NEAT

### Opis

Bot wyznacza maksymalną cenę kupna (lub minimalną cenę sprzedaży) poprzez iterowanie krokami co 10 jednostek i wywołanie pełnego forward-passu sieci przy każdym kroku.

```go
// player.go:287
func (p *NEATMonopolyPlayer) findMaxBuyPrice(...) int {
    price := minPrice
    decision := p.getBuyFromPlayerDecision(...)  // forward-pass #1
    for decision {
        price += 10
        if price >= maxPrice { return maxPrice }
        decision = p.getBuyFromPlayerDecision(...)  // forward-pass #2, #3, ...#200
    }
    return price - 10
}
```

Analogicznie `findMinSellPrice` (linia 377):

```go
func (p *NEATMonopolyPlayer) findMinSellPrice(...) (bool, int) {
    price := minPrice
    decision := p.getSellDecision(...)  // forward-pass #1
    for !decision {
        price += 10
        if price >= cfg.MAX_MONEY { return false, 0 }
        decision = p.getSellDecision(...)  // forward-pass #2...#200
    }
    return true, price
}
```

### Dlaczego to jest krytyczne

Przy zakresie cen `[0, MAX_MONEY=2000]` i kroku `10` — maksymalnie **200 wywołań sieci na jedną decyzję cenową**.

Każde wywołanie `getBuyFromPlayerDecision` / `getSellDecision`:
1. Alokuje `MonopolySensors` (nowy `[]float64` długości `INPUT_COUNT=20`)
2. Ładuje wszystkie inputy — `LoadState()` iteruje przez `state.Players`, `state.Properties`, oblicza `loadAvailablePropertiesInput()` (28 iteracji)
3. Wykonuje forward-pass sieci neuronowej

Ścieżka wywołań w trakcie negocjacji:

```
BuyFromPlayerDecision          (wywoływane przez engine co rundę negocjacji)
  └─ findMaxBuyPrice           (liniowy scan cenowy)
       └─ getBuyFromPlayerDecision × N
            └─ GetDecision
                 ├─ NewMonopolySensors()   ← alokacja
                 ├─ LoadState()            ← iteracje przez properties
                 └─ network.ForwardSteps() ← najdroższa operacja
```

W trybie `single_round`, przy 1000 gier i MAX_TRADE_ROUNDS=10, łączna liczba forward-passów tylko z `findMaxBuyPrice` może być w rzędzie **milionów na epokę**.

### Kierunek poprawki

**Binary search** — sieć jest funkcją monotoniczną względem ceny (zakładamy że im wyższa cena tym mniej bot chce kupić), więc można zastosować wyszukiwanie binarne:

```go
// Szkic — O(log 200) ≈ 8 wywołań zamiast O(200)
lo, hi := minPrice, maxPrice
for lo < hi {
    mid := (lo + hi + 10) / 2 / 10 * 10  // zaokrąglenie do wielokrotności 10
    if p.getBuyFromPlayerDecision(state, propertyId, mid, ...) {
        lo = mid
    } else {
        hi = mid - 10
    }
}
return lo
```

Alternatywnie: przeprojektowanie sieci tak by jeden forward-pass zwracał cenę bezpośrednio jako output (zamiast decyzji binarnej). To eliminuje pętlę całkowicie.

---

## P2 — `fmt.Sprintf` wykonywane zanim sprawdzono czy logger jest aktywny

**Plik:** `pkg/monopoly/game.go` (ok. 80 wywołań w całym pliku)
**Komponent:** Silnik gry

### Opis

Każde wywołanie loggera wygląda tak:

```go
// game.go:363
g.logger.Log(fmt.Sprintf("Rolled dice: %d, %d", d1, d2))

// game.go:354
g.logger.Log(fmt.Sprintf("%s passed GO and collects %d$", player.Name, g.settings.StartPassMoney))

// game.go:1196
g.logger.LogWithState(fmt.Sprintf("%s lost %d$", player.Name, amount), g.getState())
```

Wywołanie `fmt.Sprintf(...)` jest **zawsze** ewaluowane, zanim wynik trafi do metody loggera — Go nie ma lazy evaluation. Nawet jeśli logger to `NullLogger`, który natychmiast zwraca, string jest już zaalokowany i pojawi się w GC.

`LogWithState` dodatkowo wywołuje `g.getState()`, które tworzy strukturę `GameState` zawierającą slice wskaźników — nawet gdy logowanie jest wyłączone.

### Dlaczego to jest krytyczne

Szacunkowa liczba alokacji stringów przy pełnym treningu:

```
100 wątków × 1000 gier × 50 rund × ~4 graczy × ~10 Sprintf/runda
= 200,000,000 alokacji stringów
```

Wszystkie te alokacje trafiają do GC, który musi je zbierać, powodując stop-the-world pauses i zwiększone zużycie pamięci.

Dotyczy zarówno `Log()` jak i `LogWithState()`. Wywołania `LogWithState` są szczególnie kosztowne bo generują zarówno string jak i `GameState`.

Problematyczne miejsca (wybrane):
- `game.go:225` — `fmt.Sprintf` wewnątrz pętli głównej gry, przy każdym graczu
- `game.go:354` — przy każdym przejściu przez GO
- `game.go:362` — przy każdym rzucie kostką
- `game.go:1196` — `LogWithState` przy każdej transkacji finansowej
- `game.go:1257` — `LogWithState` przy każdym `addMoney`
- `game.go:1263` — `LogWithState` przy każdym `setPosition`

### Kierunek poprawki

**Opcja A — guard przed Sprintf:**
```go
if g.logger.IsEnabled() {
    g.logger.Log(fmt.Sprintf("Rolled dice: %d, %d", d1, d2))
}
```

Wymaga dodania metody `IsEnabled() bool` do interfejsu `Logger`.

**Opcja B — funkcyjny logger (closure):**
```go
g.logger.Log(func() string { return fmt.Sprintf("Rolled dice: %d, %d", d1, d2) })
```

Logger wywołuje closure tylko gdy jest aktywny. Koszt to jeden pointer zamiast alokacji stringa.

**Opcja C — ustrukturyzowane logowanie** zamiast Sprintf (np. `slog`), z lazy evaluation wbudowanym w bibliotekę.

---

## P3 — Podwójne wywołanie `NewNEATPlayerGroup` w `startGroup`

**Plik:** `pkg/neat/evaluator.go:269–276`
**Komponent:** Trening

### Opis

```go
// evaluator.go:269
func startGroup(ctx context.Context, gd GroupDetails, outputDir string) error {
    ...
    playerGroup, err := NewNEATPlayerGroup(gd.GroupID, gd.Players)  // ← tworzy grupę #1
    if err != nil {
        return fmt.Errorf(...)
    }
    playerGroup, err = NewNEATPlayerGroup(gd.GroupID, gd.Players)   // ← tworzy grupę #2, #1 trafia do GC
    if err != nil {
        return fmt.Errorf(...)
    }
    ...
}
```

Pierwsza instancja `NEATPlayerGroup` jest tworzona, sprawdzany jest błąd, a następnie **zmienna jest od razu nadpisana** nową instancją. Pierwsza instancja trafia natychmiast do GC. Jest to błąd wynikający z kopiuj-wklej.

### Dlaczego to jest krytyczne

`startGroup` jest wywoływana dla każdej gry w każdej epoce:

```
1000 gier × (n_organisms / 4) grup = tysiące wywołań na epokę
```

Przy 150 organizmach to ok. 37 grup na grę × 1000 gier = **37,000 zbędnych alokacji** `NEATPlayerGroup` na epokę, z których każda trafia do GC.

### Kierunek poprawki

Usunąć drugi blok tworzenia grupy — pierwsze wywołanie jest wystarczające:

```go
playerGroup, err := NewNEATPlayerGroup(gd.GroupID, gd.Players)
if err != nil {
    return fmt.Errorf("Error in group %d (round %d): %v", gd.GroupID, gd.Round, err)
}
// usunąć drugą identyczną parę NewNEATPlayerGroup + if err
```
