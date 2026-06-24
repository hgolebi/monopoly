#!/usr/bin/env python3
"""
Sensitivity analysis for NEAT genome.

Usage:
  python analyze_genome.py <genome_file> <output_id>

output_id:
  0 = OUT_BUY_PROPERTY
  1 = OUT_SELL_PROPERTY
  2 = OUT_MORTGAGE
  3 = OUT_BUYOUT
  4 = OUT_BUY_HOUSE

Options:
  --steps N       sweep steps per input (default: 100)
  --iters N       network activation iterations (default: 10)
  --baseline V    default value for non-swept inputs, 0.0-1.0 (default: 0.5)
  --top N         how many inputs to show in the plot (default: 8)
  --no-plot       skip matplotlib plot
  --all-plots     generate a plot for every output, not just the selected one

Examples:
  python analyze_genome.py genomes/to_debug 0
  python analyze_genome.py genomes/to_debug 2 --steps 200 --baseline 0.0
"""

import sys
import math
import argparse
from collections import defaultdict

# ── Names matching InputID iota order in network_interface.go ──────────────
INPUT_NAMES = [
    "IN_PLAYER_IS_JAILED",
    "IN_PLAYER_POSITION",
    "IN_PLAYER_MONEY",
    "IN_PROPERTY_ID",
    "IN_PROPERTY_TYPE",
    "IN_PROPERTY_PRICE",
    "IN_SET_ID",
    "IN_SET_PROPERTIES_NEEDED",
    "IN_SET_PROPERTIES_OCCUPIED",
    "IN_ENEMY_INVOLVED",
    "IN_ENEMY_MONEY",
    "IN_ENEMY_SET_PROPERTIES_NEEDED",
    "IN_ENEMY_SET_PROPERTIES_OCCUPIED",
    "IN_ENEMY_SELL_COUNTEROFFER",
    "IN_ENEMY_BUY_COUNTEROFFER",
    "IN_NEGOTIATION_LAST_TRY",
    "IN_AVAILABLE_PROPERTIES",
    "IN_ROUND",
    "IN_BUY_PRICE",
    "IN_SELL_PRICE",
]

OUTPUT_NAMES = [
    "OUT_BUY_PROPERTY",
    "OUT_SELL_PROPERTY",
    "OUT_MORTGAGE",
    "OUT_BUYOUT",
    "OUT_BUY_HOUSE",
]

# Node type constants (4th field in genome `node` line)
NTYPE_HIDDEN = 0
NTYPE_INPUT  = 1
NTYPE_OUTPUT = 2
NTYPE_BIAS   = 3

# Output nodes are 22-26 in this genome (output_id + 22)
OUTPUT_NODE_BASE = 22
DECISION_THRESHOLD = 0.5


# ── Activation functions ────────────────────────────────────────────────────

def sigmoid_steepened(x: float) -> float:
    """goNEAT SigmoidSteepenedActivation used on hidden and output nodes."""
    try:
        return 1.0 / (1.0 + math.exp(-4.924273 * x))
    except OverflowError:
        return 0.0 if x < 0 else 1.0


# ── Genome parser ────────────────────────────────────────────────────────────

def parse_genome(filepath: str):
    """
    Returns:
      nodes: dict[node_id] -> {'type': int, 'activation': str}
      connections: list of (src, dst, weight, is_recurrent, enabled)
    """
    nodes = {}
    connections = []

    with open(filepath) as f:
        for raw in f:
            parts = raw.strip().split()
            if not parts:
                continue

            if parts[0] == 'node':
                # node <id> <trait> <frozen> <type> <activation>
                nid        = int(parts[1])
                ntype      = int(parts[4])
                activation = parts[5]
                nodes[nid] = {'type': ntype, 'activation': activation}

            elif parts[0] == 'gene':
                # gene <trait> <src> <dst> <weight> <recurrent> <innov> <mut> <enabled>
                src          = int(parts[2])
                dst          = int(parts[3])
                weight       = float(parts[4])
                is_recurrent = parts[5].lower() == 'true'
                enabled      = parts[8].lower() == 'true'
                connections.append((src, dst, weight, is_recurrent, enabled))

    return nodes, connections


# ── Network forward pass ─────────────────────────────────────────────────────

def build_incoming(connections):
    """incoming[dst] = list of (src, weight) for enabled connections."""
    incoming = defaultdict(list)
    for src, dst, weight, _, enabled in connections:
        if enabled:
            incoming[dst].append((src, weight))
    return incoming


def _topo_order(nodes, connections):
    """
    Topological sort of non-input, non-bias nodes.
    Recurrent edges (and edges that would create cycles among outputs/hidden)
    are ignored for ordering purposes — they are still used in activation but
    values from the previous iteration are used, which is correct for RNNs.
    """
    non_input = {nid for nid, n in nodes.items()
                 if n['type'] not in (NTYPE_INPUT, NTYPE_BIAS)}

    # Build a DAG ignoring recurrent and backward edges to outputs
    adj = defaultdict(set)
    for src, dst, _, is_recurrent, enabled in connections:
        if not enabled or is_recurrent:
            continue
        if src in non_input and dst in non_input:
            adj[src].add(dst)

    # Kahn's algorithm
    in_degree = defaultdict(int)
    for u in non_input:
        for v in adj[u]:
            in_degree[v] += 1

    queue = [n for n in non_input if in_degree[n] == 0]
    order = []
    while queue:
        queue.sort()  # deterministic order
        n = queue.pop(0)
        order.append(n)
        for v in sorted(adj[n]):
            in_degree[v] -= 1
            if in_degree[v] == 0:
                queue.append(v)

    # Any nodes not reached (cycles) go at the end
    remaining = sorted(non_input - set(order))
    return order + remaining


def forward(nodes, incoming, process_order, input_values: list, n_iters: int = 10):
    """
    Activate the network.

    For a strictly feed-forward network n_iters=1 suffices.
    For recurrent connections (output→hidden, hidden→hidden cycles)
    multiple iterations are needed to propagate values fully.
    Values from the *previous* iteration are used for recurrent edges.
    """
    values = {nid: 0.0 for nid in nodes}

    # Bias node
    for nid, n in nodes.items():
        if n['type'] == NTYPE_BIAS:
            values[nid] = 1.0

    # Input nodes (sorted order matches INPUT_NAMES order)
    input_nodes = sorted(nid for nid, n in nodes.items() if n['type'] == NTYPE_INPUT)
    for i, nid in enumerate(input_nodes):
        values[nid] = input_values[i] if i < len(input_values) else 0.0

    for _ in range(n_iters):
        new_values = dict(values)
        for nid in process_order:
            if nid not in incoming:
                continue
            total = sum(values[src] * w for src, w in incoming[nid])
            new_values[nid] = sigmoid_steepened(total)
        values = new_values

    return values


# ── Analysis ─────────────────────────────────────────────────────────────────

def sweep_input(nodes, incoming, process_order, output_node_id: int,
                input_idx: int, baseline: list, n_steps: int, n_iters: int):
    """Vary input[input_idx] over [0,1], return (xs, ys)."""
    xs, ys = [], []
    for step in range(n_steps + 1):
        x = step / n_steps
        inputs = list(baseline)
        inputs[input_idx] = x
        vals = forward(nodes, incoming, process_order, inputs, n_iters)
        xs.append(x)
        ys.append(vals.get(output_node_id, 0.0))
    return xs, ys


def run_sensitivity(nodes, incoming, process_order, output_node_id: int,
                    baseline_val: float, n_steps: int, n_iters: int):
    n_inputs = len(INPUT_NAMES)
    baseline = [baseline_val] * n_inputs
    results = {}
    for i, name in enumerate(INPUT_NAMES):
        xs, ys = sweep_input(nodes, incoming, process_order,
                             output_node_id, i, baseline, n_steps, n_iters)
        results[name] = (xs, ys)
    return results


def compute_importance(results):
    stats = {}
    for name, (xs, ys) in results.items():
        lo, hi = min(ys), max(ys)
        crosses = lo < DECISION_THRESHOLD <= hi
        direction = ("+" if ys[-1] > ys[0] + 1e-6
                     else "-" if ys[-1] < ys[0] - 1e-6
                     else "=")
        stats[name] = {
            'range':            hi - lo,
            'min':              lo,
            'max':              hi,
            'crosses_threshold': crosses,
            'direction':        direction,
        }
    return stats


# ── Report ───────────────────────────────────────────────────────────────────

def print_report(output_name: str, stats: dict):
    cols = 65
    sep = "=" * cols
    print(f"\n{sep}")
    print(f"  Sensitivity report -- {output_name}")
    print(f"  Decision threshold = {DECISION_THRESHOLD}")
    print(sep)
    hdr = f"  {'Input':<42} {'Range':>6}  {'Min':>5}  {'Max':>5}  {'Dir':>3}  {'Cross?':>6}"
    print(hdr)
    print(f"  {'-' * 61}")

    ranked = sorted(stats.items(), key=lambda kv: -kv[1]['range'])
    for name, s in ranked:
        cross = "YES" if s['crosses_threshold'] else "-"
        print(f"  {name:<42} {s['range']:>6.4f}  {s['min']:>5.3f}  {s['max']:>5.3f}"
              f"  {s['direction']:>3}  {cross:>6}")

    print(sep)
    influential = [n for n, s in ranked if s['crosses_threshold']]
    if influential:
        print(f"\n  Inputs that cross the 0.5 threshold ({len(influential)}):")
        for n in influential:
            print(f"    * {n}")
    else:
        print("\n  No single input crosses the decision threshold from this baseline.")


# ── Plot ──────────────────────────────────────────────────────────────────────

def plot_sensitivity(results, stats, output_name: str, top_n: int, save_path: str | None = None):
    try:
        import matplotlib.pyplot as plt
        import matplotlib.gridspec as gridspec
    except ImportError:
        print("\nmatplotlib not installed — skipping plot.  pip install matplotlib")
        return

    ranked = sorted(stats.items(), key=lambda kv: -kv[1]['range'])
    top_names = [n for n, _ in ranked[:top_n]]

    cols = 4
    rows = math.ceil(top_n / cols)
    fig = plt.figure(figsize=(cols * 4, rows * 3.2))
    gs  = gridspec.GridSpec(rows, cols, figure=fig, hspace=0.55, wspace=0.4)

    for i, name in enumerate(top_names):
        r, c = divmod(i, cols)
        ax = fig.add_subplot(gs[r, c])
        xs, ys = results[name]
        ax.plot(xs, ys, linewidth=2, color='steelblue')
        ax.axhline(DECISION_THRESHOLD, color='crimson', linestyle='--',
                   linewidth=1, alpha=0.7, label='threshold')
        short = name.replace("IN_", "")
        s = stats[name]
        title = f"{short}\n(range={s['range']:.4f} {s['direction']})"
        ax.set_title(title, fontsize=8)
        ax.set_xlabel("value [0→1]", fontsize=7)
        ax.set_ylabel("output", fontsize=7)
        ax.set_ylim(-0.05, 1.05)
        ax.tick_params(labelsize=7)
        ax.grid(True, alpha=0.25)

    fig.suptitle(
        f"Sensitivity Analysis — {output_name}\n"
        f"(other inputs held at baseline, top {top_n} by range)",
        fontsize=11, y=1.01
    )

    out_file = save_path or f"sensitivity_{output_name}.png"
    plt.savefig(out_file, dpi=150, bbox_inches='tight')
    print(f"\n  Plot saved -> {out_file}")
    plt.show()


# ── CLI ───────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(
        description="Sensitivity analysis for NEAT genome",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__,
    )
    parser.add_argument("genome",     help="Path to genome file (e.g. genomes/to_debug)")
    parser.add_argument("output_id",  type=int,
                        help="Output index 0-4  (0=BUY 1=SELL 2=MORTGAGE 3=BUYOUT 4=BUY_HOUSE)")
    parser.add_argument("--steps",    type=int,   default=100,
                        help="Sweep steps per input (default: 100)")
    parser.add_argument("--iters",    type=int,   default=10,
                        help="Forward-pass iterations for recurrent stability (default: 10)")
    parser.add_argument("--baseline", type=float, default=0.5,
                        help="Default value for non-swept inputs [0,1] (default: 0.5)")
    parser.add_argument("--top",      type=int,   default=8,
                        help="Top N inputs shown in plot (default: 8)")
    parser.add_argument("--no-plot",  action="store_true",  help="Skip plot")
    parser.add_argument("--all-plots",action="store_true",
                        help="Generate plots for all 5 outputs, not just the selected one")
    args = parser.parse_args()

    if not 0 <= args.output_id < len(OUTPUT_NAMES):
        parser.error(f"output_id must be 0–{len(OUTPUT_NAMES)-1}")

    # ── Parse & build ──
    nodes, connections = parse_genome(args.genome)
    incoming      = build_incoming(connections)
    process_order = _topo_order(nodes, connections)

    input_nodes  = [nid for nid, n in nodes.items() if n['type'] == NTYPE_INPUT]
    hidden_nodes = [nid for nid, n in nodes.items() if n['type'] == NTYPE_HIDDEN]
    active_conns = sum(1 for *_, enabled in connections if enabled)

    print(f"\n  Genome : {args.genome}")
    print(f"  Nodes  : {len(nodes)} total  "
          f"({len(input_nodes)} inputs, {len(hidden_nodes)} hidden, "
          f"{len(OUTPUT_NAMES)} outputs, 1 bias)")
    print(f"  Edges  : {active_conns} active / {len(connections)} total")
    print(f"  Baseline: {args.baseline}  Steps: {args.steps}  Iters: {args.iters}")

    output_ids = list(range(len(OUTPUT_NAMES))) if args.all_plots else [args.output_id]

    for oid in output_ids:
        output_node_id = OUTPUT_NODE_BASE + oid
        output_name    = OUTPUT_NAMES[oid]

        if output_node_id not in nodes:
            print(f"  WARNING: node {output_node_id} not found — skipping {output_name}")
            continue

        print(f"\n  Sweeping inputs for {output_name} (node {output_node_id}) ...")
        results    = run_sensitivity(nodes, incoming, process_order,
                                     output_node_id, args.baseline,
                                     args.steps, args.iters)
        importance = compute_importance(results)
        print_report(output_name, importance)

        if not args.no_plot:
            plot_sensitivity(results, importance, output_name, args.top)


if __name__ == "__main__":
    main()
