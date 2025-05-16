package redis

import "testing"

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Address: "localhost:6379",
				DB:      0,
			},
			wantErr: false,
		},
		{
			name: "empty address",
			config: Config{
				Address: "",
				DB:      0,
			},
			wantErr: true,
		},
		{
			name: "invalid DB",
			config: Config{
				Address: "localhost:6379",
				DB:      -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
