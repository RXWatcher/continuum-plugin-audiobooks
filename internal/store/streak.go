package store

import (
	"context"
	"fmt"
	"time"
)

// Streak is the result of StreakForUser — current + longest run of
// consecutive days with at least one progress update, plus the date
// of the most recent active day. Days are computed in the user's
// timezone — the caller passes a *time.Location; UTC is a fine
// default when no per-user setting exists.
type Streak struct {
	Current        int       `json:"current"`
	Longest        int       `json:"longest"`
	LastActiveDate time.Time `json:"last_active_date"`
}

// StreakForUser computes the user's listening streak from
// progress.updated_at. A "day" is any calendar date in loc where at
// least one progress row was updated. Current streak counts
// backwards from today (or yesterday — we tolerate a 1-day grace
// so users don't lose their streak by going to bed early), longest
// is the max consecutive run ever recorded.
//
// We pull distinct dates rather than a row-per-progress aggregate
// because some users have thousands of progress updates per day
// during binge sessions, and the dedup happens cheaper in SQL.
func (s *Store) StreakForUser(ctx context.Context, userID string, loc *time.Location) (Streak, error) {
	if userID == "" {
		return Streak{}, fmt.Errorf("user_id required")
	}
	if loc == nil {
		loc = time.UTC
	}
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT date_trunc('day', updated_at AT TIME ZONE $2)::date AS day
		FROM progress
		WHERE user_id = $1
		ORDER BY day DESC
		LIMIT 365
	`, userID, loc.String())
	if err != nil {
		return Streak{}, fmt.Errorf("streak query: %w", err)
	}
	defer rows.Close()
	var days []time.Time
	for rows.Next() {
		var d time.Time
		if err := rows.Scan(&d); err != nil {
			return Streak{}, fmt.Errorf("scan: %w", err)
		}
		days = append(days, d)
	}
	if err := rows.Err(); err != nil {
		return Streak{}, fmt.Errorf("streak rows: %w", err)
	}
	if len(days) == 0 {
		return Streak{}, nil
	}

	// Current streak: walk backwards from the most-recent day; each
	// adjacent day must be exactly 1 day earlier than the previous.
	// First-day tolerance: if the most recent active day is today or
	// yesterday, we count it; if it's older, current is 0.
	today := time.Now().In(loc).Format("2006-01-02")
	yesterday := time.Now().In(loc).AddDate(0, 0, -1).Format("2006-01-02")
	mostRecent := days[0].Format("2006-01-02")

	current := 0
	if mostRecent == today || mostRecent == yesterday {
		current = 1
		for i := 1; i < len(days); i++ {
			diff := days[i-1].Sub(days[i]).Hours() / 24
			if diff > 1.5 { // 1.5 to tolerate DST shifts
				break
			}
			current++
		}
	}

	// Longest streak: walk every consecutive run.
	longest := 0
	run := 0
	for i := range days {
		if i == 0 {
			run = 1
			longest = 1
			continue
		}
		diff := days[i-1].Sub(days[i]).Hours() / 24
		if diff <= 1.5 {
			run++
		} else {
			run = 1
		}
		if run > longest {
			longest = run
		}
	}

	return Streak{
		Current:        current,
		Longest:        longest,
		LastActiveDate: days[0],
	}, nil
}
