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

type wireWikiVarsResp struct {
	Data wireWikiData `json:"data"`
}

type wireWikiData struct {
	BasePath string `json:"basePath"`
	SiteName string `json:"siteName"`
	Lang     string `json:"lang"`
	Topic    string `json:"topic"`
}
