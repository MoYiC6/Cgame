package visitor

import (
	"context"
	"fmt"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) TrackVisitor(ctx context.Context, req *TrackVisitorRequest) (*TrackVisitorResponse, error) {
	if err := s.validateRequest(req); err != nil {
		return &TrackVisitorResponse{Success: false, Message: err.Error()}, nil
	}

	session := &VisitorSession{
		VisitorID:  req.VisitorID,
		SessionID:  req.SessionID,
		Platform:   req.Platform,
		Version:    req.Version,
		UserAgent:  req.UserAgent,
	}
	if err := s.repo.UpsertSession(ctx, session); err != nil {
		return nil, fmt.Errorf("track visitor: %w", err)
	}

	pageView := &VisitorPageView{
		VisitorID: req.VisitorID,
		SessionID: req.SessionID,
		PagePath:  req.Page,
		PageTitle: req.PageTitle,
		VisitTime: time.UnixMilli(req.Timestamp),
		Duration:  req.Duration,
		Referrer:  req.Referrer,
	}
	if err := s.repo.BatchInsertPageViews(ctx, []*VisitorPageView{pageView}); err != nil {
		return nil, fmt.Errorf("track page view: %w", err)
	}

	return &TrackVisitorResponse{Success: true, Message: "ok"}, nil
}

func (s *Service) BatchTrack(ctx context.Context, reqs []*TrackVisitorRequest) (*TrackVisitorResponse, error) {
	if len(reqs) == 0 {
		return &TrackVisitorResponse{Success: true, Message: "ok"}, nil
	}

	var sessions []*VisitorSession
	var pageViews []*VisitorPageView
	for _, req := range reqs {
		if err := s.validateRequest(req); err != nil {
			continue
		}
		sessions = append(sessions, &VisitorSession{
			VisitorID: req.VisitorID,
			SessionID: req.SessionID,
			Platform:  req.Platform,
			Version:   req.Version,
			UserAgent: req.UserAgent,
		})
		pageViews = append(pageViews, &VisitorPageView{
			VisitorID:  req.VisitorID,
			SessionID:  req.SessionID,
			PagePath:   req.Page,
			PageTitle:  req.PageTitle,
			VisitTime:  time.UnixMilli(req.Timestamp),
			Duration:   req.Duration,
			Referrer:   req.Referrer,
		})
	}

	if len(sessions) == 0 {
		return &TrackVisitorResponse{Success: true, Message: "ok"}, nil
	}

	for _, session := range sessions {
		if err := s.repo.UpsertSession(ctx, session); err != nil {
			return nil, fmt.Errorf("batch track session: %w", err)
		}
	}
	if err := s.repo.BatchInsertPageViews(ctx, pageViews); err != nil {
		return nil, fmt.Errorf("batch track page views: %w", err)
	}

	return &TrackVisitorResponse{Success: true, Message: "ok"}, nil
}

func (s *Service) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterday := today.Add(-24 * time.Hour)

	todayStats, err := s.calculateDailyStats(ctx, today)
	if err != nil {
		return nil, fmt.Errorf("get today stats: %w", err)
	}
	yesterdayStats, err := s.calculateDailyStats(ctx, yesterday)
	if err != nil {
		return nil, fmt.Errorf("get yesterday stats: %w", err)
	}

	stats := &DashboardStats{
		Today:     *todayStats,
		Yesterday: *yesterdayStats,
	}
	if yesterdayStats.UniqueVisitors > 0 {
		stats.VisitorGrowth = fmt.Sprintf("%.1f%%", float64(todayStats.UniqueVisitors-yesterdayStats.UniqueVisitors)/float64(yesterdayStats.UniqueVisitors)*100)
	} else if todayStats.UniqueVisitors > 0 {
		stats.VisitorGrowth = "新增"
	}
	if yesterdayStats.TotalPageViews > 0 {
		stats.PageViewGrowth = fmt.Sprintf("%.1f%%", float64(todayStats.TotalPageViews-yesterdayStats.TotalPageViews)/float64(yesterdayStats.TotalPageViews)*100)
	} else if todayStats.TotalPageViews > 0 {
		stats.PageViewGrowth = "新增"
	}

	return stats, nil
}

func (s *Service) GetTrend(ctx context.Context, days int) ([]TrendData, error) {
	if days <= 0 {
		days = 7
	}
	if days > 365 {
		days = 365
	}

	end := time.Now()
	start := end.AddDate(0, 0, -days)

	statsList, err := s.repo.GetDailyStatsByRange(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("get trend stats: %w", err)
	}

	var trend []TrendData
	for _, stats := range statsList {
		trend = append(trend, TrendData{
			Date:      stats.StatDate.Format("01-02"),
			Visitors:  stats.UniqueVisitors,
			PageViews: stats.TotalPageViews,
			Sessions:  stats.TotalSessions,
		})
	}
	return trend, nil
}

func (s *Service) validateRequest(req *TrackVisitorRequest) error {
	if req.VisitorID == "" || req.SessionID == "" || req.Page == "" || req.Timestamp == 0 {
		return fmt.Errorf("missing required fields")
	}
	if len(req.VisitorID) > 64 || len(req.SessionID) > 64 || len(req.Page) > 255 {
		return fmt.Errorf("field too long")
	}
	if time.Since(time.UnixMilli(req.Timestamp)) > 24*time.Hour {
		return fmt.Errorf("timestamp too old")
	}
	return nil
}

func (s *Service) calculateDailyStats(ctx context.Context, date time.Time) (*DailyStats, error) {
	stats, err := s.repo.GetDailyStats(ctx, date)
	if err != nil {
		return nil, err
	}
	if stats != nil {
		return &DailyStats{
			Date:             date.Format("2006-01-02"),
			UniqueVisitors:   stats.UniqueVisitors,
			TotalSessions:    stats.TotalSessions,
			TotalPageViews:   stats.TotalPageViews,
			AvgSessionDuration: stats.AvgSessionDuration,
			BounceRate:       stats.BounceRate,
		}, nil
	}

	uv, _ := s.repo.CountUniqueVisitorsByDate(ctx, date)
	sessions, _ := s.repo.CountSessionsByDate(ctx, date)
	pv, _ := s.repo.CountPageViewsByDate(ctx, date)

	uvVal := 0
	if uv > 0 {
		uvVal = uv
	}
	sessionsVal := 0
	if sessions > 0 {
		sessionsVal = sessions
	}
	pvVal := 0
	if pv > 0 {
		pvVal = pv
	}
	ds := &DailyStats{
		Date:           date.Format("2006-01-02"),
		UniqueVisitors: uvVal,
		TotalSessions:  sessionsVal,
		TotalPageViews: pvVal,
	}
	return ds, nil
}
