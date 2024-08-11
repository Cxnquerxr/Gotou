package main

import (
	"log"
	"os"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/wav"
)

var pause = make(chan bool)
var mute = make(chan bool)

var playing = false
var quit = false

var volumeLevel = float64(0)
var defSR = beep.SampleRate(48000)

func playFlac(file os.File) {
	streamer, format, err := flac.Decode(&file)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()
	defer file.Close()

	playing = true

	ctrl := &beep.Ctrl{Streamer: streamer, Paused: false}
	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   volumeLevel,
		Silent:   false,
	}
	resample := beep.Resample(3, format.SampleRate, defSR, volume)
	speaker.Init(defSR, defSR.N(time.Second/5))
	done := make(chan bool)
	speaker.Play(beep.Seq(resample, beep.Callback(func() {
		done <- true
	})))

	for {
		select {
		case <-done:
			playing = false
			return
		default:
			select {
			case <-time.After(time.Second):

			case <-pause:
				speaker.Lock()
				ctrl.Paused = !ctrl.Paused
				speaker.Unlock()
			case <-mute:
				if volume.Silent {
					volume.Silent = false
				} else {
					volume.Silent = true
				}
			}
			if quit {
				quit = false
				streamer.Close()
				file.Close()
				return
			}
		}
	}
}

func playWav(file os.File) {
	streamer, format, err := wav.Decode(&file)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()
	defer file.Close()

	playing = true

	ctrl := &beep.Ctrl{Streamer: streamer, Paused: false}
	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   volumeLevel,
		Silent:   false,
	}
	resample := beep.Resample(3, format.SampleRate, defSR, volume)
	speaker.Init(defSR, defSR.N(time.Second/5))
	done := make(chan bool)
	speaker.Play(beep.Seq(resample, beep.Callback(func() {
		done <- true
	})))

	for {
		select {
		case <-done:
			playing = false
			return
		default:
			select {
			case <-time.After(time.Second):

			case <-pause:
				speaker.Lock()
				ctrl.Paused = !ctrl.Paused
				speaker.Unlock()
			case <-mute:
				if volume.Silent {
					volume.Silent = false
				} else {
					volume.Silent = true
				}
			}
			if quit {
				quit = false
				streamer.Close()
				file.Close()
				return
			}
		}
	}
}

func playMp3(file os.File) {
	streamer, format, err := mp3.Decode(&file)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()
	defer file.Close()

	playing = true

	ctrl := &beep.Ctrl{Streamer: streamer, Paused: false}
	volume := &effects.Volume{
		Streamer: ctrl,
		Base:     2,
		Volume:   volumeLevel,
		Silent:   false,
	}
	resample := beep.Resample(3, format.SampleRate, defSR, volume)
	speaker.Init(defSR, defSR.N(time.Second/5))
	done := make(chan bool)
	speaker.Play(beep.Seq(resample, beep.Callback(func() {
		done <- true
	})))

	for {
		select {
		case <-done:
			playing = false
			return
		default:
			select {
			case <-time.After(time.Second):

			case <-pause:
				speaker.Lock()
				ctrl.Paused = !ctrl.Paused
				speaker.Unlock()
			case <-mute:
				if volume.Silent {
					volume.Silent = false
				} else {
					volume.Silent = true
				}
			}
			if quit {
				quit = false
				streamer.Close()
				file.Close()
				return
			}
		}
	}
}

func togglePlay() {
	pause <- true
}

func toggleMute() {
	mute <- true
}

func fileType(song string, artist string) {
	fileDetails := testList[item{song, artist}]
	file, err := os.Open(fileDetails.directory + song)
	if err != nil {
		os.Exit(1)
	}

	if playing {
		speaker.Clear()
		quit = true
		time.Sleep(time.Millisecond * 100)
		playing = false
	}

	switch fileDetails.fileExt {
	case 1:
		go playFlac(*file)
	case 2:
		go playWav(*file)
	case 3:
		go playMp3(*file)
	}
}
