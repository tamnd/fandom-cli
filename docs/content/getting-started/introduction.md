---
title: "Introduction"
description: "What fandom is, how it is structured, and what it can do."
weight: 10
---

fandom is a single pure-Go binary that puts Fandom's public data behind a tool that feels like curl.
Search wikis, fetch full articles, stream every page in a wiki, or discover new wikis by topic.
There are no credentials to set up and nothing to run alongside it.

## Two API planes

fandom speaks two protocols depending on what you want.

**Fandom v1 REST API** handles the lightweight, fan-facing surfaces: search, popular articles, recent edit activity, and wiki metadata.
These endpoints return fast, terse records and are what most browsing commands use.

**MediaWiki Action API** handles the deep, structured content: full wikitext for every page, revision history, site statistics, and complete recent-changes logs.
Because every Fandom wiki runs on MediaWiki, this API is always available at `{wiki}.fandom.com/api.php`.
The `page` command uses it to return every field the API exposes, including the wikitext source, a rendered Markdown version, infobox key-value pairs, the internal link graph, images, templates, and thumbnail.

The `allpages` command paginates through the MediaWiki `allpages` list to enumerate every page ID in a wiki.
That makes it the BFS seed for a complete crawl: pipe its output into `page` calls to reconstruct the full wiki without missing a page.

## How it is structured

- `fandom/` holds the HTTP client, typed data models, wikitext parser, and all API functions. It paces requests, sets an honest User-Agent, and retries transient failures.
- `cli/` wraps the library in subcommands with shared output formats and flags.
- `cmd/fandom/` is the one-line entry point.

## Scope

fandom is a read-only client over data Fandom already serves publicly.
It reads that data and shapes it for you.
That narrow scope keeps it a single small binary with no database, no daemon, and no setup.

Next: [install it](/getting-started/installation/), then take the [quick start](/getting-started/quick-start/).
