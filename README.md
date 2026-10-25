# Monopoly NEAT Bot

A Monopoly bot trained with the NEAT algorithm (using the [goNEAT](https://github.com/yaricom/goNEAT) library).

## Requirements

- Go 1.23 or newer: <https://go.dev/doc/install> (Go's `bin` directory must be on your `PATH`).
- Dependencies are downloaded on the first `go run`/`go build`. To download them up front:

  ```bash
  go mod download
  ```

## Running

Every mode is started from the project directory with a single command, with no code changes:

```bash
go run . <command> [flags]
```

| Command  | What it does |
|----------|--------------|
| `play`   | game server: human players (console clients) against bots |
| `client` | console client of a human player (connects to the server) |
| `train`  | NEAT training |
| `match`  | bots play `GAMES_PER_EPOCH` games against each other (like one training epoch) and the scores are printed |
| `debug`  | a single game of 4 bots with the full game log and decision logs of selected bots |

List of commands: `go run . help`. Flags of a command with their defaults: `go run . <command> -h`.

Instead of `go run`, you can build the program once and run the executable with the same arguments:

```bash
go build -o monopoly.exe .
./monopoly.exe train -seed 42
```

Lists of genome files are comma-separated. Paths containing spaces must be quoted:
`-genomes "genomes/bestForNow,genomes/bestForNow copy"`.

### Playing against people: `play` and `client`

1. Start the server, which listens on port `12345`:

   ```bash
   go run . play -humans 2
   ```

   | Flag      | Default              | Description |
   |-----------|----------------------|-------------|
   | `-humans` | (asked in the console) | number of human players, 0–4. Bots take the remaining seats up to 4. |
   | `-genome` | `genomes/bestForNow` | genome of the bots (all bots play with the same genome) |

2. Each human player starts a client in a separate console window and follows the instructions:

   ```bash
   go run . client
   ```

   The client connects to `localhost:12345`. The game starts once all players have connected.

With `-humans 0` only bots play, and the course of the game is shown in the server console.

### Training: `train`

```bash
go run . train -seed 42
```

| Flag          | Default             | Description |
|---------------|---------------------|-------------|
| `-seed`       | `1`                 | seed of the whole run |
| `-options`    | `neat_options.yaml` | NEAT options file |
| `-out`        | `output/seed<seed>` | output directory |
| `-genome`     | (none)              | starting genome from a file. Without this flag the starting genome is built from `INPUT_COUNT`/`OUTPUT_COUNT`. |
| `-all-to-all` | `true`              | built starting genome: every input and the bias connected to every output (`false`: no connections) |

- **Reproducibility:** the seed determines the whole run (group draws, dice rolls, mutations, crossovers). The same seed and the same options give the same result. Use a different seed for each independent run and record the seed together with the results. The default output directory includes the seed, so consecutive runs do not overwrite each other.
- **NEAT options** live in `neat_options.yaml`, among others `pop_size` (population size) and `num_generations` (number of epochs). `pop_size` must be a multiple of 4 (games have 4 players); otherwise training stops immediately with an error message.
- **Training constants** live in `pkg/config/config.go`, among others `GAMES_PER_EPOCH` (games each organism plays per epoch), scoring and `PRINT_EVERY`.
- **Stopping:** `Ctrl+C` stops training. Results saved up to that point are kept.

Contents of the output directory:

| File / directory | Description |
|------------------|-------------|
| `champions.txt` | the champion of every epoch: fitness, wins, draws |
| `gen_<N>_champion` | genome of the champion of epoch `N` (to play with it: `-genome`, `-genomes`) |
| population | the whole population every `PRINT_EVERY` (20) epochs and in the last epoch |
| `games/epoch<N>/...` | full game logs: round 0 every `PRINT_EVERY` epochs and the first 10 rounds of the last epoch |

Game logging is slow and takes a lot of disk space, about 0.8 s and 2 MB per game (`raport_wydajnosc.md`, S4). Epochs with logs therefore take noticeably longer than the others, especially the last one: 10 rounds × `pop_size`/4 games.

### Bot match: `match`

```bash
go run . match -genomes "output/seed42/gen_299_champion,output/seed42/gen_100_champion,genomes/bestForNow,genomes/bestForNow copy" -seed 7
```

| Flag       | Default                     | Description |
|------------|-----------------------------|-------------|
| `-genomes` | the 4 `genomes/bestForNow*` genomes | genomes of the bots, a multiple of 4 |
| `-seed`    | `0` (from the clock)        | seed of the group draws and the games. It is printed at the end, so the match can be repeated. |
| `-options` | `neat_options.yaml`         | NEAT options file |
| `-out`     | `experiment`                | directory for game logs |

Each bot plays `GAMES_PER_EPOCH` games in random groups of 4. Result: average score per game, number of wins and draws.

### Diagnostic game: `debug`

```bash
go run . debug -log-bots 0,2
```

| Flag        | Default                     | Description |
|-------------|-----------------------------|-------------|
| `-genomes`  | the 4 `genomes/bestForNow*` genomes | genomes of the 4 bots |
| `-log-dir`  | `debug_logs`                | directory for the game log (`game_<date>`) and the decision logs (`decision_log/`) |
| `-log-bots` | `0`                         | indexes of the bots (0–3) whose decisions (network inputs and outputs) are logged. Empty value: none. |

## Tests

```bash
go test ./...          # all tests
go test ./... -short   # without the tests that play the whole performance workload
```

### Measuring performance

A deterministic workload of 100 games (`pkg/neat/perf_test.go`) is used to compare performance before and after a change:

```powershell
.\scripts\perf.ps1 -Label before
# ... change the code ...
.\scripts\perf.ps1 -Label after
.\scripts\perf.ps1 -Compare before,after
```

Results go to `.ai_reports/perf/`. Comparing times requires `benchstat`:
`go install golang.org/x/perf/cmd/benchstat@latest`.

- A change that only affects performance must keep the **fingerprint**, i.e. the results of all games.
- A change of the bot's decisions changes the fingerprint. Then you compare the number of network activations and the time.
