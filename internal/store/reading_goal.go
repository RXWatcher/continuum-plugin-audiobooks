package store

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ReadingGoal is one row in reading_goal — a user's target for a
// given year and kind. Progress is derived read-side; not stored.
type ReadingGoal struct {
	UserID    string    `json:"user_id"`
	Year      int       `json:"year"`
	Kind      string    `json:"kind"` // "books" | "hours"
	Target    int       `json:"target"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Store) UpsertReadingGoal(ctx context.Context, g ReadingGoal) error {
	if g.UserID == "" || g.Year < 2000 || g.Year > 2100 || g.Kind == "" || g.Target <= 0 {
		return errors.New("user_id, year, kind, target required")
	}
	if g.Kind != "books" && g.Kind != "hours" {
		return errors.New("kind must be 'books' or 'hours'")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO reading_goal (user_id, year, kind, target)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, year, kind) DO UPDATE SET
			target     = EXCLUDED.target,
			updated_at = now()
	`, g.UserID, g.Year, g.Kind, g.Target)
	if err != nil {
		return fmt.Errorf("upsert reading_goal: %w", err)
	}
	return nil
}

func (s *Store) ListReadingGoals(ctx context.Context, userID string, year int) ([]ReadingGoal, error) {
	if userID == "" {
		return nil, errors.New("user_id required")
	}
	q := `SELECT user_id, year, kind, target, created_at, updated_at
	      FROM reading_goal WHERE user_id = $1`
	args := []any{userID}
	if year > 0 {
		q += " AND year = $2"
		args = append(args, year)
	}
	q += " ORDER BY year DESC, kind"
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list reading_goal: %w", err)
	}
	defer rows.Close()
	var out []ReadingGoal
	for rows.Next() {
		var g ReadingGoal
		if err := rows.Scan(&g.UserID, &g.Year, &g.Kind, &g.Target,
			&g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan reading_goal: %w", err)
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *Store) DeleteReadingGoal(ctx context.Context, userID string, year int, kind string) error {
	if userID == "" || year < 2000 || kind == "" {
		return errors.New("user_id, year, kind required")
	}
	_, err := s.pool.Exec(ctx, `
		DELETE FROM reading_goal WHERE user_id = $1 AND year = $2 AND kind = $3
	`, userID, year, kind)
	if err != nil {
		return fmt.Errorf("delete reading_goal: %w", err)
	}
	return nil
}

// GoalProgress is the SPA-facing shape: target + actual + percent
// + whether we'll hit it at the current pace.
type GoalProgress struct {
	Year             int     `json:"year"`
	Kind             string  `json:"kind"`
	Target           int     `json:"target"`
	Actual           int     `json:"actual"`
	PercentComplete  float64 `json:"percent_complete"`
	OnPaceForTarget  bool    `json:"on_pace_for_target"`
	DaysIntoYear     int     `json:"days_into_year"`
	DaysInYear       int     `json:"days_in_year"`
}

// GoalProgressForUser computes progress for every goal the user
// has set in `year`. Books progress = count of progress rows
// where is_finished = true; hours progress = sum of
// reading_session.seconds_played / 3600. on_pace = actual /
// daysIntoYear >= target / daysInYear.
func (s *Store) GoalProgressForUser(ctx context.Context, userID, profileID string, year int, loc *time.Location) ([]GoalProgress, error) {
	goals, err := s.ListReadingGoals(ctx, userID, year)
	if err != nil {
		return nil, err
	}
	if len(goals) == 0 {
		return nil, nil
	}
	if loc == nil {
		loc = time.UTC
	}
	yearStart := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	yearEnd := yearStart.AddDate(1, 0, 0)
	daysInYear := int(yearEnd.Sub(yearStart).Hours() / 24)
	now := time.Now().In(loc)
	daysIntoYear := int(now.Sub(yearStart).Hours() / 24)
	if now.Before(yearStart) {
		daysIntoYear = 0
	}
	if now.After(yearEnd) {
		daysIntoYear = daysInYear
	}

	out := make([]GoalProgress, 0, len(goals))
	for _, g := range goals {
		gp := GoalProgress{
			Year:         g.Year,
			Kind:         g.Kind,
			Target:       g.Target,
			DaysIntoYear: daysIntoYear,
			DaysInYear:   daysInYear,
		}
		switch g.Kind {
		case "books":
			// Count distinct finished books whose progress was last
			// updated within `year`.
			var n int
			err := s.pool.QueryRow(ctx, `
				SELECT COUNT(DISTINCT book_id)::int FROM progress
				WHERE user_id = $1 AND profile_id = $2 AND is_finished = TRUE
				  AND updated_at >= $3 AND updated_at < $4
			`, userID, profileID, yearStart, yearEnd).Scan(&n)
			if err != nil {
				return nil, fmt.Errorf("books progress: %w", err)
			}
			gp.Actual = n
		case "hours":
			var secs int64
			err := s.pool.QueryRow(ctx, `
				SELECT COALESCE(SUM(seconds_played), 0)::bigint FROM reading_session
				WHERE user_id = $1 AND started_at >= $2 AND started_at < $3
			`, userID, yearStart, yearEnd).Scan(&secs)
			if err != nil {
				return nil, fmt.Errorf("hours progress: %w", err)
			}
			gp.Actual = int(secs / 3600)
		}
		if g.Target > 0 {
			gp.PercentComplete = float64(gp.Actual) / float64(g.Target) * 100
		}
		if daysIntoYear > 0 && daysInYear > 0 {
			expectedAtPace := float64(g.Target) * float64(daysIntoYear) / float64(daysInYear)
			gp.OnPaceForTarget = float64(gp.Actual) >= expectedAtPace
		} else {
			gp.OnPaceForTarget = true // start-of-year — assume on pace
		}
		out = append(out, gp)
	}
	return out, nil
}
