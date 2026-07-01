package visitor

import "time"

type VisitorSession struct {
	ID            int64
	VisitorID     string
	SessionID     string
	FirstVisitTime time.Time
	LastVisitTime  time.Time
	PageCount     int
	Platform      *string
	Version       *string
	UserAgent     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type VisitorPageView struct {
	ID         int64
	VisitorID  string
	SessionID  string
	PagePath   string
	PageTitle  *string
	VisitTime  time.Time
	Duration   *int
	Referrer   *string
	CreatedAt  time.Time
}

type VisitorDailyStats struct {
	ID                  int64
	StatDate            time.Time
	UniqueVisitors      int
	TotalSessions       int
	TotalPageViews      int
	AvgSessionDuration  float64
	BounceRate          float64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type TrackVisitorRequest struct {
	VisitorID  string
	SessionID  string
	Page       string
	PageTitle  *string
	Timestamp  int64
	UserAgent  *string
	Platform   *string
	Version    *string
	Referrer   *string
	Duration   *int
}

type TrackVisitorResponse struct {
	Success bool
	Message string
}

type DailyStats struct {
	Date             string  `json:"date"`
	UniqueVisitors   int     `json:"uniqueVisitors"`
	TotalSessions    int     `json:"totalSessions"`
	TotalPageViews   int     `json:"totalPageViews"`
	AvgSessionDuration float64 `json:"avgSessionDuration"`
	BounceRate       float64 `json:"bounceRate"`
}

type TrendData struct {
	Date      string  `json:"date"`
	Visitors  int     `json:"visitors"`
	PageViews int     `json:"pageViews"`
	Sessions  int     `json:"sessions"`
	GrowthRate string `json:"growthRate"`
}

type DashboardStats struct {
	Today             DailyStats  `json:"today"`
	Yesterday         DailyStats  `json:"yesterday"`
	VisitorGrowth     string      `json:"visitorGrowth"`
	PageViewGrowth    string      `json:"pageViewGrowth"`
	Trend             []TrendData `json:"trend"`
}
