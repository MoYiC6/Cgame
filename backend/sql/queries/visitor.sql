-- name: UpsertSession :one
INSERT INTO visitor_sessions (visitor_id, session_id, first_visit_time, last_visit_time, page_count, platform, version, user_agent)
VALUES (@visitorId::text, @sessionId::text, NOW(), NOW(), 1, @platform::text, @version::text, @userAgent::text)
ON CONFLICT (session_id) DO UPDATE SET
    last_visit_time = NOW(),
    page_count = visitor_sessions.page_count + 1;

-- name: BatchInsertPageViews :exec
INSERT INTO visitor_page_views (visitor_id, session_id, page_path, page_title, visit_time, duration, referrer)
VALUES
    @p1::(visitor_id text, session_id text, page_path text, page_title text, visit_time timestamptz, duration integer, referrer text),
    @p2::(visitor_id text, session_id text, page_path text, page_title text, visit_time timestamptz, duration integer, referrer text);

-- name: GetDailyStats :one
SELECT id, stat_date, unique_visitors, total_sessions, total_page_views, avg_session_duration, bounce_rate, created_at, updated_at
FROM visitor_daily_stats
WHERE stat_date = @date::date;

-- name: GetDailyStatsByRange :many
SELECT id, stat_date, unique_visitors, total_sessions, total_page_views, avg_session_duration, bounce_rate, created_at, updated_at
FROM visitor_daily_stats
WHERE stat_date BETWEEN @startDate::date AND @endDate::date
ORDER BY stat_date ASC;

-- name: InsertDailyStats :one
INSERT INTO visitor_daily_stats (stat_date, unique_visitors, total_sessions, total_page_views, avg_session_duration, bounce_rate)
VALUES (@statDate::date, @uniqueVisitors, @totalSessions, @totalPageViews, @avgSessionDuration, @bounceRate)
ON CONFLICT (stat_date) DO UPDATE SET
    unique_visitors = EXCLUDED.unique_visitors,
    total_sessions = EXCLUDED.total_sessions,
    total_page_views = EXCLUDED.total_page_views,
    avg_session_duration = EXCLUDED.avg_session_duration,
    bounce_rate = EXCLUDED.bounce_rate,
    updated_at = NOW();

-- name: CountUniqueVisitorsByDate :one
SELECT COUNT(DISTINCT visitor_id) FROM visitor_page_views
WHERE visit_time >= @date::date AND visit_time < @date::date + INTERVAL '1 day';

-- name: CountSessionsByDate :one
SELECT COUNT(*) FROM visitor_sessions
WHERE DATE(first_visit_time) = @date::date;

-- name: CountPageViewsByDate :one
SELECT COUNT(*) FROM visitor_page_views
WHERE visit_time >= @date::date AND visit_time < @date::date + INTERVAL '1 day';
