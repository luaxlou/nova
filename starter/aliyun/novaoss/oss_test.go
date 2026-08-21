package novaoss

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/luaxlou/nova/internal/registry"
	"github.com/luaxlou/nova/starter/config/novaconfig"
)

func TestBucketBuildsSingleInstanceFromNovaConfig(t *testing.T) {
	loadConfig(t, `
aliyun:
  oss:
    endpoint: https://oss-cn-hangzhou.aliyuncs.com
    bucket: release-artifacts
    access_key_id: test-access-key-id
    access_key_secret: test-access-key-secret
    security_token: test-security-token
`)
	resetForTest()

	if err := initFromConfig(); err != nil {
		t.Fatalf("initFromConfig() error = %v", err)
	}

	bucket, err := Bucket()
	if err != nil {
		t.Fatalf("Bucket() error = %v", err)
	}
	if bucket.BucketName != "release-artifacts" {
		t.Fatalf("bucket name = %q, want release-artifacts", bucket.BucketName)
	}
	if bucket.Client.Config.Endpoint != "https://oss-cn-hangzhou.aliyuncs.com" {
		t.Fatalf("endpoint = %q, want configured endpoint", bucket.Client.Config.Endpoint)
	}
	if bucket.Client.Config.SecurityToken != "test-security-token" {
		t.Fatalf("security token was not applied")
	}
}

func TestNamedUsesConfiguredInstances(t *testing.T) {
	loadConfig(t, `
aliyun:
  oss:
    reports:
        endpoint: https://oss-cn-shanghai.aliyuncs.com
        bucket: report-archives
        access_key_id: reports-access-key-id
        access_key_secret: reports-access-key-secret
    reporting:
        endpoint: https://oss-cn-beijing.aliyuncs.com
        bucket: reporting-data
        access_key_id: reporting-access-key-id
        access_key_secret: reporting-access-key-secret
`)
	resetForTest()

	if err := initFromConfig(); err != nil {
		t.Fatalf("initFromConfig() error = %v", err)
	}

	reportsBucket, err := Named("reports").Bucket()
	if err != nil {
		t.Fatalf("Named(reports).Bucket() error = %v", err)
	}
	if reportsBucket.BucketName != "report-archives" {
		t.Fatalf("named bucket = %q, want report-archives", reportsBucket.BucketName)
	}

	reportingBucket, err := Named("reporting").Bucket()
	if err != nil {
		t.Fatalf("Named(reporting).Bucket() error = %v", err)
	}
	if reportingBucket.BucketName != "reporting-data" {
		t.Fatalf("named bucket = %q, want reporting-data", reportingBucket.BucketName)
	}
}

func TestBucketRequiresNameWhenMultipleInstancesConfigured(t *testing.T) {
	loadConfig(t, `
aliyun:
  oss:
    reports:
        endpoint: https://oss-cn-shanghai.aliyuncs.com
        bucket: report-archives
        access_key_id: reports-access-key-id
        access_key_secret: reports-access-key-secret
    reporting:
        endpoint: https://oss-cn-beijing.aliyuncs.com
        bucket: reporting-data
        access_key_id: reporting-access-key-id
        access_key_secret: reporting-access-key-secret
`)
	resetForTest()

	if _, err := Bucket(); err == nil || !strings.Contains(err.Error(), "oss instance name is required") {
		t.Fatalf("Bucket() error = %v, want instance name required error", err)
	}
}

func TestDefaultHelpersInitializeBeforeResolvingRegistryHandle(t *testing.T) {
	config := `
aliyun:
  oss:
    endpoint: https://oss-cn-hangzhou.aliyuncs.com
    bucket: release-artifacts
    access_key_id: test-access-key-id
    access_key_secret: test-access-key-secret
`

	t.Run("Bucket", func(t *testing.T) {
		loadConfig(t, config)
		resetForTest()

		bucket, err := Bucket()
		if err != nil {
			t.Fatalf("Bucket() error = %v", err)
		}
		if bucket.BucketName != "release-artifacts" {
			t.Fatalf("bucket name = %q, want release-artifacts", bucket.BucketName)
		}
	})

	t.Run("Named empty", func(t *testing.T) {
		loadConfig(t, config)
		resetForTest()

		bucket, err := Named("").Bucket()
		if err != nil {
			t.Fatalf("Named(\"\").Bucket() error = %v", err)
		}
		if bucket.BucketName != "release-artifacts" {
			t.Fatalf("bucket name = %q, want release-artifacts", bucket.BucketName)
		}
	})

	t.Run("Reload", func(t *testing.T) {
		loadConfig(t, config)
		resetForTest()

		if err := Reload(); err != nil {
			t.Fatalf("Reload() error = %v", err)
		}
	})

	t.Run("Close", func(t *testing.T) {
		loadConfig(t, config)
		resetForTest()

		if err := Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
		if _, err := Bucket(); err != nil {
			t.Fatalf("Bucket() after Close() error = %v", err)
		}
	})
}

func TestReloadRebuildsDefinitionsFromCurrentNovaConfig(t *testing.T) {
	loadConfig(t, `
aliyun:
  oss:
    endpoint: https://oss-cn-hangzhou.aliyuncs.com
    bucket: release-artifacts
    access_key_id: first-access-key-id
    access_key_secret: first-access-key-secret
`)
	resetForTest()

	handle := Named("")
	firstBucket, err := handle.Bucket()
	if err != nil {
		t.Fatalf("first Bucket() error = %v", err)
	}
	if firstBucket.BucketName != "release-artifacts" {
		t.Fatalf("first bucket name = %q, want release-artifacts", firstBucket.BucketName)
	}

	loadConfig(t, `
aliyun:
  oss:
    endpoint: https://oss-cn-shanghai.aliyuncs.com
    bucket: compliance-reports
    access_key_id: second-access-key-id
    access_key_secret: second-access-key-secret
`)

	if err := Reload(); err != nil {
		t.Fatalf("Reload() error = %v", err)
	}

	reloadedBucket, err := handle.Bucket()
	if err != nil {
		t.Fatalf("Bucket() after Reload() error = %v", err)
	}
	if reloadedBucket.BucketName != "compliance-reports" {
		t.Fatalf("reloaded bucket name = %q, want compliance-reports", reloadedBucket.BucketName)
	}
	if reloadedBucket.Client.Config.Endpoint != "https://oss-cn-shanghai.aliyuncs.com" {
		t.Fatalf("reloaded endpoint = %q, want https://oss-cn-shanghai.aliyuncs.com", reloadedBucket.Client.Config.Endpoint)
	}
}

func TestReloadInvalidConfigClearsCachedBucket(t *testing.T) {
	loadConfig(t, `
aliyun:
  oss:
    endpoint: https://oss-cn-hangzhou.aliyuncs.com
    bucket: release-artifacts
    access_key_id: first-access-key-id
    access_key_secret: first-access-key-secret
`)
	resetForTest()

	handle := Named("")
	if _, err := handle.Bucket(); err != nil {
		t.Fatalf("first Bucket() error = %v", err)
	}

	loadConfig(t, `
aliyun:
  oss:
    endpoint: https://oss-cn-shanghai.aliyuncs.com
    bucket: compliance-reports
    access_key_id: second-access-key-id
`)

	err := Reload()
	if err == nil || !strings.Contains(err.Error(), "reload aliyun.oss config") || !strings.Contains(err.Error(), "missing access_key_secret") {
		t.Fatalf("Reload() error = %v, want contextual missing access_key_secret error", err)
	}

	if _, err := handle.Bucket(); err == nil || !strings.Contains(err.Error(), "missing access_key_secret") {
		t.Fatalf("Bucket() after invalid Reload() error = %v, want missing access_key_secret error", err)
	}
}

func TestExportedLifecyclePropagatesInitializationErrors(t *testing.T) {
	config := `
aliyun:
  oss: {}
`

	loadConfig(t, config)
	resetForTest()

	err := Reload()
	if err == nil || !strings.Contains(err.Error(), "missing endpoint") {
		t.Fatalf("Reload() error = %v, want missing endpoint error", err)
	}
}

func TestBucketRejectsIncompleteConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{
			name: "endpoint",
			config: `
aliyun:
  oss:
    bucket: release-artifacts
    access_key_id: test-access-key-id
    access_key_secret: test-access-key-secret
`,
			wantErr: "missing endpoint",
		},
		{
			name: "bucket",
			config: `
aliyun:
  oss:
    endpoint: https://oss-cn-hangzhou.aliyuncs.com
    access_key_id: test-access-key-id
    access_key_secret: test-access-key-secret
`,
			wantErr: "missing bucket",
		},
		{
			name: "access key id",
			config: `
aliyun:
  oss:
    endpoint: https://oss-cn-hangzhou.aliyuncs.com
    bucket: release-artifacts
    access_key_secret: test-access-key-secret
`,
			wantErr: "missing access_key_id",
		},
		{
			name: "access key secret",
			config: `
aliyun:
  oss:
    endpoint: https://oss-cn-hangzhou.aliyuncs.com
    bucket: release-artifacts
    access_key_id: test-access-key-id
`,
			wantErr: "missing access_key_secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loadConfig(t, tt.config)
			resetForTest()

			if _, err := Bucket(); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Bucket() error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func loadConfig(t *testing.T, contents string) {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	novaconfig.SetConfigPath(configPath)
	if err := novaconfig.Reload(); err != nil {
		t.Fatalf("reload config: %v", err)
	}
}

func resetForTest() {
	initialized = false
	reg = registry.New[*aliyunoss.Bucket]()
}
