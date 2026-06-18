---
title: "Configuration"
description: "Global flags and environment variables."
weight: 20
---

fandom needs almost no configuration.
All settings can be passed as flags and none require environment variables.

## Global flags

Every command inherits these flags from the root command.

| Flag | Default | Meaning |
|---|---|---|
| `--wiki` | `starwars` | Wiki subdomain slug to query (e.g. `minecraft`, `harrypotter`, `onepiece`) |
| `-o`, `--output` | `auto` | Output format: `table` on a terminal, `jsonl` in a pipe |
| `-n`, `--limit` | `0` | Maximum records to return (0 = command default) |
| `-q`, `--quiet` | `false` | Suppress progress messages written to stderr |
| `--delay` | `500ms` | Minimum time between requests (pacing) |
| `--timeout` | `15s` | Per-request HTTP timeout |
| `--retries` | `3` | Number of retries on 429 or 5xx responses |
| `--user-agent` | see below | User-Agent sent with every request |

The default User-Agent is `fandom/dev (+https://github.com/tamnd/fandom-cli)`, with the version number substituted at release time.

## Picking a wiki

`--wiki` takes the subdomain slug of any Fandom wiki.
For `starwars.fandom.com` the slug is `starwars`.
For `minecraft.fandom.com` it is `minecraft`.
Use the `wikis` command to discover slugs for a topic you are interested in.

```bash
fandom wikis --search "dragon ball" -o jsonl | jq .slug
fandom --wiki dragonball info
```

## Output format

`auto` picks `table` when stdout is a terminal and `jsonl` when it is a pipe.
Override with `-o`:

```bash
fandom --wiki starwars top -o json
fandom --wiki starwars top -o csv > top.csv
```

See [Output formats](/reference/output/) for the full breakdown.

## Rate limiting

The default `--delay 500ms` keeps requests at two per second, which is well within Fandom's tolerance for CLI use.
For bulk work (e.g. a full `allpages` run followed by `page` calls) raise it to `--delay 1s` to be safe.

```bash
fandom --wiki minecraft allpages --delay 1s -o jsonl > stubs.jsonl
```
