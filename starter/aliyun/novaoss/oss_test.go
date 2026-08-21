package novaoss

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	aliyunoss "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/luaxlou/nova/starter/internal/registry"
	"github.com/luaxlou/nova/starter/novaconfig"
)

func TestBucketBuildsDefaultInstanceFromNovaConfig(t *testing.T) {
	loadConfig(t, `
oss:
  endpoint: https://oss-cn-hangzhou.aliyuncs.com
  bucket: release-artifacts
  access_key_id: test-access-key-id
  access_key_secret: test-access-key-secret
  security_token: test-security-token
`)
	resetForTest()

	if err := Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
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

func TestNamedUsesConfiguredDefaultAndNamedInstances(t *testing.T) {
	loadConfig(t, `
oss:
  default: reporting
  instances:
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

	if err := Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	defaultBucket, err := Get().Bucket()
	if err != nil {
		t.Fatalf("Get().Bucket() error = %v", err)
	}
	if defaultBucket.BucketName != "reporting-data" {
		t.Fatalf("default bucket = %q, want reporting-data", defaultBucket.BucketName)
	}

	reportsBucket, err := Named("reports").Bucket()
	if err != nil {
		t.Fatalf("Named(reports).Bucket() error = %v", err)
	}
	if reportsBucket.BucketName != "report-archives" {
		t.Fatalf("named bucket = %q, want report-archives", reportsBucket.BucketName)
	}
}

func TestDefaultHelpersInitializeBeforeResolvingRegistryHandle(t *testing.T) {
	config := `
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

func TestExportedLifecyclePropagatesInitializationErrors(t *testing.T) {
	config := `
oss:
  instances: {}
`

	for _, lifecycle := range []struct {
		name string
		call func() error
	}{
		{name: "Reload", call: Reload},
		{name: "Close", call: Close},
	} {
		t.Run(lifecycle.name, func(t *testing.T) {
			loadConfig(t, config)
			resetForTest()

			err := lifecycle.call()
			if err == nil || !strings.Contains(err.Error(), "oss config instances must define at least one named instance") {
				t.Fatalf("%s() error = %v, want empty instances error", lifecycle.name, err)
			}
		})
	}
}

func TestInitRejectsUnknownDefaultInstance(t *testing.T) {
	loadConfig(t, `
oss:
  default: reporting
  instances:
    reports:
      endpoint: https://oss-cn-shanghai.aliyuncs.com
      bucket: report-archives
      access_key_id: reports-access-key-id
      access_key_secret: reports-access-key-secret
`)
	resetForTest()

	err := Init()
	if err == nil || !strings.Contains(err.Error(), `default instance "reporting" is not defined`) {
		t.Fatalf("Init() error = %v, want missing default instance error", err)
	}
}

func TestInitRejectsMalformedNamedInstances(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{
			name: "instances is not a map",
			config: `
oss:
  endpoint: https://oss-cn-hangzhou.aliyuncs.com
  bucket: fallback-bucket
  access_key_id: fallback-access-key-id
  access_key_secret: fallback-access-key-secret
  instances: invalid
`,
			wantErr: "oss config instances must be a map",
		},
		{
			name: "instances is empty",
			config: `
oss:
  instances: {}
`,
			wantErr: "oss config instances must define at least one named instance",
		},
		{
			name: "named instance key is empty",
			config: `
oss:
  instances:
    "":
      endpoint: https://oss-cn-hangzhou.aliyuncs.com
      bucket: unnamed-bucket
      access_key_id: unnamed-access-key-id
      access_key_secret: unnamed-access-key-secret
`,
			wantErr: "oss config instances contains an empty instance name",
		},
		{
			name: "named entry is not a map",
			config: `
oss:
  default: reporting
  instances:
    reporting:
      endpoint: https://oss-cn-hangzhou.aliyuncs.com
      bucket: reporting-bucket
      access_key_id: reporting-access-key-id
      access_key_secret: reporting-access-key-secret
    archive: invalid
`,
			wantErr: "oss config instances.archive must be a map",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loadConfig(t, tt.config)
			resetForTest()

			if err := Init(); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Init() error = %v, want error containing %q", err, tt.wantErr)
			}
			if _, err := Bucket(); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Bucket() error = %v, want error containing %q", err, tt.wantErr)
			}
		})
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

			if err := Init(); err != nil {
				t.Fatalf("Init() error = %v", err)
			}
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
