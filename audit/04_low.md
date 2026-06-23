# Problemy o niskim priorytecie

Poniższe problemy mają mały wpływ na całościowy czas treningu, ale są proste do naprawienia i poprawiają czytelność lub poprawność kodu.

---

## P11 — Pętla `for` zamiast `if` w `movePlayer`

**Plik:** `pkg/monopoly/game.go:349–358`
**Komponent:** Silnik gry

### Opis

```go
// game.go:349
func (g *Game) movePlayer(count int) {
    player := g.getCurrPlayer()
    curr_pos := player.CurrentPosition
    new_pos := curr_pos + count
    for new_pos > len(g.fields)-1 {
        g.logger.Log(fmt.Sprintf("%s passed GO...", player.Name, g.settings.StartPassMoney))
        g.addMoney(player, g.settings.StartPassMoney)
        new_pos = new_pos - len(g.fields)
    }
    g.setPosition(player, new_pos)
}
```

Maksymalny rzut kośćmi = 12 (6+6), długość planszy = 40 pól. Gracz nigdy nie obejdzie planszy więcej niż raz w jednym ruchu (`curr_pos max = 39`, `count max = 12` → `new_pos max = 51` = jedno przekroczenie granicy 39). Pętla `for` nigdy nie wykona drugiej iteracji, co może wprowadzać czytelnika w błąd i sugeruje, że scenariusz z wielokrotnym okrążeniem jest możliwy.

### Kierunek poprawki

```go
if new_pos > len(g.fields)-1 {
    g.addMoney(player, g.settings.StartPassMoney)
    new_pos -= len(g.fields)
}
```

---

## P12 — `new(SimplePlayerBot)` alokowany przy każdej grupie

**Plik:** `pkg/neat/evaluator.go:254`
**Komponent:** Trening (tryb z `INCLUDE_HEURISTIC_BOT = true`)

### Opis

```go
// evaluator.go:254
func (e *MonopolyEvaluator) prepareGroupsWithHeuristicBot(players []MonopolyPlayer, groupSize int) [][]MonopolyPlayer {
    ...
    for i := 0; i < len(players); i += (e.groupSize - 1) {
        group := make([]MonopolyPlayer, 0, e.groupSize)
        group = append(group, players[i:end]...)
        group = append(group, new(SimplePlayerBot))   // ← alokacja przy każdej grupie
        ...
    }
}
```

Przy 150 organizmach, grupach po 4 i 1000 grach:
```
ceil(150 / 3) = 50 grup × 1000 gier = 50,000 alokacji SimplePlayerBot per epokę
```

`SimplePlayerBot` ma pola `score`, `wins`, `draws`, które są akumulowane, więc nie może być trivialnie współdzielony między grami. Jednak `ResetScore()` jest wywoływany między rundami w `bracketTournament`. W `single_round` bot i tak nie ma organism'u i nie wpływa na fitness NEAT — alokacja jest całkowicie zbędna poza tworzeniem instancji do funkcji API.

### Kierunek poprawki

Pula botów per epokę — stworzyć jeden `SimplePlayerBot` per slot w grupie i resetować między grami:

```go
type MonopolyEvaluator struct {
    ...
    heuristicBots []*SimplePlayerBot  // pula, tworzona raz per PlayRound
}
```

Lub prościej: stworzyć jeden bot statyczny na poziomie pakietu (bezpiecznie, bo `SimplePlayerBot.GetStdAction` nie modyfikuje stanu między wywołaniami w tej samej grze, tylko odczytuje `state`).

---

## P13 — `FindKeyProperties` → `getSetMaps` przy każdym kroku aukcji (SimplePlayerBot)

**Plik:** `pkg/neat/bot.go:257`
**Komponent:** SimplePlayerBot

### Opis

```go
// bot.go:253
func (bot *SimplePlayerBot) BiddingDecision(player int, state monopoly.GameState, ...) int {
    if state.Players[player].Money - currentPrice < 200 {
        return 0
    }
    isKeyProperty := slices.Contains(FindKeyProperties(state, player), propertyId)  // ← getSetMaps per krok
    if isKeyProperty { return currentPrice + 20 }
    ...
}
```

`FindKeyProperties` → `getSetMaps` iteruje przez wszystkie nieruchomości (~20 po wykluczeniu railroad/utility). Ta funkcja jest wywoływana przy **każdym kroku aukcji** dla każdego uczestnika. W aukcji z 3 graczami i kilkoma rundami przelicytowywania może być wywołana 10–15 razy z identycznym wynikiem (stan nieruchomości nie zmienia się podczas aukcji).

### Kierunek poprawki

Engine mógłby przekazywać wynik `FindKeyProperties` (lub wynik `getSetMaps`) jako część `GameState` albo FullActionList. Ewentualnie `BiddingDecision` mogłoby cachować wynik na czas jednej aukcji (trudniejsze bez zmiany interfejsu).

Najprostsza poprawka bez zmiany architektury: wywołać `FindKeyProperties` poza `BiddingDecision` i sprawdzać wynik na liście.

---

## P14 — Alokacja nowej mapy `usedActions` co turę

**Plik:** `pkg/monopoly/game.go:255`
**Komponent:** Silnik gry

### Opis

```go
// game.go:251
func (g *Game) resetRoundState(idx int, player *Player) {
    g.std_actions_used = 0
    g.sell_offer_tries = 0
    g.buy_offer_tries = 0
    g.usedActions = make(map[StdAction]map[int]bool)  // ← nowa mapa przy każdej turze
}
```

Stara mapa jest odrzucana do GC i nowa tworzona przy każdym `resetRoundState()`.

```
50 rund × 4 graczy = 200 alokacji map na grę × 1000 gier = 200,000 alokacji per epokę
```

### Kierunek poprawki

**Opcja A** — Czyszczenie istniejącej mapy zamiast tworzenia nowej (Go 1.21+):
```go
clear(g.usedActions)
```

**Opcja B** — Zamiana na slice (bardziej efektywne dla małej liczby wpisów — max `MAX_STD_ACTIONS × MAX_ACTIONS = 35`):
```go
type usedAction struct { action StdAction; propertyId int }
usedActionsList []usedAction

// sprawdzenie:
func (g *Game) isUsed(action StdAction, propertyId int) bool {
    for _, ua := range g.usedActionsList {
        if ua.action == action && ua.propertyId == propertyId { return true }
    }
    return false
}
// reset:
g.usedActionsList = g.usedActionsList[:0]  // bez alokacji
```

---

## P15 — `getActivePlayers` wywoływane kilkukrotnie bez zmiany stanu

**Plik:** `pkg/monopoly/game.go:259–264`, `1212–1233`
**Komponent:** Silnik gry

### Opis

W `endGame()` (linia 259):

```go
func (g *Game) endGame() {
    active_players := g.getActivePlayers()   // iteracja #1
    if len(active_players) == 0 {
        g.endDraw()
    } else if len(active_players) == 1 {
        g.endWinner(g.players[active_players[0]])
    } else if g.round > g.settings.MaxRounds {
        g.endRoundLimit()  // wewnątrz: g.getActivePlayers() × 2  (linia 272, 275)
    }
}
```

`endRoundLimit` (linia 269) wywołuje `getActivePlayers()` ponownie, mimo że wynik nie zmienił się od wywołania w `endGame()`.

W `bankrupt()` (linia 1212):

```go
func (g *Game) bankrupt(player *Player, creditor *Player) {
    ...
    active_players := g.getActivePlayers()   // iteracja #1
    if len(active_players) <= 1 { g.finished = true }
    ...
    for _, property := range lostProperties {
        if !g.finished {
            queue := g.getAuctionQueue(active_players[g.randomSource.Intn(len(active_players))])
            g.auction(...)
        }
    }
}
```

`getAuctionQueue` (linia 1125) wewnętrznie nie wywołuje `getActivePlayers`, więc tu nie ma duplikatu — ale `endRoundLimit` tak.

### Kierunek poprawki

W `endGame` przekazać wynik pierwszego wywołania do `endRoundLimit`:

```go
func (g *Game) endGame() {
    active := g.getActivePlayers()
    if len(active) == 0 {
        g.endDraw()
    } else if len(active) == 1 {
        g.endWinner(g.players[active[0]])
    } else if g.round > g.settings.MaxRounds {
        g.endRoundLimit(active)  // przekazujemy gotowy wynik
    }
}
```
