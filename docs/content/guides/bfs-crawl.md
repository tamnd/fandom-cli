---
title: "BFS full-wiki crawl"
description: "Use allpages and page together to reconstruct an entire wiki without missing any page."
weight: 30
---

`allpages` streams every page stub in namespace 0 from the MediaWiki `allpages` API.
Combined with `page`, it gives you a two-step pipeline to fetch every article in a wiki with complete structured content: wikitext, infobox, categories, and the full link graph.

## Step 1: seed

`allpages` enumerates every page ID and title without fetching content:

```bash
fandom --wiki starwars allpages -o jsonl > stubs.jsonl
wc -l stubs.jsonl          # how many pages the wiki has
jq .title stubs.jsonl | head -20
```

Each stub has `id`, `ns`, `title`, and `url`.
Without `--limit` the command paginates to completion automatically.

For a very large wiki, add `--delay 1s` to be polite:

```bash
fandom --wiki starwars allpages --delay 1s -o jsonl > stubs.jsonl
```

Resume a partial run with `--continue`:

```bash
# The last line of a previous run tells you where to continue.
# Pass that token:
fandom --wiki starwars allpages --continue "Lightsaber" -o jsonl >> stubs.jsonl
```

## Step 2: crawl

Pipe the title list into `page` calls to fetch full content for each article:

```bash
jq -r .title stubs.jsonl \
  | while read title; do
      fandom --wiki starwars page "$title" --no-wikitext -q -o jsonl
    done > articles.jsonl
```

`--no-wikitext` drops the raw wikitext (you still get `plain_text`, `infobox_fields`, `categories`, and `internal_links`).
`-q` suppresses progress output so only JSONL goes to stdout.

## Process the results

Once you have `articles.jsonl`, every standard JSON tool works:

```bash
# Word count distribution
jq .word_count articles.jsonl | sort -n | uniq -c

# All unique categories across the wiki
jq '.categories[]' articles.jsonl | sort -u

# Pages with an infobox
jq 'select(.infobox_fields != null) | .title' articles.jsonl

# Pages that link to "Death Star"
jq 'select(.internal_links[] | contains("Death Star")) | .title' articles.jsonl

# Redirect pages
jq 'select(.is_redirect) | .title' articles.jsonl
```

## BFS link-graph traversal

The `internal_links` field in each `page` response is the outlink graph for that page.
Use it to do a proper BFS from a starting article rather than a full wiki sweep:

```bash
# Fetch one article and extract its links
fandom --wiki starwars page "Death Star" -o json \
  | jq -r '.internal_links[]' > queue.txt

# Process the queue, adding new links as you go
while read -r title; do
  fandom --wiki starwars page "$title" --no-wikitext -q -o jsonl \
    | tee -a visited.jsonl \
    | jq -r '.internal_links[]' >> queue.txt
done < queue.txt
```

For a real BFS crawler, track visited pages to avoid cycles.

## Tips

- Set `--delay 1s` for long runs on large wikis to stay within Fandom's rate limits.
- Use `-q` to suppress progress on stderr so only JSONL goes to stdout.
- The `allpages` output is stable across runs for any given wiki snapshot; you can diff two snapshots to find new or deleted pages.
- Redirect pages (`is_redirect: true`) are included in `allpages` output. Filter them out with `jq 'select(.is_redirect == false)'` after running `page`.
