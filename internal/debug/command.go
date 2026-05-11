package debug

type CommandType int

const (
	CmdToggleHUD CommandType = iota
	CmdToggleFOV
	CmdToggleGrid
	CmdToggleEntityID
	CmdRevealMap
	CmdSlowAnim
	CmdFastAnim
	CmdPauseAnim
	CmdStepFrame
	CmdSkipAnims
	CmdHealPlayer
	CmdKillAll
)

type Command struct {
	Type   CommandType
	Amount int
}
