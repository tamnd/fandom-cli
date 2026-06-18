package fandom

import (
	"context"
	"fmt"
	"html"
	"net/url"
	"strings"
)

// AllPages fetches one page of results from the MediaWiki allpages list.
// Pass continueToken="" to start from the beginning.
// Returns the page stubs and the next continue token ("" when done).
func (c *Client) AllPages(ctx context.Context, wiki, continueToken string, limit int) ([]PageStub, string, error) {
	wiki = c.effectiveWiki(wiki)
	if limit <= 0 || limit > 500 {
		limit = 500
	}
	rawURL := c.wikiBase(wiki) + fmt.Sprintf(
		"/api.php?action=query&list=allpages&aplimit=%d&apnamespace=0&format=json",
		limit,
	)
	if continueToken != "" {
		rawURL += "&apcontinue=" + url.QueryEscape(continueToken)
	}

	var resp wireMWAllPagesResp
	if err := c.getJSON(ctx, rawURL, &resp); err != nil {
		return nil, "", err
	}

	base := c.wikiBase(wiki)
	stubs := make([]PageStub, 0, len(resp.Query.AllPages))
	for _, p := range resp.Query.AllPages {
		stubs = append(stubs, PageStub{
			ID:    p.PageID,
			NS:    p.NS,
			Title: p.Title,
			URL:   base + "/wiki/" + strings.ReplaceAll(p.Title, " ", "_"),
		})
	}
	return stubs, resp.Continue.APContinue, nil
}

// GetPage fetches full article content by title via the MediaWiki Action API.
// It returns a FullArticle with wikitext, parsed markdown, categories, links, infobox, and more.
func (c *Client) GetPage(ctx context.Context, wiki, title string) (FullArticle, error) {
	wiki = c.effectiveWiki(wiki)
	rawURL := c.wikiBase(wiki) + fmt.Sprintf(
		"/api.php?action=query&titles=%s&prop=revisions|categories|info|pageprops|extracts|pageimages&"+
			"rvslots=main&rvprop=ids|content|timestamp|user|size&cllimit=100&exintro=1&piprop=thumbnail&pithumbsize=600&"+
			"redirects=1&format=json",
		url.QueryEscape(title),
	)

	var resp wireMWQueryResp
	if err := c.getJSON(ctx, rawURL, &resp); err != nil {
		return FullArticle{}, err
	}

	for _, p := range resp.Query.Pages {
		if p.Missing != "" || p.PageID <= 0 {
			return FullArticle{}, fmt.Errorf("page %q: %w", title, ErrNotFound)
		}
		return parseMWPage(p, c.wikiBase(wiki)), nil
	}
	return FullArticle{}, fmt.Errorf("page %q: %w", title, ErrNotFound)
}

// GetRevisions fetches the revision history of a page by title.
func (c *Client) GetRevisions(ctx context.Context, wiki, title string, limit int) ([]Revision, error) {
	wiki = c.effectiveWiki(wiki)
	if limit <= 0 {
		limit = 50
	}
	rawURL := c.wikiBase(wiki) + fmt.Sprintf(
		"/api.php?action=query&titles=%s&prop=revisions&rvprop=ids|user|timestamp|comment|size&rvlimit=%d&redirects=1&format=json",
		url.QueryEscape(title), limit,
	)

	var resp wireMWQueryResp
	if err := c.getJSON(ctx, rawURL, &resp); err != nil {
		return nil, err
	}

	for _, p := range resp.Query.Pages {
		if p.Missing != "" || p.PageID <= 0 {
			return nil, fmt.Errorf("page %q: %w", title, ErrNotFound)
		}
		revs := make([]Revision, 0, len(p.Revisions))
		for _, r := range p.Revisions {
			revs = append(revs, Revision{
				RevID:     r.RevID,
				PageID:    p.PageID,
				Title:     p.Title,
				Editor:    r.User,
				Timestamp: r.Timestamp,
				Comment:   r.Comment,
				Size:      r.Size,
			})
		}
		return revs, nil
	}
	return nil, fmt.Errorf("page %q: %w", title, ErrNotFound)
}

// GetSiteInfo fetches aggregated statistics and metadata for a wiki.
func (c *Client) GetSiteInfo(ctx context.Context, wiki string) (SiteInfo, error) {
	wiki = c.effectiveWiki(wiki)
	rawURL := c.wikiBase(wiki) + "/api.php?action=query&meta=siteinfo&siprop=statistics|general&format=json"

	var resp wireMWSiteInfoResp
	if err := c.getJSON(ctx, rawURL, &resp); err != nil {
		return SiteInfo{}, err
	}
	return SiteInfo{
		SiteName: resp.Query.General.SiteName,
		Wiki:     wiki,
		URL:      resp.Query.General.Base,
		Articles: resp.Query.Statistics.Articles,
		Pages:    resp.Query.Statistics.Pages,
		Edits:    resp.Query.Statistics.Edits,
		Images:   resp.Query.Statistics.Images,
		Users:    resp.Query.Statistics.Users,
	}, nil
}

// RecentChanges fetches recent changes via the MediaWiki Action API.
// It is richer than the Fandom v1 Activity endpoint: it includes all change types
// (edit, new, log) with revision IDs, user names, and edit comments.
func (c *Client) RecentChanges(ctx context.Context, wiki string, limit int) ([]RecentChange, error) {
	wiki = c.effectiveWiki(wiki)
	if limit <= 0 {
		limit = 50
	}
	rawURL := c.wikiBase(wiki) + fmt.Sprintf(
		"/api.php?action=query&list=recentchanges&rclimit=%d&rcprop=title|ids|user|timestamp|comment|flags&rcdir=older&format=json",
		limit,
	)

	var resp wireMWRecentChangesResp
	if err := c.getJSON(ctx, rawURL, &resp); err != nil {
		return nil, err
	}

	changes := make([]RecentChange, 0, len(resp.Query.RecentChanges))
	for _, rc := range resp.Query.RecentChanges {
		changes = append(changes, RecentChange{
			Type:      rc.Type,
			NS:        rc.NS,
			Title:     rc.Title,
			PageID:    rc.PageID,
			RevID:     rc.RevID,
			User:      rc.User,
			Timestamp: rc.Timestamp,
			Comment:   rc.Comment,
		})
	}
	return changes, nil
}

// parseMWPage converts a wireMWPage into a FullArticle.
func parseMWPage(p wireMWPage, base string) FullArticle {
	var wikitext string
	if len(p.Revisions) > 0 {
		rev := p.Revisions[0]
		if s, ok := rev.Slots["main"]; ok {
			wikitext = s.Asterisk
			if wikitext == "" {
				wikitext = s.Content
			}
		}
	}

	images := extractImages(wikitext)
	cats := extractCategories(wikitext)
	internalLinks := extractInternalLinks(wikitext)
	externalLinks := extractExternalLinks(wikitext)
	templates := extractTemplates(wikitext)
	infobox := extractInfobox(wikitext)

	catSet := make(map[string]bool)
	for _, c := range cats {
		catSet[c] = true
	}
	for _, c := range p.Categories {
		name := strings.TrimPrefix(c.Title, "Category:")
		if !catSet[name] {
			catSet[name] = true
			cats = append(cats, name)
		}
	}

	plainText := ""
	if wikitext != "" {
		plainText = wikitextToMarkdown(wikitext)
	}

	abstract := ""
	if p.Extract != "" {
		// The extracts API returns HTML; strip tags and unescape entities.
		a := reHTMLTag.ReplaceAllString(p.Extract, "")
		a = html.UnescapeString(a)
		a = reMultiNewline.ReplaceAllString(a, "\n\n")
		abstract = strings.TrimSpace(a)
	}
	if abstract == "" && plainText != "" {
		abstract = truncateText(plainText, 500)
	}

	displayTitle := p.Title
	if dt, ok := p.PageProps["displaytitle"].(string); ok && dt != "" {
		displayTitle = html.UnescapeString(dt)
	}

	var lastEditor string
	var revID int64
	var updatedAt string
	if len(p.Revisions) > 0 {
		rev := p.Revisions[0]
		lastEditor = rev.User
		revID = rev.RevID
		updatedAt = rev.Timestamp
	}

	thumbnail := ""
	if p.Thumbnail != nil {
		thumbnail = p.Thumbnail.Source
	} else if len(images) > 0 {
		// Fallback: construct a Commons thumbnail URL from the first image name.
		// Works for most File: links; the API returns a proper URL when piprop succeeds.
		name := strings.ReplaceAll(images[0], " ", "_")
		thumbnail = "https://commons.wikimedia.org/wiki/Special:FilePath/" + url.QueryEscape(name)
	}

	articleURL := base + "/wiki/" + strings.ReplaceAll(p.Title, " ", "_")

	return FullArticle{
		ID:            p.PageID,
		Title:         p.Title,
		DisplayTitle:  displayTitle,
		NS:            p.NS,
		URL:           articleURL,
		Abstract:      abstract,
		Wikitext:      wikitext,
		PlainText:     plainText,
		Categories:    cats,
		Images:        images,
		InternalLinks: internalLinks,
		ExternalLinks: externalLinks,
		Templates:     templates,
		InfoboxFields: infobox,
		Thumbnail:     thumbnail,
		LastEditor:    lastEditor,
		RevisionID:    revID,
		PageLength:    p.Length,
		WordCount:     countWords(plainText),
		IsRedirect:    p.Redirect != nil,
		UpdatedAt:     updatedAt,
	}
}
