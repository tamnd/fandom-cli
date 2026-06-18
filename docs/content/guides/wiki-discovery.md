---
title: "Discovering wikis"
description: "Find wikis by keyword or content hub."
weight: 40
---

Fandom hosts over 350,000 wikis.
The `wikis` command finds them by keyword or content category without needing any API key.

## Search by keyword

```bash
fandom wikis --search "one piece"
fandom wikis --search "dragon ball" -n 10
fandom wikis --search "minecraft" -o jsonl | jq .slug
```

Results have `slug`, `name`, `hub`, and `url`.
The slug is what you pass to `--wiki` for every other command.

## Browse by hub

Fandom organizes wikis into content hubs.

```bash
fandom wikis --hub Gaming
fandom wikis --hub Anime -n 20
fandom wikis --hub Movies -o jsonl | jq .name
```

Known hubs: Gaming, Movies, TV, Books, Comics, Anime, Music, Lifestyle, Entertainment, Education.

## Explore across all hubs

Without a flag, `wikis` samples the top results from every known hub:

```bash
fandom wikis -n 50 -o jsonl > discovered.jsonl
```

## From slug to content

Once you have a slug, use it directly with any other command:

```bash
# discover
fandom wikis --search "attack on titan" -o jsonl | jq -r .slug | head -1

# use the slug
fandom --wiki attackontitan info
fandom --wiki attackontitan top -n 10
fandom --wiki attackontitan page "Eren Yeager" -o json | jq .infobox_fields
```

## Pipe into further commands

```bash
# Check article counts for all Gaming wikis
fandom wikis --hub Gaming -n 20 -o jsonl \
  | jq -r .slug \
  | while read slug; do
      printf "%s\t" "$slug"
      fandom --wiki "$slug" siteinfo -q -o json | jq .articles
    done
```

## How discovery works

`wikis --search` hits the Fandom f2 JSON feed first.
If that returns results, they are used directly.
When the feed returns nothing (or returns an error), the command falls back to scraping `*.fandom.com` subdomain references from the Fandom search page HTML.
`wikis --hub` scrapes the topic page for the hub (e.g. `fandom.com/topics/gaming`) and supplements with the f2 feed.

Because the fallback is HTML scraping, results may vary by what Fandom shows on those pages.
The `--search` path is more reliable for targeted queries.
