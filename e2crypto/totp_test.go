package e2crypto

import (
	"testing"
	"time"
)

// Test TOTP Configuration
func TestTOTPConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  TOTPConfig
		wantErr bool
	}{
		{
			name:    "valid default config",
			config:  DefaultTOTPConfig("JBSWY3DPEHPK3PXP"),
			wantErr: false,
		},
		{
			name: "valid custom config",
			config: TOTPConfig{
				Secret:    "JBSWY3DPEHPK3PXP",
				Digits:    8,
				Algorithm: TOTPAlgorithmSHA256,
				Period:    60,
			},
			wantErr: false,
		},
		{
			name: "empty secret",
			config: TOTPConfig{
				Secret:    "",
				Digits:    6,
				Algorithm: TOTPAlgorithmSHA1,
				Period:    30,
			},
			wantErr: true,
		},
		{
			name: "invalid digits",
			config: TOTPConfig{
				Secret:    "JBSWY3DPEHPK3PXP",
				Digits:    5,
				Algorithm: TOTPAlgorithmSHA1,
				Period:    30,
			},
			wantErr: true,
		},
		{
			name: "invalid period",
			config: TOTPConfig{
				Secret:    "JBSWY3DPEHPK3PXP",
				Digits:    6,
				Algorithm: TOTPAlgorithmSHA1,
				Period:    0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TOTPConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Test TOTP Generation and Verification
func TestTOTP(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP" // Base32 encoded "Hello!"
	now := time.Now()

	t.Run("generate and verify", func(t *testing.T) {
		config := DefaultTOTPConfig(secret)
		code, err := GenerateTOTP(secret, now, config.Period)
		if err != nil {
			t.Errorf("GenerateTOTP() error = %v", err)
		}

		if len(code) != 6 {
			t.Errorf("Expected 6-digit code, got %d digits: %s", len(code), code)
		}

		t.Logf("Generated TOTP code: %s", code)

		if !VerifyTOTPWithConfig(DefaultTOTPConfig(secret), code, now, config.Period) {
			t.Error("Failed to verify valid TOTP code")
		}

	})

	t.Run("time skew", func(t *testing.T) {
		config := DefaultTOTPConfig(secret)
		code, err := GenerateTOTP(secret, now, config.Period)
		if err != nil {
			t.Errorf("GenerateTOTP() error = %v", err)
		}

		// Code should verify with time skew
		pastTime := now.Add(-time.Duration(config.Period) * time.Second)
		//if !VerifyTOTP(secret, code, pastTime, config.Period, 1) {
		//	t.Error("Code should verify with time skew")
		//}
		if !VerifyTOTPWithConfig(DefaultTOTPConfig(secret), code, pastTime, config.Period) {
			t.Error("Code should verify with time skew")
		}
	})

	t.Run("different algorithms", func(t *testing.T) {
		algorithms := []TOTPAlgorithm{TOTPAlgorithmSHA1, TOTPAlgorithmSHA256, TOTPAlgorithmSHA512}

		for _, alg := range algorithms {
			config := TOTPConfig{
				Secret:    secret,
				Digits:    6,
				Algorithm: alg,
				Period:    30,
			}

			code, err := generateTOTPWithConfig(config, now, 0)
			if err != nil {
				t.Errorf("generateTOTP() error = %v", err)
			}
			if code == "" {
				t.Errorf("Failed to generate TOTP with algorithm %s", alg)
				continue
			}

			if !VerifyTOTPWithConfig(config, code, now, 0) {
				t.Errorf("Failed to verify TOTP with algorithm %s", alg)
			}
		}
	})

	t.Run("different digits", func(t *testing.T) {
		digits := []int{6, 7, 8}

		for _, d := range digits {
			config := TOTPConfig{
				Secret:    secret,
				Digits:    d,
				Algorithm: TOTPAlgorithmSHA1,
				Period:    30,
			}

			code, err := generateTOTPWithConfig(config, now, 0)
			if err != nil {
				t.Errorf("generateTOTP() error = %v", err)
			}
			if len(code) != d {
				t.Errorf("Expected %d digits, got %d: %s", d, len(code), code)
			}
		}
	})
}
