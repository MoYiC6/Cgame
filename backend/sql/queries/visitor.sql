-- name: UpsertSession :exec
INSERT INTO visitor_sessions (visitor_id, session_id, first_visit_time, last_visit_time, page_count, platform, version, user_agent)
VALUES ($1, $2, NOW(), NOW(), 1, $3, $4, $5)
ON CONFLICT (session_id) DO UPDATE SET
    last_visit_time = NOW(),
    page_count = visitor_sessions.page_count + 1;

-- name: BatchInsertPageViews :exec
INSERT INTO visitor_page_views (visitor_id, session_id, page_path, page_title, visit_time, duration, referrer)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetDailyStats :one
SELECT id, stat_date, unique_visitors, total_sessions, total_page_views, avg_session_duration, bounce_rate, created_at, updated_at
FROM visitor_daily_stats
WHERE stat_date = $1;

-- name: GetDailyStatsByRange :many
SELECT id, stat_date, unique_visitors, total_sessions, total_page_views, avg_session_duration, bounce_rate, created_at, updated_at
FROM visitor_daily_stats
WHERE stat_date BETWEEN $1 AND $2
ORDER BY stat_date ASC;

-- name: InsertDailyStats :exec
INSERT INTO visitor_daily_stats (stat_date, unique_visitors, total_sessions, total_page_views, avg_session_duration, bounce_rate)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (stat_date) DO UPDATE SET
    unique_visitors = EXCLUDED.unique_visitors,
    total_sessions = EXCLUDED.total_sessions,
    total_page_views = EXCLUDED.total_page_views,
    avg_session_duration = EXCLUDED.avg_session_duration,
    bounce_rate = EXCLUDED.bounce_rate,
    updated_at = NOW();

-- name: CountUniqueVisitorsByDate :one
SELECT COUNT(DISTINCT visitor_id) FROM visitor_page_views
WHERE visit_time >= $1 AND visit_time < $1 + INTERVAL '1 day';

-- name: CountSessionsByDate :one
SELECT COUNT(*) FROM visitor_sessions
WHERE DATE(first_visit_time) = $1;

-- name: CountPageViewsByDate :one
SELECT COUNT(*) FROM visitor_page_views
WHERE visit_time >= $1 AND visit_time < $1 + INTERVAL '1 day';
