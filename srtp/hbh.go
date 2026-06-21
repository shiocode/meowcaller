package srtp

import "errors"

// errBadHbhKeyLen is returned when the hop-by-hop key is not exactly 30 bytes,
// the only valid length the relay produces.
var errBadHbhKeyLen = errors.New("srtp: hbh key must be exactly 30 bytes")

// hbh session-key derivation labels (libsrtp srtp_kdf_generate).
const (
	labelRTPEncryption = 0x00
	labelRTPAuth       = 0x01
	labelRTPSalt       = 0x02
)

// SrtpKeyingMaterial is the 16-byte master key + 14-byte master salt split.
type SrtpKeyingMaterial struct {
	MasterKey  [16]byte
	MasterSalt [14]byte
}

// LibsrtpSessionKeys is the expanded per-session keying (AES_CM_128_HMAC_SHA1_80).
type LibsrtpSessionKeys struct {
	SessionKey  [16]byte
	SessionSalt [14]byte
	AuthKey     [20]byte
}

// keyingFromCryptoKey splits a 30-byte crypto key into master key (16) + salt (14).
func keyingFromCryptoKey(cryptoKey []byte) SrtpKeyingMaterial {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L34-L42
	// TODO
	// agent suggestion: var m SrtpKeyingMaterial; copy(m.MasterKey[:], cryptoKey[0:16]);
	// copy(m.MasterSalt[:], cryptoKey[16:30]); return m.
	// human input:
	return SrtpKeyingMaterial{}
}

// deriveHbhSrtpKeyWithLabels runs the two-stage WA-SFU KDF (HKDF-SHA256 with the
// literal label as info): stage 1 derives the srtcp salt, stage 2 the 30-byte key.
func deriveHbhSrtpKeyWithLabels(hbhKey []byte, saltLabel, keyLabel string) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L46-L64
	// TODO
	// agent suggestion: if len(hbhKey)!=30 return nil, errBadHbhKeyLen; masterKey=hbhKey[0:16],
	// masterSalt=hbhKey[16:30]; srtcpSalt := util.HKDFSHA256(make([]byte,32), masterSalt,
	// []byte(saltLabel), 32); return util.HKDFSHA256(srtcpSalt, masterKey, []byte(keyLabel), 30).
	// human input:
	return nil, nil
}

// DeriveHbhSrtpKeyUplink derives the 30-byte uplink HBH SRTP key from hbhKey (30B).
func DeriveHbhSrtpKeyUplink(hbhKey []byte) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L66-L68
	// TODO
	// agent suggestion: deriveHbhSrtpKeyWithLabels(hbhKey, "uplink hbh srtcp salt", "uplink hbh srtcp key").
	// human input:
	return nil, nil
}

// DeriveHbhSrtpKeyDownlink derives the 30-byte downlink HBH SRTP key from hbhKey (30B).
func DeriveHbhSrtpKeyDownlink(hbhKey []byte) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L70-L72
	// TODO
	// agent suggestion: deriveHbhSrtpKeyWithLabels(hbhKey, "downlink hbh srtcp salt", "downlink hbh srtcp key").
	// human input:
	return nil, nil
}

// KeyingFromHbhKeyUplink derives the uplink key and splits it into keying material.
func KeyingFromHbhKeyUplink(hbhKey []byte) (SrtpKeyingMaterial, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L74-L78
	// TODO
	// agent suggestion: k, err := DeriveHbhSrtpKeyUplink(hbhKey); if err return zero, err;
	// return keyingFromCryptoKey(k), nil.
	// human input:
	return SrtpKeyingMaterial{}, nil
}

// KeyingFromHbhKeyDownlink derives the downlink key and splits it into keying material.
func KeyingFromHbhKeyDownlink(hbhKey []byte) (SrtpKeyingMaterial, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L80-L84
	// TODO
	// agent suggestion: k, err := DeriveHbhSrtpKeyDownlink(hbhKey); ...; return keyingFromCryptoKey(k), nil.
	// human input:
	return SrtpKeyingMaterial{}, nil
}

// aesICMKey30 concatenates a 16-byte AES key with a 14-byte salt into the 30-byte
// libsrtp AES-ICM key layout.
func aesICMKey30(aesKey, salt []byte) [30]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L87-L92
	// TODO
	// agent suggestion: var out [30]byte; copy(out[:16], aesKey[:16]); copy(out[16:30], salt[:14]); return out.
	// human input:
	return [30]byte{}
}

// aesICMCrypt is libsrtp AES-ICM: counter = (salt padded to 16) XOR iv, keystream =
// AES(counter), counter increments byte 15 with a single carry into byte 14 (2-level,
// NOT a 128-bit CTR — this divergence is faithful to libsrtp and load-bearing).
func aesICMCrypt(key30, iv16, data []byte) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L96-L121
	// TODO
	// agent suggestion: aesKey=key30[:16], salt=key30[16:30]; counter[0:14]=salt; counter[i]^=iv16[i]
	// for all 16; block,err := aes.NewCipher(aesKey); out := append([]byte(nil), data...); per 16-byte
	// chunk encrypt counter -> XOR into out; then counter[15]++ and if it wrapped to 0, counter[14]++.
	// Do NOT use cipher.NewCTR (it carries across all 16 bytes).
	// human input:
	return nil, nil
}

// deriveSessionBytes is libsrtp srtp_kdf_generate: IV all-zero except byte 7 = label.
func deriveSessionBytes(masterKey, masterSalt []byte, label byte, length int) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L124-L129
	// TODO
	// agent suggestion: kdfKey := aesICMKey30(masterKey, masterSalt); var iv [16]byte; iv[7]=label;
	// return aesICMCrypt(kdfKey[:], iv[:], make([]byte, length)).
	// human input:
	return nil, nil
}

// ExpandLibsrtpSessionKeys runs the libsrtp session-key expansion (labels 0x00 enc,
// 0x01 auth, 0x02 salt).
func ExpandLibsrtpSessionKeys(keying *SrtpKeyingMaterial) (LibsrtpSessionKeys, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L132-L157
	// TODO
	// agent suggestion: derive session_key (0x00,16), session_salt (0x02,14), auth_key (0x01,20)
	// via deriveSessionBytes from keying.MasterKey/MasterSalt; copy each into the struct; bubble errors.
	// human input:
	return LibsrtpSessionKeys{}, nil
}

// BuildRtpICMNonce builds the RTP AES-ICM nonce: zero, SSRC at bytes 4-7 (BE),
// (packetIndex << 16) at bytes 8-15 (BE).
func BuildRtpICMNonce(ssrc uint32, packetIndex uint64) [16]byte {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L160-L165
	// TODO
	// agent suggestion: var iv [16]byte; binary.BigEndian.PutUint32(iv[4:8], ssrc);
	// binary.BigEndian.PutUint64(iv[8:16], packetIndex<<16); return iv.
	// human input:
	return [16]byte{}
}

// CryptRtpPayload encrypts/decrypts an RTP payload with the expanded session key
// (symmetric).
func CryptRtpPayload(session *LibsrtpSessionKeys, ssrc uint32, packetIndex uint64, payload []byte) ([]byte, error) {
	// Source of truth: https://github.com/oxidezap/whatsapp-rust/blob/41095d4e6ba4610e054e9ede3af1d5e88a83faee/wacore/src/voip/hbh_srtp.rs#L168-L177
	// TODO
	// agent suggestion: icmKey := aesICMKey30(session.SessionKey[:], session.SessionSalt[:]);
	// nonce := BuildRtpICMNonce(ssrc, packetIndex); return aesICMCrypt(icmKey[:], nonce[:], payload).
	// human input:
	return nil, nil
}
