package game

import (
	"strings"
	"testing"
)

func TestResolveCheckUsesAdvantageAndCriticalSuccess(t *testing.T) {
	rolls := sequenceRoller(4, 20)

	result := resolveCheck(rolls, 30, 0, rollAdvantage)

	if !result.Success || result.Critical != "critical success" {
		t.Fatalf("expected natural 20 to critically succeed, got %#v", result)
	}
	if result.Kept != 20 || result.Total != 20 {
		t.Fatalf("expected advantage to keep 20, got %#v", result)
	}
}

func TestResolveCheckUsesDisadvantageAndCriticalFailure(t *testing.T) {
	rolls := sequenceRoller(18, 1)

	result := resolveCheck(rolls, 2, 20, rollDisadvantage)

	if result.Success || result.Critical != "critical failure" {
		t.Fatalf("expected natural 1 to critically fail, got %#v", result)
	}
	if result.Kept != 1 || result.Total != 21 {
		t.Fatalf("expected disadvantage to keep 1 with modifier applied, got %#v", result)
	}
}

func TestFormatCheckShowsRollModeAndMath(t *testing.T) {
	result := CheckResult{
		Rolls:    []int{7, 15},
		Kept:     15,
		Modifier: 3,
		Total:    18,
		DC:       14,
		Mode:     rollAdvantage,
		Success:  true,
	}

	text := formatCheck(result)
	if !strings.Contains(text, "7/15 advantage") || !strings.Contains(text, "15 +3 = 18") {
		t.Fatalf("expected readable check math, got %q", text)
	}

	result.Critical = "critical success"
	text = formatCheck(result)
	if !strings.Contains(text, "critical success") {
		t.Fatalf("expected critical result in check text, got %q", text)
	}
}

func TestSuccessChanceAccountsForCriticalsAndRollMode(t *testing.T) {
	normal := successChance(14, 2, rollNormal)
	if normal != 0.45 {
		t.Fatalf("expected normal chance 45%%, got %v", normal)
	}

	advantage := successChance(14, 2, rollAdvantage)
	if formatChance(advantage) != "69.8%" {
		t.Fatalf("expected advantage chance 69.8%%, got %s", formatChance(advantage))
	}

	disadvantage := successChance(14, 2, rollDisadvantage)
	if formatChance(disadvantage) != "20.2%" {
		t.Fatalf("expected disadvantage chance 20.2%%, got %s", formatChance(disadvantage))
	}
}

func sequenceRoller(values ...int) func(sides int) int {
	index := 0
	return func(sides int) int {
		if index >= len(values) {
			return values[len(values)-1]
		}
		value := values[index]
		index++
		return value
	}
}
