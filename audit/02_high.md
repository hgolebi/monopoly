# Problemy o wysokim priorytecie

---

## P4 — Sześć oddzielnych iteracji przez properties w `standardActions`

**Plik:** `pkg/monopoly/game.go:467–499`
**Komponent:** Silnik gry

### Opis

Przy każdym wywołaniu `standardActions()` (rekurencyjne, wywoływane do `MAX_STD_ACTIONS=5` razy na turę) engine buduje listę dostępnych akcji poprzez sześć osobnych przebiegów:

```go
// game.go:474
action_list.MortgageList     = g.filterUsed(MORTGAGE, g.getMortgageList(g.currentPlayerIdx))
action_list.BuyOutList       = g.filterUsed(BUYOUT,   g.getBuyOutList(g.currentPlayerIdx))
action_list.SellPropertyList = g.filterUsed(SELLOFFER, g.getSellPropertyList(g.currentPlayerIdx))
action_list.BuyPropertyList  = g.filterUsed(BUYOFFER,  g.getBuyPropertyList(g.currentPlayerIdx))
action_list.BuyHouseList     = g.filterUsed(BUYHOUSE,  g.getBuyHouseList(g.currentPlayerIdx))
action_list.SellHouseList    = g.filterUsed(SELLHOUSE, g.getSellHouseList(g.currentPlayerIdx))
```

Każda z sześciu funkcji iteruje osobno:

| Funkcja | Iteruje przez | Dodatkowe iteracje |
|---------|--------------|-------------------|
| `getMortgageList` | `player.Properties` (max 28) | `checkHouses()` → `Sets[setIdx]` |
| `getBuyOutList` | `player.Properties` | — |
| `getSellPropertyList` | `player.Properties` | `checkHouses()` → `Sets[setIdx]` |
| `getBuyPropertyList` | **wszystkie 28 properties** | — |
| `getBuyHouseList` | `Sets` (~28 properties) | — |
| `getSellHouseList` | `player.Properties` | — |

`checkHouses()` (wywoływane wewnątrz `getMortgageList` i `getSellPropertyList`) sama iteruje przez `Sets[property.SetIndex]` dla każdej nieruchomości gracza — co oznacza iteracje zagnieżdżone.

### Skala problemu

Przy 5 akcjach per turę × 4 graczy × 50 rund = 1000 wywołań `standardActions` per grę.
Przy 1000 grach per epoka = 1,000,000 wywołań — każde z 6 iteracjami przez properties.

### Kierunek poprawki

Jedna pętla przez `player.Properties` zbierająca wszystkie informacje naraz:

```go
// Szkic — jedna iteracja zamiast sześciu
for _, id := range g.players[playerIdx].Properties {
    prop := g.properties[id]
    hasHouses := g.checkHouses(prop)   // wywoływane raz per property
    if !prop.IsMortgaged && !hasHouses {
        mortgageList = append(mortgageList, id)
        sellList = append(sellList, id)
    }
    if prop.IsMortgaged {
        buyOutList = append(buyOutList, id)
    }
    if prop.Houses > 0 {
        sellHouseList = append(sellHouseList, id)
    }
}
```

`getBuyPropertyList` i `getBuyHouseList` muszą zostać osobne (iterują przez inne zakresy), ale `checkHouses()` można zcachować jako mapę `propertyId → bool` obliczaną raz na wywołanie.

---

## P5 — `loadAvailablePropertiesInput` przelicza 28 properties przy każdym wywołaniu sieci

**Plik:** `pkg/neat/network_interface.go:104–112`
**Komponent:** Bot NEAT

### Opis

```go
// network_interface.go:104
func (s MonopolySensors) loadAvailablePropertiesInput(state monopoly.GameState) {
    availableProperties := 0
    for _, prop := range state.Properties {  // 28 iteracji
        if prop.Owner == nil {
            availableProperties++
        }
    }
    s[IN_AVAILABLE_PROPERTIES] = normalize(availableProperties, 0, cfg.LAST_PROPERTY_ID+1, false)
}
```

Ta metoda jest wywoływana w `LoadState()`, który jest wywoływany w każdym `GetDecision()`.

Łańcuch wywołań podczas jednej decyzji bota:
```
GetStdAction
  └─ getBuyHouseDecision
       └─ for each property in BuyHouseList:
            ├─ NewMonopolySensors()
            ├─ LoadState()
            │    └─ loadAvailablePropertiesInput()  ← 28 iteracji
            └─ GetDecision()
```

Wartość `availableProperties` **nie zmienia się przez całą turę** — nieruchomości zmieniają właściciela dopiero po decyzji gracza, a `standardActions` jest wywoływane sekwencyjnie.

### Kierunek poprawki

Obliczyć raz i cachować w `GameState`:

```go
// io.go — rozszerzenie GameState
type GameState struct {
    ...
    AvailableProperties int  // obliczane raz przy tworzeniu/aktualizacji stanu
}
```

Albo przekazywać jako parametr do `LoadState()`. Ewentualnie obliczyć w `getState()` i nie przeliczać w sensorach.

---

## P6 — `NewMonopolySensors` alokuje nowy slice przy każdym wywołaniu sieci

**Plik:** `pkg/neat/network_interface.go:56–58`
**Komponent:** Bot NEAT

### Opis

```go
// network_interface.go:56
func NewMonopolySensors() MonopolySensors {
    return make([]float64, int(INPUT_COUNT))  // alokacja heap przy każdym wywołaniu
}
```

Każde wywołanie `GetDecision` w bocie:

```go
sensors := NewMonopolySensors()   // ← alokacja 20×8 = 160 bajtów na heap
sensors.LoadState(...)
outputList := p.GetDecision(sensors)
```

W połączeniu z problemem P1 (linearne wywołania sieci) — przy 200 wywołaniach na decyzję cenową to 200 alokacji slice'a per decyzja.

Nawet bez P1, tylko `getBuyHouseDecision` może wywołać `GetDecision` dla każdej nieruchomości z `BuyHouseList` (do 28 wpisów) — to 28 alokacji na jedną fazę decyzji.

### Kierunek poprawki

Reużywanie bufora przez `NEATMonopolyPlayer`:

```go
type NEATMonopolyPlayer struct {
    ...
    sensorsBuf MonopolySensors  // bufor reużywany między wywołaniami
}

// Inicjalizacja w NewNEATMonopolyPlayer:
sensorsBuf: NewMonopolySensors(),

// Użycie zamiast alokacji:
sensors := p.sensorsBuf
clear(sensors)  // zerowanie — O(n), bez alokacji
sensors.LoadState(...)
```

Uwaga: przy goroutine-per-game każdy `NEATMonopolyPlayer` jest używany przez co najwyżej jedną goroutinę na raz (gry są kolejkowane, nie nakładają się dla tego samego organizmu), więc współdzielenie bufora jest bezpieczne.

---

## P7 — `getSetMaps` wywoływane dwukrotnie w jednym `GetStdAction` (SimplePlayerBot)

**Plik:** `pkg/neat/bot.go:89`, `129`
**Komponent:** SimplePlayerBot

### Opis

```go
// bot.go:89
func (bot *SimplePlayerBot) GetStdAction(player int, state monopoly.GameState, ...) monopoly.ActionDetails {
    ...
    fullSetProperties := findPropertiesInFullSets(state, player)  // → getSetMaps() #1
    ...
    unwantedProperties := findUnwantedProperties(state, player)   // → getSetMaps() #2
    ...
}
```

`findPropertiesInFullSets` i `findUnwantedProperties` obydwie wywołują `getSetMaps()`:

```go
// bot.go:296
func findPropertiesInFullSets(state monopoly.GameState, playerId int) []int {
    have, missing := getSetMaps(state, playerId)  // ← pełna iteracja przez 28 properties
    ...
}

// bot.go:284
func findUnwantedProperties(state monopoly.GameState, playerId int) []int {
    have, _ := getSetMaps(state, playerId)        // ← ta sama pełna iteracja
    ...
}
```

`getSetMaps` iteruje przez wszystkie 28 nieruchomości (minus railroad/utility = ~20), tworząc dwie mapy `map[int][]int` z alokacjami wewnątrz. Obie iteracje wykonują identyczną pracę — druga jest zbędna.

### Kierunek poprawki

Wywołać `getSetMaps` raz i wyprowadzić obie listy:

```go
have, missing := getSetMaps(state, player)

// Pełne sety — z have + missing
fullSetProperties := []int{}
for setIdx, props := range missing {
    if len(props) == 0 {
        fullSetProperties = append(fullSetProperties, have[setIdx]...)
    }
}

// Niechciane — z have
unwantedProperties := []int{}
for setIdx, properties := range have {
    if len(properties) == 1 && len(monopoly.Sets[setIdx]) > 2 {
        unwantedProperties = append(unwantedProperties, properties[0])
    }
}
```
