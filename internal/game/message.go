package game

// Message はシステム内でのイベント通知に使用されるインターフェース
type Message interface {
	IsMessage()
}

// MsgLog はログ出力イベント
type MsgLog struct {
	Text string
}

func (m MsgLog) IsMessage() {}

// MsgSFX は効果音再生イベント
type MsgSFX struct {
	PCM []byte
}

func (m MsgSFX) IsMessage() {}

// MsgDamage はダメージ発生イベント
type MsgDamage struct {
	AttackerID int64
	TargetID   int64
	Damage     int
}

func (m MsgDamage) IsMessage() {}

// MsgDeath は死亡イベント
type MsgDeath struct {
	Battler Battler
}

func (m MsgDeath) IsMessage() {}

// MsgXP は経験値獲得イベント
type MsgXP struct {
	Actor  *Actor
	Amount int
}

func (m MsgXP) IsMessage() {}

// MsgDropChest は宝箱ドロップイベント
type MsgDropChest struct {
	X, Y      int
	Inventory []InventoryEntry
}

func (m MsgDropChest) IsMessage() {}

// MsgUseItem はアイテム使用リクエスト
type MsgUseItem struct {
	Actor *Actor
	Index int
}

func (m MsgUseItem) IsMessage() {}

// MsgChangeFloor は階層移動リクエスト
type MsgChangeFloor struct {
	CurrentFloor int
	Direction    int // 1: 下へ, -1: 上へ
}

func (m MsgChangeFloor) IsMessage() {}

// MsgTransition はシーン遷移リクエスト
type MsgTransition struct {
	Next Scene
}

func (m MsgTransition) IsMessage() {}

// EventBus はメッセージの購読と発行を管理する
type EventBus struct {
	logHandler     func(string)
	sfxHandler     func([]byte)
	damageHandler  func(MsgDamage)
	deathHandler   func(MsgDeath)
	xpHandler      func(MsgXP)
	chestHandler   func(MsgDropChest)
	useItemHandler func(MsgUseItem)
	floorHandler   func(MsgChangeFloor)
	transitHandler func(MsgTransition)
}

func NewEventBus() *EventBus {
	return &EventBus{}
}

func (e *EventBus) OnLog(h func(string))                 { e.logHandler = h }
func (e *EventBus) OnSFX(h func([]byte))                 { e.sfxHandler = h }
func (e *EventBus) OnDamage(h func(MsgDamage))           { e.damageHandler = h }
func (e *EventBus) OnDeath(h func(MsgDeath))             { e.deathHandler = h }
func (e *EventBus) OnXP(h func(MsgXP))                   { e.xpHandler = h }
func (e *EventBus) OnChest(h func(MsgDropChest))         { e.chestHandler = h }
func (e *EventBus) OnUseItem(h func(MsgUseItem))         { e.useItemHandler = h }
func (e *EventBus) OnChangeFloor(h func(MsgChangeFloor)) { e.floorHandler = h }
func (e *EventBus) OnTransition(h func(MsgTransition))   { e.transitHandler = h }

// Publish はメッセージを発行する
func (e *EventBus) Publish(m Message) {
	switch msg := m.(type) {
	case MsgLog:
		if e.logHandler != nil {
			e.logHandler(msg.Text)
		}
	case MsgSFX:
		if e.sfxHandler != nil {
			e.sfxHandler(msg.PCM)
		}
	case MsgDamage:
		if e.damageHandler != nil {
			e.damageHandler(msg)
		}
	case MsgDeath:
		if e.deathHandler != nil {
			e.deathHandler(msg)
		}
	case MsgXP:
		if e.xpHandler != nil {
			e.xpHandler(msg)
		}
	case MsgDropChest:
		if e.chestHandler != nil {
			e.chestHandler(msg)
		}
	case MsgUseItem:
		if e.useItemHandler != nil {
			e.useItemHandler(msg)
		}
	case MsgChangeFloor:
		if e.floorHandler != nil {
			e.floorHandler(msg)
		}
	case MsgTransition:
		if e.transitHandler != nil {
			e.transitHandler(msg)
		}
	}
}
