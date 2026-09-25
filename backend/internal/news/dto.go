package news

type NewsQuery struct {
	Region   string `form:"region" binding:"required,max=160"`
	District string `form:"district" binding:"omitempty,max=160"`
}

type NewsResponse struct {
	Title       string `json:"title"`
	Source      string `json:"source"`
	PublishedAt string `json:"publishedAt"`
	URL         string `json:"url"`
}