package store

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ReadingSession is one discrete listening session. SPA records on
// pause / close; aggregated read-side for the heatmap + year-in-
// review surfaces.
type ReadingSession struct {
	ID            string
	UserID        string
	BookID        string
	StartedAt     time.Time
	EndedAt       *time.Time
	SecondsPlayed int
	DeviceLabel   string
}

// InsertReadingSession records a completed session. Caller supplies
// the ULID id. EndedAt + SecondsPlayed required; open-ended
// sessions go through UpsertReadingSession with EndedAt nil and
// SecondsPlayed=0.
func (s *Store) InsertReadingSession(ctx context.Context, sess ReadingSession) error {
	if sess.ID == "" || sess.UserID == "" || sess.BookID == "" {
		return errors.New("id, user_id, book_id required")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO reading_session (id, user_id, book_id, started_at, ended_at, seconds_played, device_label)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7,''))
	`, sess.ID, sess.UserID, sess.BookID, sess.StartedAt, sess.EndedAt, sess.SecondsPlayed, sess.DeviceLabel)
	if err != nil {
		return fmt.Errorf("insert reading_session: %w", err)
	}
	return nil
}

// HeatmapDay is one (date, seconds_played) row used by the per-user
// heatmap. Date is the user's local day.
type HeatmapDay struct {
	Day            time.Time `json:"day"`
	SecondsPlayed  int       `json:"seconds_played"`
	SessionCount   int       `json:"session_count"`
}

// ListeningHeatmap returns one row per calendar day in the last
// `daysBack` days where the user had at least one session. Days
// with zero sessions are omitted — the SPA fills the gaps.
func (s *Store) ListeningHeatmap(ctx context.Context, userID string, daysBack int, loc *time.Location) ([]HeatmapDay, error) {
	if userID == "" {
		return nil, errors.New("user_id required")
	}
	if daysBack <= 0 || daysBack > 730 {
		daysBack = 365
	}
	if loc == nil {
		loc = time.UTC
	}
	rows, err := s.pool.Query(ctx, `
		SELECT date_trunc('day', started_at AT TIME ZONE $3)::date AS day,
		       COALESCE(SUM(seconds_played), 0)::int AS seconds_played,
		       COUNT(*)::int AS session_count
		FROM reading_session
		WHERE user_id = $1
		  AND started_at >= now() - ($2 || ' days')::interval
		GROUP BY day
		ORDER BY day
	`, userID, daysBack, loc.String())
	if err != nil {
		return nil, fmt.Errorf("heatmap query: %w", err)
	}
	defer rows.Close()
	var out []HeatmapDay
	for rows.Next() {
		var h HeatmapDay
		if err := rows.Scan(&h.Day, &h.SecondsPlayed, &h.SessionCount); err != nil {
			return nil, fmt.Errorf("scan heatmap day: %w", err)
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// YearStats aggregates one year of listening into the headline
// numbers a year-in-review screen renders.
type YearStats struct {
	Year             int           `json:"year"`
	TotalSeconds     int64         `json:"total_seconds"`
	SessionCount     int           `json:"session_count"`
	DistinctBooks    int           `json:"distinct_books"`
	DistinctDays     int           `json:"distinct_days"`
	LongestSession   int           `json:"longest_session_seconds"`
	TopBooks         []YearTopBook `json:"top_books"`
}

// YearTopBook is one entry in the "books I spent the most time on
// this year" list.
type YearTopBook struct {
	BookID        string `json:"book_id"`
	SecondsPlayed int64  `json:"seconds_played"`
	SessionCount  int    `json:"session_count"`
}

// ListeningStatsForYear computes the headline aggregates + top-N
// books for one calendar year (in the supplied timezone). Year is
// inclusive on both ends (Jan 1 → Dec 31).
func (s *Store) ListeningStatsForYear(ctx context.Context, userID string, year int, loc *time.Location) (YearStats, error) {
	if userID == "" {
		return YearStats{}, errors.New("user_id required")
	}
	if loc == nil {
		loc = time.UTC
	}
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	end := start.AddDate(1, 0, 0)
	row := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(seconds_played), 0)::bigint AS total_seconds,
		       COUNT(*)::int AS session_count,
		       COUNT(DISTINCT book_id)::int AS distinct_books,
		       COUNT(DISTINCT date_trunc('day', started_at AT TIME ZONE $4))::int AS distinct_days,
		       COALESCE(MAX(seconds_played), 0)::int AS longest
		FROM reading_session
		WHERE user_id = $1 AND started_at >= $2 AND started_at < $3
	`, userID, start, end, loc.String())
	out := YearStats{Year: year}
	if err := row.Scan(&out.TotalSeconds, &out.SessionCount, &out.DistinctBooks,
		&out.DistinctDays, &out.LongestSession); err != nil {
		return YearStats{}, fmt.Errorf("year stats: %w", err)
	}
	// Top books — 5 entries, ordered by total seconds desc.
	bookRows, err := s.pool.Query(ctx, `
		SELECT book_id, SUM(seconds_played)::bigint, COUNT(*)::int
		FROM reading_session
		WHERE user_id = $1 AND started_at >= $2 AND started_at < $3
		GROUP BY book_id
		ORDER BY SUM(seconds_played) DESC
		LIMIT 5
	`, userID, start, end)
	if err != nil {
		return out, fmt.Errorf("top books: %w", err)
	}
	defer bookRows.Close()
	for bookRows.Next() {
		var b YearTopBook
		if err := bookRows.Scan(&b.BookID, &b.SecondsPlayed, &b.SessionCount); err != nil {
			return out, fmt.Errorf("scan top book: %w", err)
		}
		out.TopBooks = append(out.TopBooks, b)
	}
	return out, bookRows.Err()
}
