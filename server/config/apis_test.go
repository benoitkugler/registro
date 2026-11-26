package config

import "testing"

func TestAPIConstructors(t *testing.T) {
	t.Setenv("IMMICH_BASE_URL", "https://example.com")
	t.Setenv("IMMICH_API_KEY", "test-key")
	t.Setenv("STRIPE_KEY", "test-key")
	t.Setenv("STRIPE_WEBHOOK", "test-webhook")
	t.Setenv("PAYREXX_INSTANCE", "test-instance")
	t.Setenv("PAYREXX_API_KEY", "test-key")
	t.Setenv("ANCV_SHOP_ID", "10000065")
	t.Setenv("ANCV_SEAL_KEY", "test-seal-key")
	t.Setenv("ANCV_SEAL_VERSION", "version-test")
	t.Setenv("ANCV_PROD", "TRUE")
	t.Setenv("HELLOASSO_SLUG", "test-slug")
	t.Setenv("HELLOASSO_ID", "test-id")
	t.Setenv("HELLOASSO_SECRET", "test-secret")
	t.Setenv("HELLOASSO_SANDBOX", "TRUE")
	t.Setenv("HELLOASSO_HORIZON_KEY", "test-horizon-key")
	t.Setenv("HELLOASSO_RELAIS_KEY", "test-relais-key")
	t.Setenv("COMPTA_GOOGLE_DRIVE_KEY_FILE", "test-key-file.json")
	t.Setenv("COMPTA_GOOGLE_DRIVE_FOLDER_ID", "test-folder-id")
	t.Setenv("COMPTA_FILES_DIR", "/tmp")
	t.Setenv("COMPTA_ADMIN_PASSWORD", "test-password")
	t.Setenv("COMPTA_ADMIN_MAIL", "admin@example.com")

	tests := []struct {
		name string
		call func() error
	}{
		{name: "NewImmich", call: func() error { _, err := NewImmich(); return err }},
		{name: "NewStripe", call: func() error { _, err := NewStripe(); return err }},
		{name: "NewPayrexx", call: func() error { _, err := NewPayrexx(); return err }},
		{name: "NewANCV", call: func() error { _, err := NewANCV(); return err }},
		{name: "NewHelloasso", call: func() error { _, err := NewHelloasso(); return err }},
		{name: "NewHorizon", call: func() error { _, err := NewHorizon(); return err }},
		{name: "NewCompta", call: func() error { _, err := NewCompta(true); return err }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); err != nil {
				t.Fatalf("constructor returned an error: %v", err)
			}
		})
	}
}

func TestNewANCV(t *testing.T) {
	t.Setenv("ANCV_SHOP_ID", "10000065")
	t.Setenv("ANCV_SEAL_KEY", "663768ff68ad8ea6768bbf65163e9b0a")
	t.Setenv("ANCV_SEAL_VERSION", "version-3620")
	t.Setenv("ANCV_PROD", "TRUE")

	cfg, err := NewANCV()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ShopID != 10000065 || cfg.SealKey == "" || cfg.SealVersion != "version-3620" || !cfg.Prod {
		t.Fatalf("unexpected config: %+v", cfg)
	}

	t.Setenv("ANCV_PROD", "FALSE")
	cfg, err = NewANCV()
	if err != nil || cfg.Prod {
		t.Fatalf("recette should be selected: %+v, %v", cfg, err)
	}

	t.Setenv("ANCV_PROD", "")
	cfg, err = NewANCV()
	if err != nil || cfg.Prod {
		t.Fatalf("unset ANCV_PROD should select recette: %+v, %v", cfg, err)
	}

	t.Setenv("ANCV_SHOP_ID", "invalid")
	if _, err := NewANCV(); err == nil {
		t.Fatal("an invalid shop id should be rejected")
	}
}
