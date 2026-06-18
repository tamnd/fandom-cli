---
title: "Fetching a full article"
description: "Get wikitext, infobox fields, link graphs, and categories with the page command."
weight: 20
---

The `page` command talks to the MediaWiki Action API and returns everything the API exposes for a single article, plus parsed derivatives computed from the wikitext.

## Basic fetch

```bash
fandom --wiki starwars page "Luke Skywalker"
fandom --wiki minecraft page "Creeper"
fandom --wiki harrypotter page "Horcrux" -o json
```

## Fields

Every `page` response includes:

| Field | What it is |
|---|---|
| `id` | Numeric page ID |
| `title` | Canonical page title |
| `display_title` | Display title from `pageprops` (may differ from title) |
| `ns` | MediaWiki namespace number |
| `url` | Full canonical URL |
| `abstract` | Intro paragraph (from the API's `extracts` prop) |
| `wikitext` | Raw wikitext source |
| `plain_text` | Wikitext converted to Markdown |
| `categories` | Array of category names the page belongs to |
| `images` | Array of `File:` image references in the wikitext |
| `internal_links` | Array of `[[Page Title]]` links (the outlink graph) |
| `external_links` | Array of `[https://... text]` external links |
| `templates` | Array of template names used in the wikitext |
| `infobox_fields` | Key-value map extracted from the first infobox template |
| `thumbnail` | URL of the page thumbnail image |
| `last_editor` | Username of the most recent editor |
| `revision_id` | Revision ID of the current version |
| `page_length` | Page length in bytes (from the API) |
| `word_count` | Word count of the rendered plain text |
| `is_redirect` | True if the page is a redirect |
| `updated_at` | Timestamp of the most recent revision |

## Infobox fields

`infobox_fields` is a flat key-value map extracted from the first infobox or structured template in the page.
Keys and values are strings; wikitext markup is cleaned out.

```bash
fandom --wiki starwars page "Luke Skywalker" -o json | jq .infobox_fields
# {
#   "name": "Luke Skywalker",
#   "homeworld": "Tatooine",
#   "species": "Human",
#   "affiliation": "Rebel Alliance / New Republic / Jedi Order"
# }

fandom --wiki minecraft page "Creeper" -o json | jq .infobox_fields
```

## Link graph

`internal_links` is the full outlink graph for the page.
Use it to discover related articles or drive a focused crawl:

```bash
fandom --wiki starwars page "Death Star" -o json | jq '.internal_links[]' | head -20
```

Combine with `page` calls to follow the graph:

```bash
fandom --wiki starwars page "Death Star" -o json \
  | jq -r '.internal_links[]' \
  | while read title; do
      fandom --wiki starwars page "$title" --no-wikitext -o jsonl
    done
```

## Categories

```bash
fandom --wiki starwars page "Jedi" -o json | jq .categories
```

## Suppress wikitext for compact output

For most structured data work you do not need the raw wikitext (which can be large).
`--no-wikitext` drops it:

```bash
fandom --wiki starwars page "Luke Skywalker" --no-wikitext -o json
```

## Revision history

For the full edit history of a page, use `revisions`:

```bash
fandom --wiki starwars revisions "Darth Vader" -n 20
fandom --wiki starwars revisions "Death Star" -o jsonl | jq '.editor' | sort | uniq -c | sort -rn
```
