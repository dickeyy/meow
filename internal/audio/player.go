package audio

import (
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

const (
	channels   = 2
	sampleRate = 48000
	frameSize  = 960 // 20ms at 48kHz
	maxBytes   = (frameSize * channels) * 2
)

type Player struct {
	mu sync.Mutex
}

func NewPlayer() *Player {
	return &Player{}
}

func (p *Player) Play(session *Session, discord *discordgo.Session) error {
	vc := session.VoiceConnection()
	if vc == nil {
		return fmt.Errorf("not connected to voice channel")
	}

	track := session.Queue().Current()
	if track == nil {
		return fmt.Errorf("no track to play")
	}

	if track.StreamURL == "" {
		return fmt.Errorf("track has no stream URL")
	}

	session.SetState(StatePlaying)
	session.SetStartedAt(time.Now())

	if session.OnTrackChange != nil {
		session.OnTrackChange(track)
	}

	go p.stream(session, track)

	return nil
}

func (p *Player) stream(session *Session, track *Track) {
	vc := session.VoiceConnection()
	if vc == nil {
		return
	}

	vc.Speaking(true)
	defer vc.Speaking(false)

	volume := float64(session.Volume()) / 100.0

	// Create FFmpeg command to get raw PCM audio
	ffmpeg := exec.Command("ffmpeg",
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-i", track.StreamURL,
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"-af", fmt.Sprintf("volume=%.2f", volume),
		"-loglevel", "warning",
		"pipe:1",
	)

	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		fmt.Printf("Failed to get ffmpeg stdout: %v\n", err)
		session.SetState(StateStopped)
		return
	}

	if err := ffmpeg.Start(); err != nil {
		fmt.Printf("Failed to start ffmpeg: %v\n", err)
		session.SetState(StateStopped)
		return
	}

	defer func() {
		ffmpeg.Process.Kill()
		ffmpeg.Wait()
	}()

	// Create Opus encoder
	encoder, err := gopus.NewEncoder(sampleRate, channels, gopus.Audio)
	if err != nil {
		fmt.Printf("Failed to create opus encoder: %v\n", err)
		session.SetState(StateStopped)
		return
	}

	// Set encoder bitrate (128kbps for quality)
	encoder.SetBitrate(128000)

	// Buffer for reading PCM data
	pcmBuffer := make([]int16, frameSize*channels)
	byteBuffer := make([]byte, maxBytes)

	for {
		select {
		case <-session.StopChan():
			session.Queue().ClearAll()
			session.SetState(StateStopped)
			if session.OnTrackEnd != nil {
				session.OnTrackEnd()
			}
			return

		case <-session.SkipChan():
			next := session.Queue().Next()
			if next != nil {
				session.SetState(StateStopped)
				go p.playNext(session, next)
			} else {
				session.SetState(StateStopped)
				if session.OnTrackEnd != nil {
					session.OnTrackEnd()
				}
			}
			return

		case <-session.PauseChan():
			// Wait for resume or stop
			select {
			case <-session.ResumeChan():
				// Continue streaming
			case <-session.StopChan():
				session.Queue().ClearAll()
				session.SetState(StateStopped)
				if session.OnTrackEnd != nil {
					session.OnTrackEnd()
				}
				return
			case <-session.SkipChan():
				next := session.Queue().Next()
				if next != nil {
					session.SetState(StateStopped)
					go p.playNext(session, next)
				} else {
					session.SetState(StateStopped)
					if session.OnTrackEnd != nil {
						session.OnTrackEnd()
					}
				}
				return
			}

		default:
			// Read PCM data from ffmpeg
			n, err := io.ReadFull(stdout, byteBuffer)
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					// Track finished
					next := session.Queue().Next()
					if next != nil {
						go p.playNext(session, next)
					} else {
						session.SetState(StateStopped)
						if session.OnTrackEnd != nil {
							session.OnTrackEnd()
						}
					}
					return
				}
				fmt.Printf("Error reading from ffmpeg: %v\n", err)
				session.SetState(StateStopped)
				return
			}

			if n < maxBytes {
				continue
			}

			// Convert bytes to int16
			for i := 0; i < frameSize*channels; i++ {
				pcmBuffer[i] = int16(binary.LittleEndian.Uint16(byteBuffer[i*2 : (i+1)*2]))
			}

			// Encode to Opus
			opus, err := encoder.Encode(pcmBuffer, frameSize, maxBytes)
			if err != nil {
				fmt.Printf("Error encoding opus: %v\n", err)
				continue
			}

			// Send to Discord
			if vc.Ready && vc.OpusSend != nil {
				select {
				case vc.OpusSend <- opus:
				case <-time.After(time.Second):
					// Timeout sending
				}
			}
		}
	}
}

func (p *Player) playNext(session *Session, track *Track) {
	session.SetState(StatePlaying)
	session.SetStartedAt(time.Now())

	if session.OnTrackChange != nil {
		session.OnTrackChange(track)
	}

	p.stream(session, track)
}

func (p *Player) PlayPrevious(session *Session) error {
	prev := session.Queue().Previous()
	if prev == nil {
		return fmt.Errorf("no previous track")
	}

	session.Skip()
	return nil
}
