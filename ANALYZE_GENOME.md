# analyze_genome.py — Neural Network Sensitivity Analysis

`analyze_genome.py` inspects a trained NEAT genome by measuring how much each
input affects a chosen output. It runs a full forward pass through the network
while sweeping one input at a time from `0.0` to `1.0`, keeping all other inputs
at a configurable baseline value.

---

## Prerequisites

- Python 3.10 or newer
- `matplotlib` (for plots)

```bash
pip install matplotlib
```

---

## Quick Start

```bash
python analyze_genome.py genomes/to_debug <output_id>
```

Replace `<output_id>` with one of:

| ID | Output name | Meaning |
|----|-------------|---------|
| 0  | `OUT_BUY_PROPERTY`  | Buy a property from the bank, a player, or at auction |
| 1  | `OUT_SELL_PROPERTY` | Sell a property to another player |
| 2  | `OUT_MORTGAGE`      | Mortgage a property |
| 3  | `OUT_BUYOUT`        | Lift a mortgage |
| 4  | `OUT_BUY_HOUSE`     | Build a house or hotel |

**Example — analyse the buy decision:**

```bash
python analyze_genome.py genomes/to_debug 0
```

This prints a ranked sensitivity table in the terminal and saves a PNG plot
named `sensitivity_OUT_BUY_PROPERTY.png` in the current directory.

---

## All Options

```
python analyze_genome.py <genome> <output_id> [options]

Positional arguments:
  genome       Path to the genome file (e.g. genomes/to_debug)
  output_id    Output to analyse: 0=BUY  1=SELL  2=MORTGAGE  3=BUYOUT  4=BUY_HOUSE

Optional arguments:
  --steps N       Number of sweep steps per input (default: 100).
                  Higher values give smoother curves at the cost of runtime.

  --iters N       Number of forward-pass iterations (default: 10).
                  The network may contain recurrent connections (output -> hidden).
                  Multiple iterations let activations stabilise before sampling.
                  Values of 5-15 are usually sufficient.

  --baseline V    Value held constant for all inputs that are NOT being swept,
                  in the range [0.0, 1.0] (default: 0.5).
                  Use 0.0 to simulate a "cold" game state,
                  or any domain-specific value to probe a specific scenario.

  --top N         Number of inputs shown in the plot (default: 8).
                  Inputs are ranked by their output range (highest first).

  --no-plot       Skip the matplotlib plot; print the text table only.
                  Useful in headless environments or when matplotlib is absent.

  --all-plots     Run the full analysis for all 5 outputs in one go and save
                  a separate PNG for each.
```

---

## Reading the Output Table

```
=================================================================
  Sensitivity report -- OUT_BUY_PROPERTY
  Decision threshold = 0.5
=================================================================
  Input                                       Range    Min    Max  Dir  Cross?
  -------------------------------------------------------------
  IN_BUY_PRICE                               1.0000  0.000  1.000    -     YES
  IN_SET_ID                                  0.9898  0.000  0.990    -     YES
  IN_SET_PROPERTIES_OCCUPIED                 0.2900  0.000  0.290    +       -
  ...
```

| Column | Meaning |
|--------|---------|
| **Range** | `max - min` of the output while this input is swept. The primary importance metric — larger means the network reacts more strongly. |
| **Min / Max** | Lowest and highest output values observed during the sweep. |
| **Dir** | Direction of change as the input increases: `+` = output rises, `-` = output falls, `=` = flat. |
| **Cross?** | `YES` if the output crosses the 0.5 decision threshold during the sweep. Only inputs marked `YES` can single-handedly flip the bot's decision for this output. |

The table is sorted from most to least influential.

---

## Understanding the Network

The genome file contains:

- **20 input nodes** (nodes 1-20) — one per `InputID` (see `pkg/neat/network_interface.go`)
- **1 bias node** (node 21)
- **5 output nodes** (nodes 22-26) — one per `OutputID`
- **~61 hidden nodes** — evolved by NEAT
- **~325 weighted connections** — some go from hidden/output nodes back into earlier hidden nodes (recurrent)

All hidden and output nodes use `SigmoidSteepenedActivation`:

```
f(x) = 1 / (1 + exp(-4.924273 * x))
```

Input nodes use `NullActivation` (pass-through — normalised values in `[0, 1]` are fed directly).

The 20 inputs correspond to these game-state features:

| Node | InputID | Description | Range |
|------|---------|-------------|-------|
| 1  | `IN_PLAYER_IS_JAILED`              | Whether the player is in jail              | 0 / 1 |
| 2  | `IN_PLAYER_POSITION`               | Board position                             | [0-39] -> [0,1] |
| 3  | `IN_PLAYER_MONEY`                  | Player's cash                              | [0-2000] -> [0,1] |
| 4  | `IN_PROPERTY_ID`                   | Index of the property in question          | [0-27] -> [0,1] |
| 5  | `IN_PROPERTY_TYPE`                 | 0=railroad, 0.5=utility, 1.0=street       | {0, 0.5, 1} |
| 6  | `IN_PROPERTY_PRICE`                | Price (or price + buyout if mortgaged)     | [0-2000] -> [0,1] |
| 7  | `IN_SET_ID`                        | Colour group index                         | [0-9] -> [0,1] |
| 8  | `IN_SET_PROPERTIES_NEEDED`         | Properties still needed to complete the set| [0-4] -> [0,1] |
| 9  | `IN_SET_PROPERTIES_OCCUPIED`       | Properties in the set owned by others      | [0-4] -> [0,1] |
| 10 | `IN_ENEMY_INVOLVED`                | Whether an opponent is part of this action | 0 / 1 |
| 11 | `IN_ENEMY_MONEY`                   | Opponent's cash                            | [0-2000] -> [0,1] |
| 12 | `IN_ENEMY_SET_PROPERTIES_NEEDED`   | Properties opponent needs to complete set  | [0-4] -> [0,1] |
| 13 | `IN_ENEMY_SET_PROPERTIES_OCCUPIED` | Properties in the set owned by others (excl. enemy) | [0-4] -> [0,1] |
| 14 | `IN_ENEMY_SELL_COUNTEROFFER`       | Price enemy offers to sell this property   | [0-2000] -> [0,1] |
| 15 | `IN_ENEMY_BUY_COUNTEROFFER`        | Price enemy offers to buy this property    | [0-2000] -> [0,1] |
| 16 | `IN_NEGOTIATION_LAST_TRY`          | Whether this is the final negotiation round| 0 / 1 |
| 17 | `IN_AVAILABLE_PROPERTIES`          | Number of unowned properties on the board  | [0-28] -> [0,1] |
| 18 | `IN_ROUND`                         | Current round number                       | [0-50] -> [0,1] |
| 19 | `IN_BUY_PRICE`                     | Price for the current buy action           | [0-2000] -> [0,1] |
| 20 | `IN_SELL_PRICE`                    | Price for the current sell action          | [0-2000] -> [0,1] |

---

## Example Scenarios

**1. What drives the buy decision when the game state is well-developed?**

```bash
python analyze_genome.py genomes/to_debug 0 --baseline 0.8 --steps 200
```

Sets all other inputs to `0.8` (high money, late game, etc.) and sweeps each input.

**2. Mortgage decision — minimal game state (early game, no enemies)**

```bash
python analyze_genome.py genomes/to_debug 2 --baseline 0.0
```

**3. Quick text-only scan of all five outputs**

```bash
python analyze_genome.py genomes/to_debug 0 --no-plot --all-plots
```

With `--all-plots`, the analysis runs for all 5 outputs. Adding `--no-plot`
suppresses the GUI windows while still saving the PNG files.

**4. High-resolution plot showing the 12 most influential inputs**

```bash
python analyze_genome.py genomes/to_debug 4 --steps 500 --top 12
```

---

## Limitations

- **Sensitivity is baseline-dependent.** The network is non-linear, so results at
  `--baseline 0.5` may differ from those at `0.0` or `1.0`. Try multiple baselines
  for a fuller picture.

- **Recurrent connections.** A few connections go from output or hidden nodes back
  into earlier hidden nodes. The `--iters` flag controls how many forward passes are
  run to let these stabilise. The default of 10 is conservative; lower values speed
  up the script but may give less accurate output values for inputs that activate
  recurrent paths.

- **Interactions between inputs are not captured.** The sweep varies one input at a
  time while holding the rest fixed. Non-linear interactions (e.g. "buy only when
  price is low AND money is high") are not visible in single-input curves.
