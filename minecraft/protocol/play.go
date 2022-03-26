package protocol

import (
	"github.com/Tnze/go-mc/chat"
	pk "github.com/destructiqn/kogtevran/minecraft/net/packet"
)

const (
	ClientboundKeepAlive = iota
	ClientboundJoinGame
	ClientboundChatMessage
	ClientboundTimeUpdate
	ClientboundEntityEquipment
	ClientboundSpawnPosition
	ClientboundUpdateHealth
	ClientboundRespawn
	ClientboundPlayerPositionAndLook
	ClientboundHeldItemChange
	ClientboundUseBed
	ClientboundAnimation
	ClientboundSpawnPlayer
	ClientboundCollectItem
	ClientboundSpawnObject
	ClientboundSpawnMob
	ClientboundSpawnPainting
	ClientboundSpawnExperienceOrb
	ClientboundEntityVelocity
	ClientboundDestroyEntities
	ClientboundEntity
	ClientboundEntityRelativeMove
	ClientboundEntityLook
	ClientboundEntityLookAndRelativeMove
	ClientboundEntityTeleport
	ClientboundEntityHeadLook
	ClientboundEntityStatus
	ClientboundAttachEntity
	ClientboundEntityMetadata
	ClientboundEntityEffect
	ClientboundRemoveEntityEffect
	ClientboundSetExperience
	ClientboundEntityProperties
	ClientboundChunkData
	ClientboundMultiBlockChange
	ClientboundBlockChange
	ClientboundBlockAction
	ClientboundBlockBreakAnimation
	ClientboundMapChunkBulk
	ClientboundExplosion
	ClientboundEffect
	ClientboundSoundEffect
	ClientboundParticle
	ClientboundChangeGameState
	ClientboundSpawnGlobalEntity
	ClientboundOpenWindow
	ClientboundCloseWindow
	ClientboundSetSlot
	ClientboundWindowItems
	ClientboundWindowProperty
	ClientboundConfirmTransaction
	ClientboundUpdateSign
	ClientboundMap
	ClientboundUpdateBlockEntity
	ClientboundOpenSignEditor
	ClientboundStatistics
	ClientboundPlayerListItem
	ClientboundPlayerAbilities
	ClientboundTabComplete
	ClientboundScoreboardObjective
	ClientboundUpdateScore
	ClientboundDisplayScoreboard
	ClientboundTeams
	ClientboundPluginMessage
	ClientboundDisconnect
	ClientboundServerDifficulty
	ClientboundCombatEvent
	ClientboundCamera
	ClientboundWorldBorder
	ClientboundTitle
	ClientboundSetCompression
	ClientboundPlayerListHeaderAndFooter
	ClientboundResourcePackSend
	ClientboundUpdateEntityNBT
)

type JoinGame struct {
	EntityID         pk.Int
	GameMode         pk.UnsignedByte
	Dimension        pk.Byte
	Difficulty       pk.UnsignedByte
	MaxPlayers       pk.UnsignedByte
	LevelType        pk.String
	ReducedDebugInfo pk.Boolean
}

func (j *JoinGame) Read(packet pk.Packet) error {
	return packet.Scan(&j.EntityID, &j.GameMode, &j.Dimension, &j.Difficulty, &j.MaxPlayers, &j.LevelType, &j.ReducedDebugInfo)
}

func (j *JoinGame) Marshal() pk.Packet {
	return pk.Marshal(ClientboundJoinGame, j.EntityID, j.GameMode, j.Dimension, j.Difficulty, j.MaxPlayers, j.LevelType, j.ReducedDebugInfo)
}

type UpdateHealth struct {
	Health         pk.Float
	Food           pk.VarInt
	FoodSaturation pk.Float
}

func (u *UpdateHealth) Read(packet pk.Packet) error {
	return packet.Scan(&u.Health, &u.Food, &u.FoodSaturation)
}

func (u *UpdateHealth) Marshal() pk.Packet {
	return pk.Marshal(ClientboundUpdateHealth, u.Health, u.Food, u.FoodSaturation)
}

type PlayerPositionAndLook struct {
	X, Y, Z    pk.Double
	Yaw, Pitch pk.Float
	Flags      pk.Byte
}

func (p *PlayerPositionAndLook) Read(packet pk.Packet) error {
	return packet.Scan(&p.X, &p.Y, &p.Z, &p.Yaw, &p.Pitch, &p.Flags)
}

func (p *PlayerPositionAndLook) Marshal() pk.Packet {
	return pk.Marshal(ClientboundPlayerPositionAndLook, p.X, p.Y, p.Z, p.Yaw, p.Pitch, p.Flags)
}

type HeldItemChange struct {
	Slot pk.Byte
}

func (h *HeldItemChange) Read(packet pk.Packet) error {
	return packet.Scan(&h.Slot)
}

func (h *HeldItemChange) Marshal() pk.Packet {
	return pk.Marshal(ClientboundHeldItemChange, h.Slot)
}

type SpawnPlayer struct {
	EntityID    pk.VarInt
	PlayerUUID  pk.UUID
	X, Y, Z     pk.Int
	Yaw, Pitch  pk.Angle
	CurrentItem pk.Short
	// TODO: Metadata
}

func (s *SpawnPlayer) Read(packet pk.Packet) error {
	return packet.Scan(&s.EntityID, &s.PlayerUUID, &s.X, &s.Y, &s.Z, &s.Yaw, &s.Pitch, &s.CurrentItem)
}

func (s *SpawnPlayer) Marshal() pk.Packet {
	return pk.Marshal(ClientboundSpawnPlayer, s.EntityID, s.PlayerUUID, s.X, s.Y, s.Z, s.Yaw, s.Pitch, s.CurrentItem)
}

type SpawnMob struct {
	EntityID   pk.VarInt
	Type       pk.UnsignedByte
	X, Y, Z    pk.Int
	Yaw, Pitch pk.Angle
	HeadPitch  pk.Angle
	VX, VY, VZ pk.Short
	// TODO: Metadata
}

func (s *SpawnMob) Read(packet pk.Packet) error {
	return packet.Scan(&s.EntityID, &s.Type, &s.X, &s.Y, &s.Z, &s.Yaw, &s.Pitch, &s.HeadPitch, &s.VX, &s.VY, &s.VZ)
}

func (s *SpawnMob) Marshal() pk.Packet {
	return pk.Marshal(ClientboundSpawnMob, s.EntityID, s.Type, s.X, s.Y, s.Z, s.Yaw, s.Pitch, s.HeadPitch, s.VX, s.VY, s.VZ)
}

type EntityVelocity struct {
	EntityID   pk.VarInt
	VX, VY, VZ pk.Short
}

func (e *EntityVelocity) Read(packet pk.Packet) error {
	return packet.Scan(&e.EntityID, &e.VX, &e.VY, &e.VZ)
}

func (e *EntityVelocity) Marshal() pk.Packet {
	return pk.Marshal(ClientboundEntityVelocity, e.EntityID, e.VX, e.VY, e.VZ)
}

type DestroyEntities struct {
	EntityIDs []pk.VarInt
}

func (d *DestroyEntities) Read(packet pk.Packet) error {
	var count pk.VarInt
	return packet.Scan(&count, &pk.Ary{
		Len: &count,
		Ary: &d.EntityIDs,
	})
}

func (d *DestroyEntities) Marshal() pk.Packet {
	return pk.Marshal(ClientboundDestroyEntities, pk.VarInt(len(d.EntityIDs)), pk.Ary{
		Ary: d.EntityIDs,
	})
}

type EntityRelativeMove struct {
	EntityID   pk.VarInt
	DX, DY, DZ pk.Byte
	OnGround   pk.Boolean
}

func (e *EntityRelativeMove) Read(packet pk.Packet) error {
	return packet.Scan(&e.EntityID, &e.DX, &e.DY, &e.DZ, &e.OnGround)
}

func (e *EntityRelativeMove) Marshal() pk.Packet {
	return pk.Marshal(ClientboundEntityRelativeMove, e.EntityID, e.DX, e.DY, e.DZ, e.OnGround)
}

type EntityLookAndRelativeMove struct {
	EntityID   pk.VarInt
	DX, DY, DZ pk.Byte
	Yaw, Pitch pk.Angle
	OnGround   pk.Boolean
}

func (e *EntityLookAndRelativeMove) Read(packet pk.Packet) error {
	return packet.Scan(&e.EntityID, &e.DX, &e.DY, &e.DZ, &e.Yaw, &e.Pitch, &e.OnGround)
}

func (e *EntityLookAndRelativeMove) Marshal() pk.Packet {
	return pk.Marshal(ClientboundEntityLookAndRelativeMove, e.EntityID, e.DX, e.DY, e.DZ, e.Yaw, e.Pitch, e.OnGround)
}

type EntityTeleport struct {
	EntityID   pk.VarInt
	X, Y, Z    pk.Int
	Yaw, Pitch pk.Angle
	OnGround   pk.Boolean
}

func (e *EntityTeleport) Read(packet pk.Packet) error {
	return packet.Scan(&e.EntityID, &e.X, &e.Y, &e.Z, &e.Yaw, &e.Pitch, &e.OnGround)
}

func (e *EntityTeleport) Marshal() pk.Packet {
	return pk.Marshal(ClientboundEntityTeleport, e.EntityID, e.X, e.Y, e.Z, e.Yaw, e.Pitch, e.OnGround)
}

type EntityMetadata struct {
	EntityID pk.VarInt
	Metadata pk.EntityMetadata
}

func (e *EntityMetadata) Read(packet pk.Packet) error {
	e.Metadata = make(pk.EntityMetadata)
	return packet.Scan(&e.EntityID, &e.Metadata)
}

func (e *EntityMetadata) Marshal() pk.Packet {
	return pk.Marshal(ClientboundEntityMetadata, e.EntityID, e.Metadata)
}

type EntityEffect struct {
	EntityID      pk.VarInt
	EffectID      pk.Byte
	Amplifier     pk.Byte
	Duration      pk.VarInt
	HideParticles pk.Boolean
}

func (e *EntityEffect) Read(packet pk.Packet) error {
	return packet.Scan(&e.EntityID, &e.EffectID, &e.Amplifier, &e.Duration, &e.HideParticles)
}

func (e *EntityEffect) Marshal() pk.Packet {
	return pk.Marshal(ClientboundEntityEffect, e.EntityID, e.EffectID, e.Amplifier, e.Duration, e.HideParticles)
}

type RemoveEntityEffect struct {
	EntityID pk.VarInt
	EffectID pk.Byte
}

func (r *RemoveEntityEffect) Read(packet pk.Packet) error {
	return packet.Scan(&r.EntityID, &r.EffectID)
}

func (r *RemoveEntityEffect) Marshal() pk.Packet {
	return pk.Marshal(ClientboundRemoveEntityEffect, r.EntityID, r.EffectID)
}

type EntityProperties struct {
	EntityID   pk.VarInt
	Properties []pk.Property
}

func (e *EntityProperties) Read(packet pk.Packet) error {
	var propertiesLen pk.Int
	return packet.Scan(&e.EntityID, &propertiesLen, &pk.Ary{
		Len: propertiesLen,
		Ary: &e.Properties,
	})
}

func (e *EntityProperties) Marshal() pk.Packet {
	return pk.Marshal(ClientboundEntityProperties, e.EntityID, pk.Int(len(e.Properties)), pk.Ary{
		Len: len(e.Properties),
		Ary: e.Properties,
	})
}

type BlockChange struct {
	Location pk.Position
	BlockID  pk.VarInt
}

func (b *BlockChange) Read(packet pk.Packet) error {
	return packet.Scan(&b.Location, &b.BlockID)
}

func (b *BlockChange) Marshal() pk.Packet {
	return pk.Marshal(ClientboundBlockChange, b.Location, b.BlockID)
}

type ChangeGameState struct {
	Reason pk.UnsignedByte
	Value  pk.Float
}

func (c *ChangeGameState) Read(packet pk.Packet) error {
	return packet.Scan(&c.Reason, &c.Value)
}

func (c *ChangeGameState) Marshal() pk.Packet {
	return pk.Marshal(ClientboundChangeGameState, c.Reason, c.Value)
}

type OpenWindow struct {
	WindowID      pk.UnsignedByte
	WindowType    pk.String
	WindowTitle   chat.Message
	NumberOfSlots pk.UnsignedByte
	EntityID      pk.Int
}

func (o *OpenWindow) Read(packet pk.Packet) error {
	return packet.Scan(&o.WindowID, &o.WindowType, &o.WindowTitle, &o.NumberOfSlots, &pk.Opt{
		Has:   o.WindowType == "EntityHorse",
		Field: &o.EntityID,
	})
}

func (o *OpenWindow) Marshal() pk.Packet {
	return pk.Marshal(ClientboundOpenWindow, o.WindowID, o.WindowType, o.WindowTitle, o.NumberOfSlots, pk.Opt{
		Has:   o.WindowType == "EntityHorse",
		Field: o.EntityID,
	})
}

type CloseWindow struct {
	WindowID pk.UnsignedByte
}

func (c *CloseWindow) Read(packet pk.Packet) error {
	return packet.Scan(&c.WindowID)
}

func (c *CloseWindow) Marshal() pk.Packet {
	return pk.Marshal(ClientboundCloseWindow, c.WindowID)
}

type SetSlot struct {
	WindowID pk.Byte
	Slot     pk.Short
	SlotData pk.Slot
}

func (s *SetSlot) Read(packet pk.Packet) error {
	return packet.Scan(&s.WindowID, &s.Slot, &s.SlotData)
}

func (s *SetSlot) Marshal() pk.Packet {
	return pk.Marshal(ClientboundSetSlot, s.WindowID, s.Slot, s.SlotData)
}

type WindowItems struct {
	WindowID pk.UnsignedByte
	SlotData []pk.Slot
}

func (w *WindowItems) Read(packet pk.Packet) error {
	var count pk.Short
	return packet.Scan(&w.WindowID, &count, &pk.Ary{
		Len: &count,
		Ary: &w.SlotData,
	})
}

func (w *WindowItems) Marshal() pk.Packet {
	return pk.Marshal(ClientboundWindowItems, w.WindowID, pk.Short(len(w.SlotData)), pk.Ary{Ary: w.SlotData})
}

type PlayerAbilities struct {
	Flags               pk.Byte
	FlyingSpeed         pk.Float
	FieldOfViewModifier pk.Float
}

func (p *PlayerAbilities) Read(packet pk.Packet) error {
	return packet.Scan(&p.Flags, &p.FlyingSpeed, &p.FieldOfViewModifier)
}

func (p *PlayerAbilities) Marshal() pk.Packet {
	return pk.Marshal(ClientboundPlayerAbilities, p.Flags, p.FlyingSpeed, p.FieldOfViewModifier)
}

type PluginMessage struct {
	Channel pk.String
	Data    pk.PluginMessageData
}

func (p *PluginMessage) Read(packet pk.Packet) error {
	return packet.Scan(&p.Channel, &p.Data)
}

func (p *PluginMessage) Marshal() pk.Packet {
	return pk.Marshal(ClientboundPluginMessage, p.Channel, &p.Data)
}

type Disconnect struct {
	Reason chat.Message
}

func (d *Disconnect) Read(packet pk.Packet) error {
	return packet.Scan(&d.Reason)
}

func (d *Disconnect) Marshal() pk.Packet {
	return pk.Marshal(ClientboundDisconnect, d.Reason)
}

const (
	ServerboundKeepAlive = iota
	ServerboundChatMessage
	ServerboundUseEntity
	ServerboundPlayer
	ServerboundPlayerPosition
	ServerboundPlayerLook
	ServerboundPlayerPositionAndLook
	ServerboundPlayerDigging
	ServerboundPlayerBlockPlacement
	ServerboundHeldItemChange
	ServerboundAnimation
	ServerboundEntityAction
	ServerboundSteerVehicle
	ServerboundCloseWindow
	ServerboundClickWindow
	ServerboundConfirmTransaction
	ServerboundCreativeInventoryAction
	ServerboundEnchantItem
	ServerboundUpdateSign
	ServerboundPlayerAbilities
	ServerboundTabComplete
	ServerboundClientSettings
	ServerboundClientStatus
	ServerboundPluginMessage
	ServerboundSpectate
	ServerboundResourcePackStatus
)

type ChatMessage struct {
	Message pk.String
}

func (c *ChatMessage) Read(packet pk.Packet) error {
	return packet.Scan(&c.Message)
}

func (c *ChatMessage) Marshal() pk.Packet {
	return pk.Marshal(ServerboundChatMessage, c.Message)
}

type Player struct {
	OnGround pk.Boolean
}

func (p *Player) Read(packet pk.Packet) error {
	return packet.Scan(&p.OnGround)
}

func (p *Player) Marshal() pk.Packet {
	return pk.Marshal(ServerboundPlayer, p.OnGround)
}

type PlayerPosition struct {
	X, Y, Z  pk.Double
	OnGround pk.Boolean
}

func (p *PlayerPosition) Read(packet pk.Packet) error {
	return packet.Scan(&p.X, &p.Y, &p.Z, &p.OnGround)
}

func (p *PlayerPosition) Marshal() pk.Packet {
	return pk.Marshal(ServerboundPlayerPosition, p.X, p.Y, p.Z, p.OnGround)
}

type PlayerLook struct {
	Yaw, Pitch pk.Float
	OnGround   pk.Boolean
}

func (p *PlayerLook) Read(packet pk.Packet) error {
	return packet.Scan(&p.Yaw, &p.Pitch, &p.OnGround)
}

func (p *PlayerLook) Marshal() pk.Packet {
	return pk.Marshal(ServerboundPlayerLook, p.Yaw, p.Pitch, p.OnGround)
}

type ServerPlayerPositionAndLook struct {
	X, Y, Z    pk.Double
	Yaw, Pitch pk.Float
	OnGround   pk.Boolean
}

func (p *ServerPlayerPositionAndLook) Read(packet pk.Packet) error {
	return packet.Scan(&p.X, &p.Y, &p.Z, &p.Yaw, &p.Pitch, &p.OnGround)
}

func (p *ServerPlayerPositionAndLook) Marshal() pk.Packet {
	return pk.Marshal(ServerboundPlayerPositionAndLook, p.X, p.Y, p.Z, p.Yaw, p.Pitch, p.OnGround)
}

type PlayerDigging struct {
	Status   pk.Byte
	Location pk.Position
	Face     pk.Byte
}

func (p *PlayerDigging) Read(packet pk.Packet) error {
	return packet.Scan(&p.Status, &p.Location, &p.Face)
}

func (p *PlayerDigging) Marshal() pk.Packet {
	return pk.Marshal(ServerboundPlayerDigging, p.Status, p.Location, p.Face)
}

type PlayerBlockPlacement struct {
	Location        pk.Position
	Face            pk.Byte
	HeldItem        pk.Slot
	CursorPositionX pk.Byte
	CursorPositionY pk.Byte
	CursorPositionZ pk.Byte
}

func (p *PlayerBlockPlacement) Read(packet pk.Packet) error {
	return packet.Scan(&p.Location, &p.Face, &p.HeldItem, &p.CursorPositionX, &p.CursorPositionY, &p.CursorPositionZ)
}

func (p *PlayerBlockPlacement) Marshal() pk.Packet {
	return pk.Marshal(ServerboundPlayerBlockPlacement, p.Location, p.Face, p.HeldItem, p.CursorPositionX, p.CursorPositionY, p.CursorPositionZ)
}

type ServerHeldItemChange struct {
	Slot pk.Short
}

func (s *ServerHeldItemChange) Read(packet pk.Packet) error {
	return packet.Scan(&s.Slot)
}

func (s *ServerHeldItemChange) Marshal() pk.Packet {
	return pk.Marshal(ServerboundHeldItemChange, s.Slot)
}

type ServerAnimation struct{}

func (s *ServerAnimation) Read(_ pk.Packet) error {
	return nil
}

func (s *ServerAnimation) Marshal() pk.Packet {
	return pk.Marshal(ServerboundAnimation)
}

type ClickWindow struct {
	WindowID     pk.UnsignedByte
	Slot         pk.Short
	Button       pk.Byte
	ActionNumber pk.Short
	Mode         pk.Byte
	ClickedItem  pk.Slot
}

func (c *ClickWindow) Read(packet pk.Packet) error {
	return packet.Scan(&c.WindowID, &c.Slot, &c.Button, &c.ActionNumber, &c.Mode, &c.ClickedItem)
}

func (c *ClickWindow) Marshal() pk.Packet {
	return pk.Marshal(ServerboundClickWindow, c.WindowID, c.Slot, c.Button, c.ActionNumber, c.Mode, c.ClickedItem)
}

type ServerCloseWindow struct {
	CloseWindow
}

func (s *ServerCloseWindow) Marshal() pk.Packet {
	packet := s.CloseWindow.Marshal()
	packet.ID = ServerboundCloseWindow
	return packet
}

type ServerPlayerAbilities struct {
	Flags        pk.Byte
	FlyingSpeed  pk.Float
	WalkingSpeed pk.Float
}

func (s *ServerPlayerAbilities) Read(packet pk.Packet) error {
	return packet.Scan(&s.Flags, &s.FlyingSpeed, &s.WalkingSpeed)
}

func (s *ServerPlayerAbilities) Marshal() pk.Packet {
	return pk.Marshal(ServerboundPlayerAbilities, s.Flags, s.FlyingSpeed, s.WalkingSpeed)
}

type ServerPluginMessage struct {
	PluginMessage
}

func (s *ServerPluginMessage) Marshal() pk.Packet {
	packet := s.PluginMessage.Marshal()
	packet.ID = ServerboundPluginMessage
	return packet
}
