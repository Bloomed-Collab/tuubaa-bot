package level

import "math"

const (
	lvlMax = 100

	levelStepFactor = int64(50)

	voiceAwardXP = 5.0

	textBaseXP      = 20.0
	textPerCharXP   = 0.5
	textMaxLen      = 100
	textConstA      = 0.5
	textOffsetX     = 0.6
	textOffsetY     = 0.25
	textDailyStart  = 100
	textDailyEnd    = 600
	textWeeklyStart = 1000
	textWeeklyEnd   = 4000
	textMinAwardXP  = 1.0
	textMaxAwardXP  = 40.0
)

var lockdowns = [][2]int{{0, 6 * 60}, {8 * 60, 12 * 60}}

func calcLevel(xp int64) int {
	if xp < 0 {
		return 0
	}
	l := levelFromTotalXP(xp)
	if l < 0 {
		return 0
	}
	if l > lvlMax {
		return lvlMax
	}
	return l
}

func xpToNextLevel(xp int64) int64 {
	l := calcLevel(xp)
	if l >= lvlMax {
		return 0
	}
	return totalXPForLevel(l+1) - xp
}

func xpFromThisLevel(xp int64) int64 {
	if xp < 0 {
		return 0
	}
	l := calcLevel(xp)
	return xp - totalXPForLevel(l)
}

func isLockdown(daytime int) bool {
	for _, w := range lockdowns {
		if daytime >= w[0] && daytime < w[1] {
			return true
		}
	}
	return false
}

func evalTextXP(contentLen, daytime, serverHour, userDaily, userWeekly int) float64 {
	length := float64(min(contentLen, textMaxLen))
	base := textBaseXP + textPerCharXP*length
	afterServer := base * (math.Pow(math.E, float64(serverHour)/textConstA+textOffsetX) + textOffsetY)
	penalty := personalPenalty(userDaily, userWeekly)
	afterUser := afterServer * penalty
	if math.IsNaN(afterUser) || math.IsInf(afterUser, 0) {
		afterUser = base
	}
	if afterUser < textMinAwardXP {
		afterUser = textMinAwardXP
	}
	if afterUser > textMaxAwardXP {
		afterUser = textMaxAwardXP
	}
	if isLockdown(daytime) {
		return math.Min(afterUser, base)
	}
	return afterUser
}

func personalPenalty(daily, weekly int) float64 {
	var d, w float64
	switch {
	case daily < textDailyStart:
		d = 1.0
	case daily > textDailyEnd:
		d = 0.0
	default:
		d = float64(textDailyEnd-daily) / float64(textDailyEnd-textDailyStart)
	}
	switch {
	case weekly < textWeeklyStart:
		w = 1.0
	case weekly > textWeeklyEnd:
		w = 0.0
	default:
		w = float64(textWeeklyEnd-weekly) / float64(textWeeklyEnd-textWeeklyStart)
	}
	if d < w {
		return d
	}
	return w
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func totalXPForLevel(level int) int64 {
	if level <= 0 {
		return 0
	}
	if level > lvlMax {
		level = lvlMax
	}
	l := int64(level)
	return (levelStepFactor / 2) * l * (l + 1)
}

func levelFromTotalXP(xp int64) int {
	if xp <= 0 {
		return 0
	}
	v := float64(xp) / float64(levelStepFactor/2)
	return int(math.Floor((math.Sqrt(1+4*v) - 1) / 2))
}

func levelAccentColor(level int) int {
	switch {
	case level >= 100:
		return 0xe74c3c
	case level >= 80:
		return 0x9b59b6
	case level >= 60:
		return 0x3498db
	case level >= 40:
		return 0x2ecc71
	case level >= 20:
		return 0xf39c12
	default:
		return 0x95a5a6
	}
}

func buildProgressBar(current, total int64) string {
	const width = 20
	if total <= 0 {
		total = 1
	}
	filled := int(float64(current) / float64(total) * width)
	if filled > width {
		filled = width
	}
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}
