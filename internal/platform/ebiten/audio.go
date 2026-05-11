package ebiten

type Audio struct{}

func (a *Audio) PlaySFX(id string) {
	var pcm []byte
	switch id {
	case "hit":
		pcm = sfxHitPCM
	case "stair_down":
		pcm = sfxStairDownPCM
	case "stair_up":
		pcm = sfxStairUpPCM
	case "coin":
		pcm = sfxCoinPCM
	}
	if pcm == nil || sfxVolume <= 0 {
		return
	}
	p := audioContext.NewPlayerFromBytes(pcm)
	p.SetVolume(sfxVolume)
	p.Play()
}
