---
title: "CLI"
description: "Every command and its flags."
weight: 10
---

```
fandom [--wiki <slug>] <command> [flags]
```

Run `fandom <command> --help` for the full flag list on any command.
All commands respect the global flags defined on the root command.

## Commands

| Command | API plane | What it does |
|---|---|---|
| `search <query>` | Fandom v1 | Search articles in a wiki |
| `top` | Fandom v1 | List popular articles |
| `list` | Fandom v1 | Paginated article list |
| `article <id>` | Fandom v1 | Fetch one article by numeric ID |
| `activity` | Fandom v1 | Recent edit activity |
| `info` | Fandom v1 | Wiki metadata (name, lang, topic) |
| `page <title>` | MediaWiki | Full article with wikitext, categories, links, infobox |
| `allpages` | MediaWiki | Stream all page stubs in namespace 0 (BFS seed) |
| `revisions <title>` | MediaWiki | Revision history for a page |
| `siteinfo` | MediaWiki | Wiki statistics (articles, edits, images, users) |
| `recent` | MediaWiki | Recent changes including new pages and log events |
| `wikis` | HTML scrape | Discover wikis by hub or search query |
| `version` | | Print version and exit |

---

## Global flags

These flags work on every command.

| Flag | Default | Meaning |
|---|---|---|
| `--wiki` | `starwars` | Wiki subdomain slug (e.g. `minecraft`, `harrypotter`) |
| `-o`, `--output` | `auto` | Output format: `table`, `jsonl`, `json`, `csv`, `tsv`, `url`, `raw` |
| `--fields` | | Comma-separated columns to include |
| `--no-header` | | Omit header row in `table`/`csv`/`tsv` output |
| `--template` | | Go `text/template` applied to each record |
| `-n`, `--limit` | 0 | Max records (0 = command default) |
| `-q`, `--quiet` | | Suppress progress messages on stderr |
| `--delay` | `500ms` | Minimum spacing between requests |
| `--timeout` | `15s` | Per-request timeout |
| `--retries` | `3` | Retry count on 429/5xx responses |
| `--user-agent` | | User-Agent header sent with each request |

---

## search

```
fandom --wiki <slug> search <query> [flags]
```

Searches articles using the Fandom v1 `Search/List` endpoint.

| Flag | Default | Meaning |
|---|---|---|
| `-n`, `--limit` | `20` | Max results |

**Output fields:** `id`, `title`, `url`, `abstract`, `quality`, `ns`

---

## top

```
fandom --wiki <slug> top [flags]
```

Returns popular articles from the Fandom v1 `Articles/Top` endpoint.

**Output fields:** `id`, `title`, `url`, `abstract`, `quality`, `ns`

---

## list

```
fandom --wiki <slug> list [--offset N] [flags]
```

Paginated article list from the Fandom v1 `Articles/List` endpoint.

| Flag | Default | Meaning |
|---|---|---|
| `--offset` | `0` | Pagination offset |
| `-n`, `--limit` | `25` | Page size |

**Output fields:** `id`, `title`, `url`, `abstract`, `quality`, `ns`

---

## article

```
fandom --wiki <slug> article <id> [flags]
```

Fetches one article by its numeric page ID using the Fandom v1 `Articles/Details` endpoint.

**Output fields:** `id`, `title`, `url`, `abstract`

---

## activity

```
fandom --wiki <slug> activity [flags]
```

Returns recent edit events from the Fandom v1 `Activity/LatestActivity` endpoint.

**Output fields:** `article_id`, `article_title`, `revision_id`, `timestamp`, `user`, `url`

---

## info

```
fandom --wiki <slug> info [flags]
```

Returns metadata for the wiki from the Fandom v1 `Mercury/WikiVariables` endpoint.

**Output fields:** `site_name`, `base_path`, `lang`, `topic`, `wiki`

---

## page

```
fandom --wiki <slug> page <title> [--no-wikitext] [flags]
```

Fetches a full article by title using the MediaWiki Action API.
Returns every field the API exposes, plus parsed derivatives: rendered Markdown, infobox key-value pairs, link graphs, images, and templates.

| Flag | Default | Meaning |
|---|---|---|
| `--no-wikitext` | | Omit the raw wikitext field for compact output |

**Output fields:** `id`, `title`, `display_title`, `ns`, `url`, `abstract`, `wikitext`, `plain_text`, `categories`, `images`, `internal_links`, `external_links`, `templates`, `infobox_fields`, `thumbnail`, `last_editor`, `revision_id`, `page_length`, `word_count`, `is_redirect`, `updated_at`

---

## allpages

```
fandom --wiki <slug> allpages [--continue <token>] [flags]
```

Paginates through the MediaWiki `allpages` list and streams all page stubs in namespace 0.
Without `--limit` it runs to completion, making it the BFS seed for a full-wiki crawl.

| Flag | Default | Meaning |
|---|---|---|
| `--continue` | | Resume from a previous run using the returned continue token |
| `-n`, `--limit` | `0` | Stop after N stubs (0 = all) |

**Output fields:** `id`, `ns`, `title`, `url`

---

## revisions

```
fandom --wiki <slug> revisions <title> [flags]
```

Returns the revision history of a page using the MediaWiki Action API.

| Flag | Default | Meaning |
|---|---|---|
| `-n`, `--limit` | `50` | Max revisions |

**Output fields:** `rev_id`, `page_id`, `title`, `editor`, `timestamp`, `comment`, `size`

---

## siteinfo

```
fandom --wiki <slug> siteinfo [flags]
```

Returns aggregated statistics for the wiki from the MediaWiki `siteinfo` API.

**Output fields:** `site_name`, `wiki`, `url`, `articles`, `pages`, `edits`, `images`, `users`

---

## recent

```
fandom --wiki <slug> recent [flags]
```

Returns recent changes from the MediaWiki `recentchanges` list.
Richer than `activity`: includes new page creations, log events, and categorize actions with revision IDs and edit comments.

| Flag | Default | Meaning |
|---|---|---|
| `-n`, `--limit` | `50` | Max results |

**Output fields:** `type`, `ns`, `title`, `page_id`, `rev_id`, `user`, `timestamp`, `comment`

---

## wikis

```
fandom wikis [--search <query>] [--hub <name>] [flags]
```

Discovers wikis on fandom.com.
Without flags it samples all known hubs.
Uses the Fandom f2 JSON feed when available and falls back to HTML slug extraction.

| Flag | Default | Meaning |
|---|---|---|
| `--search` | | Find wikis matching a keyword |
| `--hub` | | Browse by content hub (Gaming, Movies, TV, Books, Comics, Anime, Music, Lifestyle, Entertainment, Education) |
| `-n`, `--limit` | `25` | Max results |

**Output fields:** `slug`, `name`, `hub`, `url`
