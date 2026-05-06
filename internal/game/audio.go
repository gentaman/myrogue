package game

func playSFX(pcm []byte) {
	if pcm == nil || sfxVolume <= 0 {
		return
	}
	p := audioContext.NewPlayerFromBytes(pcm)
	p.SetVolume(sfxVolume)
	p.Play()
}

func playSFXHit() { playSFX(sfxHitPCM) }
