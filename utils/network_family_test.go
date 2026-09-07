package utils

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestPrecheckFamilyDialContextDoesNotFollowAAAAFirstResult(t *testing.T) {
	for _, family := range []string{"tcp4", "tcp6"} {
		t.Run(family, func(t *testing.T) {
			var gotNetwork, gotAddress string
			wantErr := errors.New("record dial")
			dial := precheckFamilyDialContext(func(_ context.Context, network, address string) (net.Conn, error) {
				gotNetwork, gotAddress = network, address
				return nil, wantErr
			}, family)
			_, err := dial(context.Background(), "tcp", "aaaa-first.example:443")
			if !errors.Is(err, wantErr) {
				t.Fatalf("dial error = %v, want capture error", err)
			}
			if gotNetwork != family || gotAddress != "aaaa-first.example:443" {
				t.Fatalf("dial = %q %q, want %q %q", gotNetwork, gotAddress, family, "aaaa-first.example:443")
			}
		})
	}
}
