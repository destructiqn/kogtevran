package modules

const (
	CategoryCombat   = "Combat"
	CategoryMovement = "Movement"
	CategoryPlayer   = "Player"
	CategoryVisual   = "Visual"
	CategoryMisc     = "Miscellaneous"
)

var Categories = map[string][]string{
	CategoryCombat:   {ModuleAntiKnockback, ModuleKillAura, ModuleMobAura, ModuleTPAura},
	CategoryMovement: {ModuleFlight, ModuleLongJump, ModuleSpeedHack},
	CategoryPlayer:   {ModuleNoFall, ModuleChestStealer, ModuleNoBadEffects, ModuleUnlimitedCPS, ModuleAutoSoup},
	CategoryVisual:   {ModulePlayerESP, ModuleChestESP},
	CategoryMisc:     {ModuleSpammer, ModuleCMDCam},
}

type CategoryList []Category

func (c CategoryList) Len() int {
	return len([]Category(c))
}

func (c CategoryList) Less(i, j int) bool {
	return len(c[i].ModuleIDs) < len(c[j].ModuleIDs)
}

func (c CategoryList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

type Category struct {
	Name      string
	ModuleIDs []string
}

func GetCategoryList() CategoryList {
	list := make(CategoryList, 0)
	for id, modules := range Categories {
		list = append(list, Category{
			Name:      id,
			ModuleIDs: modules,
		})
	}
	return list
}
