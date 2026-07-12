package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/pion/dtls/v3"
	"github.com/pion/dtls/v3/pkg/crypto/selfsign"
)

func probeServer(address string, timeout time.Duration, wrapKey []byte) error {
	peer, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}

	udp, err := net.ListenUDP("udp", nil)
	if err != nil {
		return fmt.Errorf("open UDP socket: %w", err)
	}

	var packetConn net.PacketConn = udp
	if len(wrapKey) != 0 {
		state, stateErr := newWrapState(wrapKey)
		if stateErr != nil {
			_ = udp.Close()
			return stateErr
		}

		wrapped := &wrapPacketConn{inner: udp, ws: state}
		var seed [22]byte
		if _, err = rand.Read(seed[:]); err != nil {
			_ = udp.Close()
			return fmt.Errorf("initialize WRAP probe: %w", err)
		}
		copy(wrapped.sessionID[:], seed[0:4])
		copy(wrapped.ssrc[:], seed[4:8])
		wrapped.sessionID[0] &^= 0x80
		wrapped.ssrc[0] &^= 0x80
		wrapped.seq.Store(uint32(binary.BigEndian.Uint16(seed[8:10])))
		wrapped.timestamp.Store(binary.BigEndian.Uint32(seed[10:14]))
		wrapped.counter.Store(binary.BigEndian.Uint64(seed[14:22]))
		packetConn = wrapped
	}
	defer func() { _ = packetConn.Close() }()

	certificate, err := selfsign.GenerateSelfSigned()
	if err != nil {
		return fmt.Errorf("generate certificate: %w", err)
	}

	conn, err := dtls.ClientWithOptions(
		packetConn,
		peer,
		dtls.WithCertificates(certificate),
		dtls.WithInsecureSkipVerify(true),
		dtls.WithExtendedMasterSecret(dtls.RequireExtendedMasterSecret),
		dtls.WithCipherSuites(dtls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256),
		dtls.WithConnectionIDGenerator(dtls.OnlySendCIDGenerator()),
	)
	if err != nil {
		return fmt.Errorf("create DTLS client: %w", err)
	}
	defer func() { _ = conn.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err = conn.HandshakeContext(ctx); err != nil {
		return fmt.Errorf("DTLS handshake: %w", err)
	}

	return nil
}
