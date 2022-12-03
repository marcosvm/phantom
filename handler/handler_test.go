package handler

import "testing"

func TestExtractOrigin(t *testing.T) {
	h := Handler{}
	cases := []struct {
		name    string
		input   string
		origin  string
		proxies string
	}{
		{"zero", "", "unknown", ""},
		{"uno", "10.1.1.1", "10.1.1.1", ""},
		{"duo", "10.1.1.1, 192.168.4.3", "10.1.1.1", "192.168.4.3"},
		{"tree", "10.1.1.1, 192.168.4.3, 192.168.8.3", "10.1.1.1", "192.168.4.3, 192.168.8.3"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o, p := h.extractOrigin(tc.input)
			if tc.origin != o || tc.proxies != p {
				t.Errorf("expecting %q and %q but got %q and %q", tc.origin, tc.proxies, o, p)
			}
		})
	}
}
