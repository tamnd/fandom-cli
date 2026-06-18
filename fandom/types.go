package fandom

// Article is the record emitted for wiki articles across search, list, top,
// and article detail commands.
type Article struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Abstract string `json:"abstract"`
	Quality  int    `json:"quality"`
	NS       int    `json:"ns"`
}

// ActivityItem is the record emitted for recent-activity entries.
type ActivityItem struct {
	ArticleID    int    `json:"article_id"`
	ArticleTitle string `json:"article_title"`
	RevisionID   int    `json:"revision_id"`
	Timestamp    string `json:"timestamp"`
	User         string `json:"user"`
	URL          string `json:"url"`
}

// WikiInfo is the record emitted by the info command.
type WikiInfo struct {
	SiteName string `json:"site_name"`
	BasePath string `json:"base_path"`
	Lang     string `json:"lang"`
	Topic    string `json:"topic"`
	Wiki     string `json:"wiki"`
}

// ─── wire types ──────────────────────────────────────────────────────────────

type wireSearchResp struct {
	Items []wireSearchItem `json:"items"`
}

type wireSearchItem struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Quality int    `json:"quality"`
	NS      int    `json:"ns"`
}

type wireArticleListResp struct {
	Items    []wireArticleItem `json:"items"`
	BasePath string            `json:"basepath"`
	Offset   int               `json:"offset"`
	Limit    int               `json:"limit"`
	Total    int               `json:"total"`
}

type wireArticleItem struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	NS       int    `json:"ns"`
	Quality  int    `json:"quality"`
	Abstract string `json:"abstract"`
}

type wireDetailsResp struct {
	Items    map[string]wireDetailItem `json:"items"`
	BasePath string                    `json:"basepath"`
}

type wireDetailItem struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Abstract string `json:"abstract"`
	URL      string `json:"url"`
}

type wireTopResp struct {
	Items    []wireArticleItem `json:"items"`
	BasePath string            `json:"basepath"`
}

type wireActivityResp struct {
	Items []wireActivityItem `json:"items"`
}

type wireActivityItem struct {
	ArticleID  int            `json:"articleId"`
	RevisionID int            `json:"revisionId"`
	Timestamp  string         `json:"timestamp"`
	User       wireUser       `json:"user"`
	Article    wireArticleRef `json:"article"`
}

type wireUser struct {
	Name string `json:"name"`
}

type wireArticleRef struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// ─── rich types ──────────────────────────────────────────────────────────────

// WikiStub is a lightweight reference to a discovered Fandom wiki.
type WikiStub struct {
	Slug string `json:"slug"`
	Name string `json:"name,omitempty"`
	Hub  string `json:"hub,omitempty"`
	URL  string `json:"url"`
}

// PageStub is one entry from MediaWiki api.php allpages.
// It is the lightweight seed record for BFS enumeration of all pages.
type PageStub struct {
	ID    int64  `json:"id"`
	NS    int    `json:"ns"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// FullArticle is the record emitted by the page command.
// It carries every field the MediaWiki Action API can return for a single page,
// plus parsed derivatives: Markdown body, infobox fields, and link graphs.
type FullArticle struct {
	ID            int64             `json:"id"`
	Title         string            `json:"title"`
	DisplayTitle  string            `json:"display_title,omitempty"`
	NS            int               `json:"ns"`
	URL           string            `json:"url"`
	Abstract      string            `json:"abstract,omitempty"`
	Wikitext      string            `json:"wikitext,omitempty"`
	PlainText     string            `json:"plain_text,omitempty"`
	Categories    []string          `json:"categories,omitempty"`
	Images        []string          `json:"images,omitempty"`
	InternalLinks []string          `json:"internal_links,omitempty"`
	ExternalLinks []string          `json:"external_links,omitempty"`
	Templates     []string          `json:"templates,omitempty"`
	InfoboxFields map[string]string `json:"infobox_fields,omitempty"`
	Thumbnail     string            `json:"thumbnail,omitempty"`
	LastEditor    string            `json:"last_editor,omitempty"`
	RevisionID    int64             `json:"revision_id,omitempty"`
	PageLength    int               `json:"page_length,omitempty"`
	WordCount     int               `json:"word_count,omitempty"`
	IsRedirect    bool              `json:"is_redirect,omitempty"`
	UpdatedAt     string            `json:"updated_at,omitempty"`
}

// Revision is one edit in the revision history of a page.
type Revision struct {
	RevID     int64  `json:"rev_id"`
	PageID    int64  `json:"page_id"`
	Title     string `json:"title"`
	Editor    string `json:"editor"`
	Timestamp string `json:"timestamp"`
	Comment   string `json:"comment,omitempty"`
	Size      int    `json:"size"`
}

// SiteInfo holds aggregated statistics and metadata for a wiki.
type SiteInfo struct {
	SiteName string `json:"site_name"`
	Wiki     string `json:"wiki"`
	URL      string `json:"url"`
	Articles int64  `json:"articles"`
	Pages    int64  `json:"pages"`
	Edits    int64  `json:"edits"`
	Images   int64  `json:"images"`
	Users    int64  `json:"users"`
}

// RecentChange is one entry from the MediaWiki recentchanges list.
type RecentChange struct {
	Type      string `json:"type"`
	NS        int    `json:"ns"`
	Title     string `json:"title"`
	PageID    int64  `json:"page_id"`
	RevID     int64  `json:"rev_id"`
	User      string `json:"user"`
	Timestamp string `json:"timestamp"`
	Comment   string `json:"comment,omitempty"`
}

// ─── MW wire types ────────────────────────────────────────────────────────────

type wireMWAllPagesResp struct {
	Query struct {
		AllPages []wireMWPageInfo `json:"allpages"`
	} `json:"query"`
	Continue struct {
		APContinue string `json:"apcontinue"`
	} `json:"continue"`
}

type wireMWPageInfo struct {
	PageID int64  `json:"pageid"`
	NS     int    `json:"ns"`
	Title  string `json:"title"`
}

type wireMWQueryResp struct {
	Query struct {
		Normalized []struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"normalized"`
		Redirects []struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"redirects"`
		Pages map[string]wireMWPage `json:"pages"`
	} `json:"query"`
}

type wireMWPage struct {
	PageID    int64                    `json:"pageid"`
	NS        int                      `json:"ns"`
	Title     string                   `json:"title"`
	Length    int                      `json:"length"`
	Touched   string                   `json:"touched"`
	LastRevID int64                    `json:"lastrevid"`
	Missing   string                   `json:"missing"`
	Redirect  *struct{}                `json:"redirect"`
	Extract   string                   `json:"extract"`
	Revisions []wireMWRevision         `json:"revisions"`
	Categories []wireMWCategory        `json:"categories"`
	PageProps map[string]any           `json:"pageprops"`
	Thumbnail *wireMWThumbnail         `json:"thumbnail"`
}

type wireMWRevision struct {
	RevID     int64  `json:"revid"`
	User      string `json:"user"`
	Timestamp string `json:"timestamp"`
	Size      int    `json:"size"`
	Comment   string `json:"comment"`
	Slots     map[string]struct {
		Content  string `json:"content"`
		Asterisk string `json:"*"`
	} `json:"slots"`
}

type wireMWCategory struct {
	NS    int    `json:"ns"`
	Title string `json:"title"`
}

type wireMWThumbnail struct {
	Source string `json:"source"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type wireMWSiteInfoResp struct {
	Query struct {
		General struct {
			SiteName string `json:"sitename"`
			Base     string `json:"base"`
		} `json:"general"`
		Statistics struct {
			Articles int64 `json:"articles"`
			Pages    int64 `json:"pages"`
			Edits    int64 `json:"edits"`
			Images   int64 `json:"images"`
			Users    int64 `json:"users"`
		} `json:"statistics"`
	} `json:"query"`
}

type wireMWRecentChangesResp struct {
	Query struct {
		RecentChanges []struct {
			Type      string `json:"type"`
			NS        int    `json:"ns"`
			Title     string `json:"title"`
			PageID    int64  `json:"pageid"`
			RevID     int64  `json:"revid"`
			User      string `json:"user"`
			Timestamp string `json:"timestamp"`
			Comment   string `json:"comment"`
		} `json:"recentchanges"`
	} `json:"query"`
}

// ─── wiki discovery wire types ────────────────────────────────────────────────

type wireF2FeedResp struct {
	Wikis []wireF2Wiki `json:"wikis"`
}

type wireF2Wiki struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Domain   string `json:"domain"`
	Hub      string `json:"hub"`
	Language string `json:"language"`
}

// ─── existing wire types ──────────────────────────────────────────────────────

type wireWikiVarsResp struct {
	Data wireWikiData `json:"data"`
}

type wireWikiData struct {
	BasePath string `json:"basePath"`
	SiteName string `json:"siteName"`
	Lang     string `json:"lang"`
	Topic    string `json:"topic"`
}
