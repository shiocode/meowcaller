package main

import (
	"crypto/rand"
	"fmt"
	"log"

	meowcaller "github.com/purpshell/meowcaller"
	"github.com/purpshell/meowcaller/mlow"
)

// loopbackSSRC is an arbitrary fixed SSRC for the local round-trip.
const loopbackSSRC = 0x12345678

// runLoopback pipes the live mic through the whole media stack and back to the
// speaker — MLow encode → E2E-SRTP protect (RTP WARP header + WARP MI tag) →
// unprotect → MLow decode — with no WhatsApp connection. It exercises every byte
// of the codec + keying + framing layers end to end on real audio hardware.
func runLoopback() error {
	a, err := newAudio()
	if err != nil {
		return fmt.Errorf("init audio: %w", err)
	}
	defer a.close()

	mic, stopMic, err := a.openMic()
	if err != nil {
		return fmt.Errorf("open mic: %w", err)
	}
	defer stopMic()
	speaker, stopSpeaker, err := a.openSpeaker()
	if err != nil {
		return fmt.Errorf("open speaker: %w", err)
	}
	defer stopSpeaker()

	enc := mlow.NewMlowEncoder()
	dec := mlow.NewMlowDecoder()

	// Throwaway callKey; same LID both directions so the loopback round-trips
	// (a real call derives send keys from the self LID, recv from the peer LID).
	var callKey [32]byte
	if _, err := rand.Read(callKey[:]); err != nil {
		return err
	}
	const lid = "10000000000000:0@lid"
	send, err := meowcaller.NewMediaPipeline(callKey[:], lid, lid, loopbackSSRC, frameSamps)
	if err != nil {
		return fmt.Errorf("send pipeline: %w", err)
	}
	recv, err := meowcaller.NewMediaPipeline(callKey[:], lid, lid, loopbackSSRC, frameSamps)
	if err != nil {
		return fmt.Errorf("recv pipeline: %w", err)
	}

	log.Println("loopback: mic → MLow → E2E-SRTP protect/unprotect → MLow → speaker. Ctrl+C to stop.")
	var n uint64
	for pcm := range mic {
		payload, err := enc.Encode(pcmToFloat(pcm))
		if err != nil {
			continue // skip frames the encoder rejects rather than abort the call
		}
		packet, err := send.ProtectAudio(payload)
		if err != nil {
			continue
		}
		_, decoded, ok := recv.UnprotectAudio(packet)
		if !ok {
			continue
		}
		speaker <- floatToPCM(dec.Decode(decoded))
		if n++; n%100 == 0 {
			log.Printf("%d frames piped through the voip stack (%ds)", n, n*60/1000)
		}
	}
	return nil
}
