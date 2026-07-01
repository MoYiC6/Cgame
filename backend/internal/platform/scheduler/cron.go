package scheduler

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CronSchedule 解析标准 5/6 字段 cron 表达式（秒/分/时/日/月/星期），最小实现。
type CronSchedule struct {
	hasSeconds bool
	seconds    []int
	minutes    []int
	hours      []int
	days       []int
	months     []int
	dows       []int // day of week
}

// ParseCron 支持字段：*、*/n、a-b、a,b,c。支持 5 字段（分时日月星期）或 6 字段（秒分时日月星期）。
func ParseCron(expr string) (*CronSchedule, error) {
	parts := strings.Fields(expr)
	if len(parts) != 5 && len(parts) != 6 {
		return nil, fmt.Errorf("cron expression must have 5 or 6 fields, got %d", len(parts))
	}

	cs := &CronSchedule{}
	var err error

	if len(parts) == 6 {
		cs.hasSeconds = true
		cs.seconds, err = parseField(parts[0], 0, 59)
		if err != nil {
			return nil, fmt.Errorf("second: %w", err)
		}
		parts = parts[1:]
	}

	cs.minutes, err = parseField(parts[0], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("minute: %w", err)
	}
	cs.hours, err = parseField(parts[1], 0, 23)
	if err != nil {
		return nil, fmt.Errorf("hour: %w", err)
	}
	cs.days, err = parseField(parts[2], 1, 31)
	if err != nil {
		return nil, fmt.Errorf("day: %w", err)
	}
	cs.months, err = parseField(parts[3], 1, 12)
	if err != nil {
		return nil, fmt.Errorf("month: %w", err)
	}
	cs.dows, err = parseField(parts[4], 0, 6)
	if err != nil {
		return nil, fmt.Errorf("dow: %w", err)
	}

	sort.Ints(cs.minutes)
	sort.Ints(cs.hours)
	sort.Ints(cs.days)
	sort.Ints(cs.months)
	sort.Ints(cs.dows)

	return cs, nil
}

// Next 计算从 now 开始的下一个触发时间。
func (c *CronSchedule) Next(now time.Time) time.Time {
	start := now
	step := time.Minute
	if c.hasSeconds {
		step = time.Second
	}
	for {
		now = now.Add(step)
		if c.hasSeconds && !matches(c.seconds, now.Second()) {
			continue
		}
		if !matches(c.months, int(now.Month())) {
			continue
		}
		if !matches(c.days, now.Day()) {
			continue
		}
		if !matches(c.dows, int(now.Weekday())) {
			continue
		}
		if !matches(c.hours, now.Hour()) {
			continue
		}
		if !matches(c.minutes, now.Minute()) {
			continue
		}
		if now.Before(start) || now.Equal(start) {
			return now.Add(step)
		}
		return now
	}
}

func parseField(field string, min, max int) ([]int, error) {
	values := make(map[int]struct{})
	parts := strings.Split(field, ",")

	for _, part := range parts {
		if strings.Contains(part, "/") {
			rangeParts := strings.SplitN(part, "/", 2)
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range: %s", part)
			}
			step, err := strconv.Atoi(rangeParts[1])
			if err != nil || step <= 0 {
				return nil, fmt.Errorf("invalid step: %s", rangeParts[1])
			}
			var rng []int
			if rangeParts[0] == "*" {
				rng = []int{min, max}
			} else {
				bounds, err := parseRange(rangeParts[0], min, max)
				if err != nil {
					return nil, err
				}
				rng = bounds
			}
			for i := rng[0]; i <= rng[1]; i += step {
				values[i] = struct{}{}
			}
			continue
		}

		if part == "*" {
			for i := min; i <= max; i++ {
				values[i] = struct{}{}
			}
			continue
		}

		bounds, err := parseRange(part, min, max)
		if err != nil {
			return nil, err
		}
		for i := bounds[0]; i <= bounds[1]; i++ {
			values[i] = struct{}{}
		}
	}

	if len(values) == 0 {
		return nil, errors.New("empty field")
	}

	result := make([]int, 0, len(values))
	for v := range values {
		result = append(result, v)
	}
	return result, nil
}

func parseRange(s string, min, max int) ([]int, error) {
	if !strings.Contains(s, "-") {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", s)
		}
		if v < min || v > max {
			return nil, fmt.Errorf("value %d out of range [%d-%d]", v, min, max)
		}
		return []int{v, v}, nil
	}

	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range: %s", s)
	}

	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid range start: %s", parts[0])
	}
	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid range end: %s", parts[1])
	}

	if start < min || end > max || start > end {
		return nil, fmt.Errorf("range %d-%d out of bounds [%d-%d]", start, end, min, max)
	}

	return []int{start, end}, nil
}

func matches(values []int, v int) bool {
	for _, val := range values {
		if val == v {
			return true
		}
	}
	return false
}
