package main

type MobType int

const (
	Creeper      MobType = 50
	Skeleton     MobType = 51
	Spider       MobType = 52
	GiantZombie  MobType = 53
	Zombie       MobType = 54
	Slime        MobType = 55
	Ghast        MobType = 56
	ZombiePigman MobType = 57
	Enderman     MobType = 58
	CaveSpider   MobType = 59
	Silverfish   MobType = 60
	Blaze        MobType = 61
	MagmaCube    MobType = 62
	EnderDragon  MobType = 63
	Wither       MobType = 64
	Bat          MobType = 65
	Witch        MobType = 66
	Endermite    MobType = 67
	Guardian     MobType = 68
	Shulker      MobType = 69
	Pig          MobType = 90
	Sheep        MobType = 91
	Cow          MobType = 92
	Chicken      MobType = 93
	Squid        MobType = 94
	Wolf         MobType = 95
	Mooshroom    MobType = 96
	Snowman      MobType = 97
	Ocelot       MobType = 98
	IronGolem    MobType = 99
	Horse        MobType = 100
	Rabbit       MobType = 101
	Villager     MobType = 120
)

type Entity interface {
	GetLocation() *Location
}

type DefaultEntity struct {
	Location *Location
}

func (e *DefaultEntity) GetLocation() *Location {
	return e.Location
}

type Mob struct {
	DefaultEntity
	Type MobType
}

type Player struct {
	DefaultEntity
}
