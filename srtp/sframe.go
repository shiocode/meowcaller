package srtp

import "errors"

// errBadCallKeyLen is returned when the call key is not exactly 32 bytes, the only
// length the SFrame key derivation accepts.
var errBadCallKeyLen = errors.New("srtp: sframe call key must be exactly 32 bytes")

// SFrame KDF labels and lengths.
const (
	KDFLabelE2ESframe = "e2e sframe key"
	KDFLabelWarpAuth  = "warp auth key"
	gcmTagLen         = 16
	aesKeyLen         = 16
)

// formatParticipantID is the device-qualified participant id: strip the resource,
// give a bare @lid an implicit :0, pass everything else through unchanged.
func formatParticipantID(jid string) string {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/mod.rs#L44-L58
	// TODO
	// agent suggestion: bare = first '/'-segment, TrimSpace; at = LastIndexByte('@');
	// if at<=0 return bare; if domain=="lid" && user has no ':' return user+":0@lid"; else bare.
	// human input:
	return ""
}

// FormatSframeParticipantID formats the participant id used as the SFrame HKDF info.
func FormatSframeParticipantID(jid string) string {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L33-L35
	// TODO
	// agent suggestion: return formatParticipantID(jid) (same surface as E2E-SRTP today).
	// human input:
	return ""
}

// SframeInfoLabel builds the HKDF info label "e2e sframe key<participantID>".
func SframeInfoLabel(participantID string) string {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L37-L39
	// TODO
	// agent suggestion: return KDFLabelE2ESframe + participantID.
	// human input:
	return ""
}

// DeriveE2eSframeKeyForParticipant derives the 32-byte per-participant SFrame key
// from callKey (exactly 32B), salt = callKey[0:16], ikm = callKey[16:32].
func DeriveE2eSframeKeyForParticipant(callKey []byte, participantID string) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L42-L54
	// TODO
	// agent suggestion: salt, ikm, err := splitCallKey(callKey); util.HKDFSHA256(salt, ikm,
	// []byte(SframeInfoLabel(participantID)), 32).
	// human input:
	return nil, nil
}

// DeriveWarpAuthKey derives the 32-byte WARP auth key from callKey (32B), empty
// salt, label "warp auth key".
func DeriveWarpAuthKey(callKey []byte) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L56-L66
	// TODO
	// agent suggestion: if len(callKey)!=32 return errBadCallKeyLen; util.HKDFSHA256(nil, callKey,
	// []byte(KDFLabelWarpAuth), 32). Left a stub — no KAT here; validate under #24 warp.
	// human input:
	return nil, nil
}

// splitCallKey splits a 32-byte call key into (salt, ikm).
func splitCallKey(callKey []byte) (salt, ikm []byte, err error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L22-L27
	// TODO
	// agent suggestion: if len(callKey)!=32 return errBadCallKeyLen; else callKey[0:16], callKey[16:32].
	// human input:
	return nil, nil, nil
}

// counterToIV builds the 16-byte GCM nonce: 8 zero bytes then counter as LE uint64.
func counterToIV(counter uint64) [16]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L88-L92
	// TODO
	// agent suggestion: var iv [16]byte; binary.LittleEndian.PutUint64(iv[8:16], counter); return iv.
	// human input:
	return [16]byte{}
}

// buildSframeHeader encodes [varint counter || varint keyID || total-length byte].
func buildSframeHeader(counter, keyID uint64) []byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L94-L101
	// TODO
	// agent suggestion: h := binary.AppendUvarint(nil, counter); h = binary.AppendUvarint(h, keyID);
	// return append(h, byte(len(h)+1)). binary.AppendUvarint == the reference encode_varint (LEB128).
	// human input:
	return nil
}

// parseSframeHeader decodes the trailing header, validating the total-length byte.
func parseSframeHeader(header []byte) (counter, keyID uint64, ok bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L103-L115
	// TODO
	// agent suggestion: len>=2 and last byte == len(header); decode two binary.Uvarint from the body.
	// human input:
	return 0, 0, false
}

// gcmEncrypt seals plaintext with AES-128-GCM under the non-standard 16-byte nonce.
func gcmEncrypt(key []byte, nonce16 [16]byte, plaintext []byte) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L117-L123
	// TODO
	// agent suggestion: aes.NewCipher(key[:16]); cipher.NewGCMWithNonceSize(block, 16); Seal(nil,
	// nonce16[:], plaintext, nil) — nil AAD, 16-byte nonce so the GHASH-derived J0 matches.
	// human input:
	return nil, nil
}

// gcmDecrypt opens ciphertext+tag with AES-128-GCM; ok=false on any auth failure.
func gcmDecrypt(key []byte, nonce16 [16]byte, ciphertextWithTag []byte) ([]byte, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L125-L129
	// TODO
	// agent suggestion: build the same GCM; Open(nil, nonce16[:], ciphertextWithTag, nil); ok=false on err.
	// human input:
	return nil, false
}

// SframeSession holds the per-direction SFrame keys (encrypt for peer, decrypt for
// self) and the send-side counter.
type SframeSession struct {
	SelfParticipantID string
	PeerParticipantID string
	encryptKey        [16]byte
	decryptKey        [16]byte
	txCounter         uint64
}

// NewSframeSession builds a session from callKey and the self/peer JIDs.
func NewSframeSession(callKey []byte, selfJID, peerJID string) (*SframeSession, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L154-L172
	// TODO
	// agent suggestion: selfID/peerID via FormatSframeParticipantID; sendKey from peerID, recvKey
	// from selfID via DeriveE2eSframeKeyForParticipant; copy [:16] into encrypt/decrypt keys.
	// human input:
	return nil, nil
}

// Encrypt seals one frame as [ciphertext || 16-byte tag || varint-header].
func (s *SframeSession) Encrypt(plaintext []byte) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L173-L187
	// TODO
	// agent suggestion: counter := s.txCounter; s.txCounter++; header := buildSframeHeader(counter, 0);
	// iv := counterToIV(counter); ct, err := gcmEncrypt(encryptKey, iv, plaintext); return ct||header.
	// human input:
	return nil, nil
}

// Decrypt classifies one frame. It returns (plaintext, true) when the trailing
// SFrame header parses and the GCM tag authenticates; otherwise (nil, false),
// meaning the frame is plain Opus the caller must use verbatim.
func (s *SframeSession) Decrypt(frame []byte) ([]byte, bool) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/sframe.rs#L188-L213
	// TODO
	// agent suggestion: guard len>=tag+3, headerLen from last byte, slice header/ciphertext, parse
	// header for counter, gcmDecrypt(decryptKey, counterToIV(counter), ciphertext); GCM auth is the
	// sole discriminator — any miss returns (nil, false).
	// human input:
	return nil, false
}
