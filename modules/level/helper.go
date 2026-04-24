package level

import "math"

const (
	lvlMax = 100

	levelStepFactor = int64(50)

	voiceDeadXP   = 3.0 
	voiceActiveXP = 2.8 
	voiceBusyXP   = 2.5 

	textMinCharLen  = 3
	textMaxLen      = 100
	textMinXP       = 5.0
	textMaxAwardXP  = 20.0
	textDailyStart  = 100
	textDailyEnd    = 600
	textWeeklyStart = 1000
	textWeeklyEnd   = 4000

	dailyXPLimit = int64(800)

	textDeadThreshold   = 10
	textActiveThreshold = 50

	textDeadMultiplier       = 1.0  
	textActiveMultiplier     = 0.65 
	textVeryActiveMultiplier = 0.35 
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
	if contentLen < textMinCharLen {
		return 0
	}
	length := float64(min(contentLen, textMaxLen))
	base := textMinXP + (textMaxAwardXP-textMinXP)*(length/float64(textMaxLen))

	var serverMult float64
	switch {
	case serverHour < textDeadThreshold:
		serverMult = textDeadMultiplier
	case serverHour < textActiveThreshold:
		serverMult = textActiveMultiplier
	default:
		serverMult = textVeryActiveMultiplier
	}

	xp := base * serverMult * personalPenalty(userDaily, userWeekly)
	if math.IsNaN(xp) || math.IsInf(xp, 0) || xp < 0 {
		xp = 0
	}
	if xp > textMaxAwardXP {
		xp = textMaxAwardXP
	}
	if isLockdown(daytime) {
		return math.Min(xp, base*serverMult)
	}
	return xp
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

func voiceXPForCount(eligibleCount int) float64 {
	switch {
	case eligibleCount < 3:
		return voiceDeadXP
	case eligibleCount < 10:
		return voiceActiveXP
	default:
		return voiceBusyXP
	}
}

