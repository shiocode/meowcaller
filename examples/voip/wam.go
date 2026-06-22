package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/purpshell/wamsdk"
	"github.com/purpshell/wamsdk/meowmetrics"
	"go.mau.fi/whatsmeow"
	waBinary "go.mau.fi/whatsmeow/binary"
)

// WAM (WhatsApp metrics) protection: a real client periodically reports telemetry
// (WAM) events; a session that never does stands out. wamsdk + meowmetrics emit a
// believable Chrome-on-macOS metrics stream, delivered as `w:stats` IQ stanzas.
// Pattern adapted from meowmeow (only the WAM wiring is taken).

const wamFingerprintFile = "wa-voip-wam.json"

// wamTransport ships encoded WAM payloads as `<iq xmlns="w:stats"><add t=…>` and
// persists the rotating fingerprint between runs.
type wamTransport struct {
	sendNode func(ctx context.Context, node waBinary.Node) error
}

func (t *wamTransport) SendWAMData(data []byte) error {
	return t.sendNode(context.Background(), waBinary.Node{
		Tag: "iq",
		Attrs: waBinary.Attrs{
			"to":    "s.whatsapp.net",
			"type":  "set",
			"xmlns": "w:stats",
			"id":    fmt.Sprintf("wam-%d", time.Now().UnixMilli()),
		},
		Content: []waBinary.Node{{
			Tag:     "add",
			Attrs:   waBinary.Attrs{"t": fmt.Sprintf("%d", time.Now().Unix())},
			Content: data,
		}},
	})
}

func (t *wamTransport) SaveFingerprint(data json.RawMessage) {
	_ = os.WriteFile(wamFingerprintFile, data, 0o644)
}

// setupWAM builds the metrics client; call onConnect after the socket is up.
func setupWAM(cli *whatsmeow.Client) (metrics *meowmetrics.Client, onConnect func()) {
	transport := &wamTransport{sendNode: cli.DangerousInternals().SendNode}
	var savedFP json.RawMessage
	if data, err := os.ReadFile(wamFingerprintFile); err == nil {
		savedFP = data
	}
	wamClient := wamsdk.NewClient(savedFP, transport, wamsdk.WithProfile(wamsdk.ChromeMacProfile()))
	metrics = meowmetrics.New(cli, wamClient)
	return metrics, func() { wamClient.OnConnect(0, "android") }
}
