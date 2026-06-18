---
title: "Quick start"
description: "Run your first fandom commands."
weight: 30
---

Once `fandom` is on your `PATH`:

```bash
fandom version
```

## Pick a wiki

Every command takes `--wiki` to select a wiki by its subdomain slug.
The default is `starwars`.

```bash
fandom --wiki minecraft info        # wiki metadata
fandom --wiki harrypotter info
fandom --wiki onepiece info
```

## Search

```bash
fandom --wiki starwars search "lightsaber"
fandom --wiki minecraft search "creeper" -n 5
```

## Browse popular and new articles

```bash
fandom --wiki starwars top
fandom --wiki minecraft top -n 10
fandom --wiki starwars list --offset 0 -n 25
```

## Fetch a full article

The `page` command returns everything the MediaWiki API knows about a page: wikitext, rendered Markdown, categories, images, internal and external links, templates, infobox fields, thumbnail, last editor, and word count.

```bash
fandom --wiki starwars page "Luke Skywalker" -o json | jq .categories
fandom --wiki starwars page "Lightsaber" -o json | jq .infobox_fields
fandom --wiki starwars page "Death Star" --no-wikitext -o json
```

## Enumerate all pages

`allpages` streams every page stub in namespace 0.
Without `--limit` it paginates to completion.

```bash
fandom --wiki minecraft allpages | wc -l
fandom --wiki starwars allpages -n 100 -o jsonl > stubs.jsonl
```

## Revisions and site stats

```bash
fandom --wiki starwars revisions "Luke Skywalker" -n 20
fandom --wiki minecraft siteinfo
fandom --wiki starwars recent -n 30
```

## Discover wikis

```bash
fandom wikis --search "one piece"
fandom wikis --hub Anime
fandom wikis --hub Gaming -n 20
```

## Output formats

Every command defaults to a table on a terminal and JSONL in a pipe.
Switch with `-o`:

```bash
fandom --wiki starwars search "yoda" -o json
fandom --wiki minecraft allpages -o jsonl | jq .title
fandom --wiki starwars top -o csv > top.csv
```

## Where to go next

- [Browsing a wiki](/guides/browsing-a-wiki/) for the full search and list pattern.
- [Fetching a full article](/guides/full-article/) for deep dives into `page` output.
- [BFS full-wiki crawl](/guides/bfs-crawl/) to reconstruct an entire wiki.
- [Discovering wikis](/guides/wiki-discovery/) to explore Fandom's topic graph.
- [CLI reference](/reference/cli/) for every flag on every command.
