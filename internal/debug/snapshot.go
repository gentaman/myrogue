package debug

type Snapshot struct {
	Floor       int
	Turn        int
	Phase       string
	EntityCount int
	PlayerHP    int
	PlayerMaxHP int
	PlayerMP    int
	PlayerMaxMP int
	PlayerX     int
	PlayerY     int
	AnimSpeed   float64
	AnimPaused  bool
	AnimQueue   int
}
