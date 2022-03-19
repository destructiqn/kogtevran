package modules

import "strings"

var shortcuts = map[string]string{
	"flight":        ModuleFlight,
	"antiknockback": ModuleAntiKnockback,
	"nofall":        ModuleNoFall,
	"killaura":      ModuleKillAura,
	"spammer":       ModuleSpammer,
	"mobaura":       ModuleMobAura,
	"cheststealer":  ModuleChestStealer,
	"cmdcam":        ModuleCMDCam,
	"longjump":      ModuleLongJump,
	"unlimitedcps":  ModuleUnlimitedCPS,
	"tpaura":        ModuleTPAura,
	"playeresp":     ModulePlayerESP,
	"chestesp":      ModuleChestESP,
	"nuker":         ModuleNuker,
	"fastbreak":     ModuleFastBreak,
	"nobadeffects":  ModuleNoBadEffects,
}

func NormalizeModuleName(input string) string {
	fullName, ok := shortcuts[strings.ToLower(input)]
	if ok {
		return fullName
	}
	return input
}
