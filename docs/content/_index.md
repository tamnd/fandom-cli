---
title: "fandom"
description: "Browse and crawl Fandom wikis from the command line."
heroTitle: "fandom, from the command line"
heroLead: "Browse and crawl any of Fandom's 350,000+ wikis. One pure-Go binary, no API key, output that pipes into the rest of your tools."
heroPrimaryURL: "/getting-started/quick-start/"
heroPrimaryText: "Get started"
---

fandom is a single binary that talks to Fandom's public APIs and returns clean structured data.
No API key, no signup, no daemon.

```bash
# search a wiki
fandom --wiki minecraft search "creeper"

# fetch a full article with wikitext, infobox, and link graph
fandom --wiki starwars page "Luke Skywalker" -o json | jq .infobox_fields

# stream every page in a wiki (BFS seed)
fandom --wiki minecraft allpages | wc -l

# discover wikis about a topic
fandom wikis --search "one piece"
```

## Where to go next

- New here? Read the [introduction](/getting-started/introduction/), then the [quick start](/getting-started/quick-start/).
- Installing? See [installation](/getting-started/installation/).
- Need every flag? The [CLI reference](/reference/cli/) is the complete command table.
- Want to crawl an entire wiki? See [BFS full-wiki crawl](/guides/bfs-crawl/).
