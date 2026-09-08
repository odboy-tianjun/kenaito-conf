package service

import (
	"fmt"
	"testing"

	"kenaito-conf/server/internal/model"
)

func TestParseConfigToKVMap(t *testing.T) {
	svc := NewConfigParseService()

	tests := []struct {
		name    string
		config  model.AppConfig
		wantKey string
		wantVal string
		wantErr bool
	}{
		{
			name:    "properties with placeholder",
			config:  model.AppConfig{Type: 1, Name: "app.properties", Contents: "a=1\nb=${a}-2\n"},
			wantKey: "b",
			wantVal: "1-2",
		},
		{
			name:    "json",
			config:  model.AppConfig{Type: 3, Name: "app.json", Contents: `{"a":"1","b":"2"}`},
			wantKey: "a",
			wantVal: "1",
		},
		{
			name:    "invalid json",
			config:  model.AppConfig{Type: 3, Name: "app.json", Contents: `{"a":`},
			wantErr: true,
		},
		{
			name:    "yaml",
			config:  model.AppConfig{Type: 1, Name: "app.yaml", Contents: "a:\n  b: 1\n"},
			wantKey: "a.b",
			wantVal: "1",
		},
		{
			name:    "invalid yaml",
			config:  model.AppConfig{Type: 1, Name: "app.yaml", Contents: "a: [1,\n"},
			wantErr: true,
		},
		{
			name:    "xml with escaped value",
			config:  model.AppConfig{Type: 1, Name: "app.xml", Contents: `<properties><entry key="a">1 &amp; 2</entry></properties>`},
			wantKey: "a",
			wantVal: "1 & 2",
		},
		{
			name:    "invalid xml",
			config:  model.AppConfig{Type: 1, Name: "app.xml", Contents: `<properties><entry key="a">1</properties>`},
			wantErr: true,
		},
		{
			name:    "key-value",
			config:  model.AppConfig{Type: 2, Name: "a", Contents: "1"},
			wantKey: "a",
			wantVal: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.ParseConfigToKVMap(&tt.config)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			value, ok := got[tt.wantKey]
			if !ok {
				t.Fatalf("key %q not found in %v", tt.wantKey, got)
			}
			if fmt.Sprint(value) != tt.wantVal {
				t.Fatalf("value of %q = %v, want %q", tt.wantKey, value, tt.wantVal)
			}
		})
	}
}
