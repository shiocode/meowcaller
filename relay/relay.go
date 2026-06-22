package relay

import (
	"net"

	"github.com/pion/datachannel"
)

// Relay media transport: a pre-negotiated WebRTC DataChannel over
// SCTP-over-DTLS-over-UDP to a single WhatsApp relay endpoint. Only
// ClassifyRelayPacket is unit-testable; the connection path talks to a live relay.

// RelayPacketKind classifies a packet seen on the relay channel by its first byte.
type RelayPacketKind int

const (
	RelayPacketStun RelayPacketKind = iota
	RelayPacketRtcp
	RelayPacketRtp
	RelayPacketOther
)

const (
	// DataChannelLabel is the pre-negotiated (id=0) DataChannel label WA Web uses.
	DataChannelLabel = "pre-negotiated"
	// SctpPort is the SCTP-over-DTLS WebRTC port.
	SctpPort = 5000
)

// ClassifyRelayPacket demuxes by first byte: top two bits zero ⇒ STUN; 0x80/0x81 ⇒
// RTCP; 0x90 ⇒ RTP (WARP); anything else ⇒ Other.
func ClassifyRelayPacket(data []byte) RelayPacketKind {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/src/voip/transport.rs#L57-L70
	// TODO
	// agent suggestion: if len<2 return Other; first := data[0]; if first&0xc0 != 0 switch first
	// {0x80,0x81: Rtcp; 0x90: Rtp; default: Other}; else Stun.
	// human input:
	return RelayPacketOther
}

// CallTransportError categorizes a relay-transport failure so a consumer can branch:
// Connect is fatal (the call can't reach the relay); Send/Recv are recoverable on an
// established channel.
type CallTransportError struct {
	Op  string // "connect", "send", or "recv"
	Err error
}

func (e *CallTransportError) Error() string { return "relay " + e.Op + ": " + e.Err.Error() }
func (e *CallTransportError) Unwrap() error { return e.Err }

// RelayMediaChannel is an open relay media channel; STUN/RTP/RTCP travel as binary
// DataChannel messages.
type RelayMediaChannel struct {
	dc *datachannel.DataChannel
}

// Send writes one media/STUN packet as a binary DataChannel message.
func (c *RelayMediaChannel) Send(data []byte) (int, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/src/voip/transport.rs#L118-L124
	// TODO
	// agent suggestion: n, err := c.dc.Write(data); wrap err in &CallTransportError{Op:"send"}.
	// human input:
	return 0, nil
}

// Recv reads one DataChannel message into buf, returning its length.
func (c *RelayMediaChannel) Recv(buf []byte) (int, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/src/voip/transport.rs#L126-L132
	// TODO
	// agent suggestion: n, err := c.dc.Read(buf); wrap err in &CallTransportError{Op:"recv"}.
	// human input:
	return 0, nil
}

// ConnectRelayMedia connects the full media stack (UDP→DTLS→SCTP→DataChannel) to one
// relay endpoint. No vector — validated only against a live relay.
func ConnectRelayMedia(relayAddr *net.UDPAddr) (*RelayMediaChannel, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/src/voip/transport.rs#L136-L195
	// TODO
	// agent suggestion: ListenUDP -> dtls.Client(conn, relayAddr, &dtls.Config{Certificates:
	// [selfsign], InsecureSkipVerify:true}) -> sctp.Client{NetConn: dtlsConn} -> datachannel.Dial(assoc,
	// 0, &datachannel.Config{Negotiated:true, Label: DataChannelLabel}); wrap failures in
	// &CallTransportError{Op:"connect"}.
	// human input:
	return nil, nil
}
