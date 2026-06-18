---
title: "Browsing a wiki"
description: "Search, list, and explore articles using the Fandom v1 API."
weight: 10
---

This guide covers the six commands that use the Fandom v1 REST API: `search`, `top`, `list`, `article`, `activity`, and `info`.
These are the fastest commands because they talk to Fandom's own CDN-cached endpoints and return compact article stubs.

## Pick a wiki

Every command takes `--wiki` to select a wiki by its subdomain slug.

```bash
fandom --wiki minecraft info
fandom --wiki harrypotter info
fandom --wiki onepiece info
```

`info` returns the wiki name, language, and topic from the `Mercury/WikiVariables` endpoint.

## Search articles

```bash
fandom --wiki starwars search "lightsaber"
fandom --wiki minecraft search "creeper" -n 5
fandom --wiki harrypotter search "horcrux" -o jsonl | jq .title
```

Results have `id`, `title`, `url`, `abstract`, `quality`, and `ns`.
`quality` is Fandom's internal quality score (0-99).
`ns` is the MediaWiki namespace number (0 = main article namespace).

## Popular articles

```bash
fandom --wiki starwars top
fandom --wiki minecraft top -n 10 -o table
```

Returns the most-viewed articles from the `Articles/Top` endpoint.
Useful for seeding a focused crawl on the most important pages first.

## Paginated article list

```bash
fandom --wiki starwars list
fandom --wiki starwars list --offset 25 -n 25
fandom --wiki starwars list -n 100 -o jsonl > first100.jsonl
```

`list` pages through the `Articles/List` endpoint in main namespace order.
Use `--offset` to pick up where you left off.
For a complete enumeration of all articles, use `allpages` instead.

## Fetch one article by ID

If you already have an article ID (e.g. from a `search` or `list` result), `article` returns full detail:

```bash
fandom --wiki starwars article 9875
fandom --wiki minecraft article 5001 -o json
```

The `article` command uses `Articles/Details`, which returns `id`, `title`, `url`, and `abstract`.
For the full article including wikitext and infobox, use `page <title>` instead.

## Recent edit activity

```bash
fandom --wiki starwars activity
fandom --wiki minecraft activity -n 50 -o jsonl | jq .article_title
```

Returns the most recent edit events: `article_id`, `article_title`, `revision_id`, `timestamp`, `user`, and `url`.
For richer change data including new page creations and log events, use `recent` instead.

## Pipe into other tools

All commands write clean JSONL in a pipe:

```bash
# Titles of the top 25 articles
fandom --wiki minecraft top | jq -r .title

# Build a CSV of article IDs and URLs for a wiki
fandom --wiki starwars list -n 500 -o csv > articles.csv

# Filter search results by quality score
fandom --wiki harrypotter search "magic" -o jsonl | jq 'select(.quality > 80)'
```
