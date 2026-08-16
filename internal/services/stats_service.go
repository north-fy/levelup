package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/north-fy/levelup/internal/domain"
	"github.com/north-fy/levelup/internal/pkg/cache"
)

// OverviewStats summarizes the user's account.
type OverviewStats struct {
	XP    int `json:"xp"`
	Gold  int `json:"gold"`
	Level int `json:"level"`
	Hours int `json:"hours"`
}

// BranchStat aggregates completed quests per branch.
type BranchStat struct {
	BranchID  uint `json:"branch_id"`
	Completed int  `json:"completed"`
	XP        int  `json:"xp"`
	Gold      int  `json:"gold"`
	Hours     int  `json:"hours"`
}

// RoadmapStat aggregates completed nodes per roadmap.
type RoadmapStat struct {
	RoadmapID uint `json:"roadmap_id"`
	Completed int  `json:"completed"`
	XP        int  `json:"xp"`
	Gold      int  `json:"gold"`
	Hours     int  `json:"hours"`
}

// QuestStat is a per-day aggregation of completed quests.
type QuestStat struct {
	Date      time.Time `json:"date"`
	Completed int       `json:"completed"`
	XP        int       `json:"xp"`
	Gold      int       `json:"gold"`
	Hours     int       `json:"hours"`
}

// StatsService reads statistics from ClickHouse.
type StatsService struct {
	db    *sql.DB
	users UserStore
	cache cache.Cache
}

// NewStatsService creates the statistics service.
func NewStatsService(db *sql.DB, users UserStore, c cache.Cache) *StatsService {
	return &StatsService{db: db, users: users, cache: c}
}

// Overview returns the user's current totals.
func (s *StatsService) Overview(ctx context.Context, userID uint) (*OverviewStats, error) {
	return getOrSet(ctx, s.cache, cache.OverviewKey(userID), cache.StatsTTL, func() (*OverviewStats, error) {
		user, err := s.users.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}

		var hours int
		if err := s.db.QueryRowContext(ctx,
			`SELECT coalesce(sum(hours), 0) FROM quest_completed WHERE user_id = ?`, userID,
		).Scan(&hours); err != nil {
			return nil, err
		}

		return &OverviewStats{
			XP:    user.XP,
			Gold:  user.Gold,
			Level: domain.LevelFor(user.XP),
			Hours: hours,
		}, nil
	})
}

// Branches returns per-branch aggregations.
func (s *StatsService) Branches(ctx context.Context, userID uint) ([]BranchStat, error) {
	rows, err := s.db.QueryContext(ctx, questBranchAggSQL, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var stats []BranchStat
	for rows.Next() {
		var stat BranchStat
		if err := rows.Scan(&stat.BranchID, &stat.Completed, &stat.XP, &stat.Gold, &stat.Hours); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

// Roadmaps returns per-roadmap aggregations.
func (s *StatsService) Roadmaps(ctx context.Context, userID uint) ([]RoadmapStat, error) {
	rows, err := s.db.QueryContext(ctx, questRoadmapAggSQL, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var stats []RoadmapStat
	for rows.Next() {
		var stat RoadmapStat
		if err := rows.Scan(&stat.RoadmapID, &stat.Completed, &stat.XP, &stat.Gold, &stat.Hours); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

// Quests returns daily aggregations within the given period.
func (s *StatsService) Quests(ctx context.Context, userID uint, period string) ([]QuestStat, error) {
	from := PeriodStart(period, time.Now())
	rows, err := s.db.QueryContext(ctx, questDailyAggSQL, userID, from)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var stats []QuestStat
	for rows.Next() {
		var stat QuestStat
		if err := rows.Scan(&stat.Date, &stat.Completed, &stat.XP, &stat.Gold, &stat.Hours); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, rows.Err()
}

// PeriodStart returns the start of the aggregation window for a period.
func PeriodStart(period string, now time.Time) time.Time {
	switch period {
	case "week":
		return now.AddDate(0, 0, -7)
	case "month":
		return now.AddDate(0, 0, -30)
	default:
		year, month, day := now.Date()
		return time.Date(year, month, day, 0, 0, 0, 0, now.Location())
	}
}

const (
	questBranchAggSQL = `
		SELECT branch_id, count(), sum(xp), sum(gold), sum(hours)
		FROM quest_completed
		WHERE user_id = ? AND branch_id > 0
		GROUP BY branch_id
		ORDER BY branch_id`

	questRoadmapAggSQL = `
		SELECT roadmap_id, count(), sum(xp), sum(gold), sum(hours)
		FROM quest_completed
		WHERE user_id = ? AND roadmap_id > 0
		GROUP BY roadmap_id
		ORDER BY roadmap_id`

	questDailyAggSQL = `
		SELECT toDate(completed_at) AS d, count(), sum(xp), sum(gold), sum(hours)
		FROM quest_completed
		WHERE user_id = ? AND completed_at >= ?
		GROUP BY d
		ORDER BY d`
)
