# Problemy o średnim priorytecie

---

## P8 — `getSetDetails` wywoływane dla tego samego setu dwukrotnie per decyzja

**Plik:** `pkg/neat/network_interface.go:99`, `122`
**Komponent:** Bot NEAT / sensory sieci

### Opis

W ramach jednego wywołania `LoadState` + `LoadEnemyInputs`, `getSetDetails` jest wywoływane dwukrotnie dla tego samego `setId`:

```go
// network_interface.go:94 — w loadSetInputs (wewnątrz LoadState)
func (s MonopolySensors) loadSetInputs(state monopoly.GameState, propertyId int, playerID int) {
    ...
    setCount, ownedByPlayer, ownedByOthers := getSetDetails(state, setId, playerID)  // iteracja #1
    ...
}

// network_interface.go:114 — w LoadEnemyInputs (wywoływane osobno)
func (s MonopolySensors) LoadEnemyInputs(state monopoly.GameState, enemyPlayerID int, propertyID int) {
    ...
    setCount, ownedByEnemy, ownedByOthers := getSetDetails(state, setId, enemyPlayerID)  // iteracja #2
    ...
}
```

Obydwa wywołania operują na tym samym `setId` (wynikającym z `propertyID`). `getSetDetails` iteruje przez `Sets[setId]` (2–4 elementy w zależności od setu) i sprawdza właściciela każdej nieruchomości.

Choć koszt pojedynczej iteracji jest mały (maks 4 elementy), jest ona wywoływana setki tysięcy razy podczas epoki.

### Kierunek poprawki

Przeliczyć statystyki setu raz i przekazać do obu metod:

```go
// Szkic
type SetStats struct {
    Total           int
    OwnedByPlayer   int
    OwnedByEnemy    int
    OwnedByOthers   int  // poza graczem i wrogiem
}

func computeSetStats(state monopoly.GameState, setId int, playerID int, enemyID int) SetStats { ... }
```

Alternatywnie: cachować wynik `getSetDetails` w kontekście `LoadState` i przekazywać dalej.

---

## P9 — `map[int]float64` zamiast slice w `getBestDecision`

**Plik:** `pkg/neat/player.go:175`, `205`, `233`, `258`, `438`
**Komponent:** Bot NEAT

### Opis

W każdej funkcji decyzyjnej bota zbierane są wyniki sieci do mapy:

```go
// player.go:175 — w tryMortgage
outputs := make(map[int]float64)
for _, propertyId := range availableActions.MortgageList {
    ...
    outputs[propertyId] = outputList[OUT_MORTGAGE]
}
decision, propertyId, _ := getBestDecision(outputs)
```

Analogicznie w `getBuyHouseDecision` (linia 205), `getBuyOutDecision` (233), `getBuyKeyPropertyDecision` (258).

`getBestDecision` następnie iteruje przez mapę szukając max:

```go
// player.go:438
func getBestDecision(decisions map[int]float64) (bool, int, float64) {
    for id, value := range decisions {
        if value > bestValue { ... }
    }
}
```

Mapy w Go mają wyższy overhead niż slice:
- Alokacja `map[int]float64` = alokacja hasha + bucketów
- Każde `map[id] = value` = obliczenie hasha + zapis
- Iteracja mapy = losowy dostęp do pamięci (gorsze cache behavior)

Przy typowej liczbie wpisów 2–5 (tyle nieruchomości zazwyczaj jest na danej liście), slice par jest szybszy.

### Kierunek poprawki

Zamiana mapy na slice par lub inline argmax:

```go
// Opcja A — inline argmax bez alokacji
bestId := -1
bestVal := 0.5  // próg

for _, propertyId := range availableActions.BuyHouseList {
    ...
    outputList := p.GetDecision(sensors)
    val := outputList[OUT_BUY_HOUSE]
    if val > bestVal {
        bestVal = val
        bestId = propertyId
    }
}
// bestId == -1 → brak decyzji
```

Eliminuje zarówno alokację mapy jak i pośrednie wywołanie `getBestDecision`.

---

## P10 — `getState()` wywoływane w `addMoney` i `setPosition` przy każdej transakcji

**Plik:** `pkg/monopoly/game.go:1255–1264`
**Komponent:** Silnik gry

### Opis

```go
// game.go:1255
func (g *Game) addMoney(player *Player, amount int) {
    player.AddMoney(amount)
    g.logger.LogWithState(fmt.Sprintf("%s receives %d$", player.Name, amount), g.getState())
}

// game.go:1260
func (g *Game) setPosition(player *Player, position int) {
    player.SetPosition(position)
    fieldName := g.fields[position].GetName()
    g.logger.LogWithState(fmt.Sprintf("%s moves to %s", player.Name, fieldName), g.getState())
}
```

`g.getState()` tworzy strukturę `GameState` zawierającą dwa slice'y wskaźników (`[]*Player`, `[]*Property`). Nawet jeśli logger jest wyłączony, `getState()` i `fmt.Sprintf()` są zawsze wywoływane.

`addMoney` jest wywoływane z wielu miejsc:
- przejście przez GO
- Chance/Community Chest
- aukcje (po każdym wygranym licytowaniu)
- transfery między graczami
- każda sprzedaż nieruchomości

`setPosition` — przy każdym ruchu gracza i każdym karcie Chance.

W sumie może być kilkanaście–kilkadziesiąt wywołań na rundę na gracza.

### Kierunek poprawki

Objęte przez poprawkę P2 (lazy logging). Guard `if g.logger.IsEnabled()` przed każdym `LogWithState` eliminuje zarówno alokację `GameState` jak i `fmt.Sprintf`.

```go
func (g *Game) addMoney(player *Player, amount int) {
    player.AddMoney(amount)
    if g.logger.IsEnabled() {
        g.logger.LogWithState(fmt.Sprintf("%s receives %d$", player.Name, amount), g.getState())
    }
}
```
