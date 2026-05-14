package turn

type Phase int

const (
	PhasePlayerInput Phase = iota
	PhaseCompanionAct
	PhaseEnemyAct
)

type Scheduler struct {
	Phase     Phase
	ActiveIdx int
	TurnCount int
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		Phase: PhasePlayerInput,
	}
}

func (s *Scheduler) StartEnemyPhase() {
	s.Phase = PhaseEnemyAct
	s.ActiveIdx = 0
}

func (s *Scheduler) StartCompanionPhase() {
	s.Phase = PhaseCompanionAct
	s.ActiveIdx = 0
}

func (s *Scheduler) StartPlayerPhase() {
	s.Phase = PhasePlayerInput
}

func (s *Scheduler) AdvanceIdx() {
	s.ActiveIdx++
}

func (s *Scheduler) IncrementTurn() {
	s.TurnCount++
}
