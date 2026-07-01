package visitor

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"backend/internal/platform/database"
)

type Repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) *Repository {
	return &Repository{dbtx: dbtx}
}

func (r *Repository) UpsertSession(ctx context.Context, s *VisitorSession) error {
	query := `
		INSERT INTO visitor_sessions (visitor_id, session_id, first_visit_time, last_visit_time, page_count, platform, version, user_agent)
		VALUES ($1, $2, NOW(), NOW(), 1, $3, $4, $5)
		ON CONFLICT (session_id) DO UPDATE SET
			last_visit_time = NOW(),
			page_count = visitor_sessions.page_count + 1
	`
	_, err := r.dbtx.ExecContext(ctx, query,
		s.VisitorID, s.SessionID, s.Platform, s.Version, s.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("upsert session: %w", err)
	}
	return nil
}

func (r *Repository) BatchInsertPageViews(ctx context.Context, views []*VisitorPageView) error {
	if len(views) == 0 {
		return nil
	}
	query := `
		INSERT INTO visitor_page_views (visitor_id, session_id, page_path, page_title, visit_time, duration, referrer)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	for i := 0; i < len(views); i += 100 {
		end := i + 100
		if end > len(views) {
			end = len(views)
		}
		batch := views[i:end]
		_, err := r.dbtx.ExecContext(ctx, query,
			batch[0].VisitorID, batch[0].SessionID, batch[0].PagePath, batch[0].PageTitle,
			batch[0].VisitTime, batch[0].Duration, batch[0].Referrer,
		)
		if err != nil {
			return fmt.Errorf("batch insert page views: %w", err)
		}
	}
	return nil
}

func (r *Repository) GetDailyStats(ctx context.Context, date time.Time) (*VisitorDailyStats, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, stat_date, unique_visitors, total_sessions, total_page_views, avg_session_duration, bounce_rate, created_at, updated_at
		 FROM visitor_daily_stats WHERE stat_date = $1`, date.Format("2006-01-02"),
	)
	var s VisitorDailyStats
	var statDate time.Time
	err := row.Scan(&s.ID, &statDate, &s.UniqueVisitors, &s.TotalSessions, &s.TotalPageViews, &s.AvgSessionDuration, &s.BounceRate, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get daily stats: %w", err)
	}
	s.StatDate = statDate
	return &s, nil
}

func (r *Repository) GetDailyStatsByRange(ctx context.Context, start, end time.Time) ([]*VisitorDailyStats, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, stat_date, unique_visitors, total_sessions, total_page_views, avg_session_duration, bounce_rate, created_at, updated_at
		 FROM visitor_daily_stats WHERE stat_date BETWEEN $1 AND $2 ORDER BY stat_date ASC`,
		start.Format("2006-01-02"), end.Format("2006-01-02"),
	)
	if err != nil {
		return nil, fmt.Errorf("get daily stats by range: %w", err)
	}
	defer rows.Close()

	var stats []*VisitorDailyStats
	for rows.Next() {
		var s VisitorDailyStats
		var statDate time.Time
		if err := rows.Scan(&s.ID, &statDate, &s.UniqueVisitors, &s.TotalSessions, &s.TotalPageViews, &s.AvgSessionDuration, &s.BounceRate, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan daily stats: %w", err)
		}
		s.StatDate = statDate
		stats = append(stats, &s)
	}
	return stats, nil
}

func (r *Repository) InsertDailyStats(ctx context.Context, s *VisitorDailyStats) error {
	_, err := r.dbtx.ExecContext(ctx,
		`INSERT INTO visitor_daily_stats (stat_date, unique_visitors, total_sessions, total_page_views, avg_session_duration, bounce_rate)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (stat_date) DO UPDATE SET
			unique_visitors = EXCLUDED.unique_visitors,
			total_sessions = EXCLUDED.total_sessions,
			total_page_views = EXCLUDED.total_page_views,
			avg_session_duration = EXCLUDED.avg_session_duration,
			bounce_rate = EXCLUDED.bounce_rate,
			updated_at = NOW()`,
		s.StatDate.Format("2006-01-02"), s.UniqueVisitors, s.TotalSessions, s.TotalPageViews, s.AvgSessionDuration, s.BounceRate,
	)
	if err != nil {
		return fmt.Errorf("insert daily stats: %w", err)
	}
	return nil
}

func (r *Repository) CountUniqueVisitorsByDate(ctx context.Context, date time.Time) (int, error) {
	var count int
	err := r.dbtx.QueryRowContext(ctx,
		"SELECT COUNT(DISTINCT visitor_id) FROM visitor_page_views WHERE visit_time >= $1 AND visit_time < $1 + INTERVAL '1 day'",
		date.Format("2006-01-02"),
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count unique visitors: %w", err)
	}
	return count, nil
}

func (r *Repository) CountSessionsByDate(ctx context.Context, date time.Time) (int, error) {
	var count int
	err := r.dbtx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM visitor_sessions WHERE DATE(first_visit_time) = $1",
		date.Format("2006-01-02"),
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count sessions: %w", err)
	}
	return count, nil
}

func (r *Repository) CountPageViewsByDate(ctx context.Context, date time.Time) (int, error) {
	var count int
	err := r.dbtx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM visitor_page_views WHERE visit_time >= $1 AND visit_time < $1 + INTERVAL '1 day'",
		date.Format("2006-01-02"),
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count page views: %w", err)
	}
	return count, nil
}
