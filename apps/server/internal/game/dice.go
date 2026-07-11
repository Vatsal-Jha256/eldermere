package game

import "fmt"

const (
	rollNormal       = ""
	rollNormalName   = "normal"
	rollAdvantage    = "advantage"
	rollDisadvantage = "disadvantage"
)

type CheckResult struct {
	Rolls    []int
	Kept     int
	Modifier int
	Total    int
	DC       int
	Mode     string
	Success  bool
	Critical string
}

func resolveCheck(roller func(sides int) int, dc int, modifier int, mode string) CheckResult {
	result := CheckResult{
		Modifier: modifier,
		DC:       dc,
		Mode:     normalizeRollMode(mode),
	}
	if result.Mode == rollAdvantage || result.Mode == rollDisadvantage {
		first := roller(20)
		second := roller(20)
		result.Rolls = []int{first, second}
		if result.Mode == rollAdvantage {
			result.Kept = maxInt(first, second)
		} else {
			result.Kept = minInt(first, second)
		}
	} else {
		result.Kept = roller(20)
		result.Rolls = []int{result.Kept}
	}

	result.Total = result.Kept + result.Modifier
	switch result.Kept {
	case 20:
		result.Success = true
		result.Critical = "critical success"
	case 1:
		result.Success = false
		result.Critical = "critical failure"
	default:
		result.Success = result.Total >= result.DC
	}
	return result
}

func formatCheck(result CheckResult) string {
	mode := result.Mode
	if mode == rollNormal {
		mode = "normal"
	}
	return fmt.Sprintf("Rolled %s %s, kept %d %+d = %d against DC %d", formatRolls(result.Rolls), mode, result.Kept, result.Modifier, result.Total, result.DC)
}

func normalizeRollMode(mode string) string {
	switch mode {
	case rollNormal, rollNormalName:
		return rollNormal
	case rollAdvantage, rollDisadvantage:
		return mode
	default:
		return rollNormal
	}
}

func validRollMode(mode string) bool {
	switch mode {
	case rollNormal, rollNormalName, rollAdvantage, rollDisadvantage:
		return true
	default:
		return false
	}
}

func formatRolls(rolls []int) string {
	if len(rolls) == 0 {
		return "none"
	}
	if len(rolls) == 1 {
		return fmt.Sprintf("%d", rolls[0])
	}
	return fmt.Sprintf("%d/%d", rolls[0], rolls[1])
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
