package audio

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type PlayState int

const (
	StateStopped PlayState = iota
	StatePlaying
	StatePaused
)

type Session struct {
	guildID         string
	channelID       string
	voiceConnection *discordgo.VoiceConnection
	queue           *Queue
	state           PlayState
	volume          int
	startedAt       time.Time
	pausedAt        time.Time
	pausedDuration  time.Duration
	mu              sync.RWMutex

	// Playback control
	stopChan   chan struct{}
	pauseChan  chan struct{}
	resumeChan chan struct{}
	skipChan   chan struct{}

	// Callback when track changes
	OnTrackChange func(track *Track)
	OnTrackEnd    func()
}

func NewSession(guildID string, defaultVolume int) *Session {
	return &Session{
		guildID:    guildID,
		queue:      NewQueue(),
		state:      StateStopped,
		volume:     defaultVolume,
		stopChan:   make(chan struct{}, 1),
		pauseChan:  make(chan struct{}, 1),
		resumeChan: make(chan struct{}, 1),
		skipChan:   make(chan struct{}, 1),
	}
}

func (s *Session) GuildID() string {
	return s.guildID
}

func (s *Session) SetVoiceConnection(vc *discordgo.VoiceConnection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.voiceConnection = vc
}

func (s *Session) VoiceConnection() *discordgo.VoiceConnection {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.voiceConnection
}

func (s *Session) SetChannelID(channelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channelID = channelID
}

func (s *Session) ChannelID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.channelID
}

func (s *Session) Queue() *Queue {
	return s.queue
}

func (s *Session) State() PlayState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}

func (s *Session) SetState(state PlayState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = state
}

func (s *Session) Volume() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.volume
}

func (s *Session) SetVolume(vol int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if vol < 0 {
		vol = 0
	}
	if vol > 100 {
		vol = 100
	}
	s.volume = vol
}

func (s *Session) IsPlaying() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state == StatePlaying
}

func (s *Session) IsPaused() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state == StatePaused
}

func (s *Session) IsStopped() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state == StateStopped
}

func (s *Session) StartedAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.startedAt
}

func (s *Session) SetStartedAt(t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.startedAt = t
	s.pausedDuration = 0
}

func (s *Session) Elapsed() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.startedAt.IsZero() {
		return 0
	}

	if s.state == StatePaused && !s.pausedAt.IsZero() {
		return s.pausedAt.Sub(s.startedAt) - s.pausedDuration
	}

	return time.Since(s.startedAt) - s.pausedDuration
}

func (s *Session) Pause() {
	s.mu.Lock()
	if s.state != StatePlaying {
		s.mu.Unlock()
		return
	}
	s.state = StatePaused
	s.pausedAt = time.Now()
	s.mu.Unlock()

	select {
	case s.pauseChan <- struct{}{}:
	default:
	}
}

func (s *Session) Resume() {
	s.mu.Lock()
	if s.state != StatePaused {
		s.mu.Unlock()
		return
	}
	s.pausedDuration += time.Since(s.pausedAt)
	s.state = StatePlaying
	s.pausedAt = time.Time{}
	s.mu.Unlock()

	select {
	case s.resumeChan <- struct{}{}:
	default:
	}
}

func (s *Session) Skip() {
	select {
	case s.skipChan <- struct{}{}:
	default:
	}
}

func (s *Session) Stop() {
	s.mu.Lock()
	s.state = StateStopped
	s.startedAt = time.Time{}
	s.pausedAt = time.Time{}
	s.pausedDuration = 0
	s.mu.Unlock()

	select {
	case s.stopChan <- struct{}{}:
	default:
	}
}

func (s *Session) StopChan() <-chan struct{} {
	return s.stopChan
}

func (s *Session) PauseChan() <-chan struct{} {
	return s.pauseChan
}

func (s *Session) ResumeChan() <-chan struct{} {
	return s.resumeChan
}

func (s *Session) SkipChan() <-chan struct{} {
	return s.skipChan
}

