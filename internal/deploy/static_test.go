package deploy

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIsStaticProject(t *testing.T) {
	// Save current working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	tests := []struct {
		name     string
		setup    func(dir string) error
		expected bool
	}{
		{
			name: "index.html in root",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html></html>"), 0644)
			},
			expected: true,
		},
		{
			name: "dist/index.html",
			setup: func(dir string) error {
				distDir := filepath.Join(dir, "dist")
				if err := os.MkdirAll(distDir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(distDir, "index.html"), []byte("<html></html>"), 0644)
			},
			expected: true,
		},
		{
			name: "build/index.html",
			setup: func(dir string) error {
				buildDir := filepath.Join(dir, "build")
				if err := os.MkdirAll(buildDir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(buildDir, "index.html"), []byte("<html></html>"), 0644)
			},
			expected: true,
		},
		{
			name: "public/index.html",
			setup: func(dir string) error {
				publicDir := filepath.Join(dir, "public")
				if err := os.MkdirAll(publicDir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(publicDir, "index.html"), []byte("<html></html>"), 0644)
			},
			expected: true,
		},
		{
			name:     "no static files",
			setup:    func(dir string) error { return nil },
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "deployer-test-")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			// Change to temp directory
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Setup test environment
			if err := tt.setup(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Test
			result := isStaticProject()
			if result != tt.expected {
				t.Errorf("isStaticProject() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFindStaticSourceDir(t *testing.T) {
	// Save current working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	tests := []struct {
		name     string
		setup    func(dir string) error
		expected string
	}{
		{
			name: "prefers dist over root",
			setup: func(dir string) error {
				// Create both dist/index.html and root index.html
				distDir := filepath.Join(dir, "dist")
				if err := os.MkdirAll(distDir, 0755); err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(distDir, "index.html"), []byte("<html></html>"), 0644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html></html>"), 0644)
			},
			expected: "dist",
		},
		{
			name: "uses build when dist not available",
			setup: func(dir string) error {
				buildDir := filepath.Join(dir, "build")
				if err := os.MkdirAll(buildDir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(buildDir, "index.html"), []byte("<html></html>"), 0644)
			},
			expected: "build",
		},
		{
			name: "falls back to root",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html></html>"), 0644)
			},
			expected: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "deployer-test-")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			// Change to temp directory
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Setup test environment
			if err := tt.setup(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Test
			d := &Deployer{}
			result := d.findStaticSourceDir()
			if result != tt.expected {
				t.Errorf("findStaticSourceDir() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetVersionForStatic(t *testing.T) {
	// Save current working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	tests := []struct {
		name            string
		setup           func(dir string) error
		expectedVersion string
		expectTimestamp bool
	}{
		{
			name: "reads version from package.json",
			setup: func(dir string) error {
				content := `{"name": "test-app", "version": "1.2.3"}`
				return os.WriteFile(filepath.Join(dir, "package.json"), []byte(content), 0644)
			},
			expectedVersion: "1.2.3",
			expectTimestamp: false,
		},
		{
			name:            "falls back to timestamp without package.json",
			setup:           func(dir string) error { return nil },
			expectTimestamp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "deployer-test-")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			// Change to temp directory
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Setup test environment
			if err := tt.setup(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Test
			d := &Deployer{}
			version, err := d.getVersionForStatic()
			if err != nil {
				t.Errorf("getVersionForStatic() error = %v", err)
				return
			}

			if tt.expectTimestamp {
				// Check that it looks like a timestamp (YYYYMMDD.HHMMSS format)
				if len(version) != 15 || version[8] != '.' {
					t.Errorf("getVersionForStatic() = %v, want timestamp format YYYYMMDD.HHMMSS", version)
				}
			} else if version != tt.expectedVersion {
				t.Errorf("getVersionForStatic() = %v, want %v", version, tt.expectedVersion)
			}
		})
	}
}

func TestDetectRuntime_Static(t *testing.T) {
	// Save current working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	// Create temp directory with just index.html (no mix.exs, package.json, go.mod)
	tmpDir, err := os.MkdirTemp("", "deployer-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Create index.html
	if err := os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte("<html></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test
	d := &Deployer{}
	runtime := d.detectRuntime()
	if runtime != "static" {
		t.Errorf("detectRuntime() = %v, want static", runtime)
	}
}

func TestBuildStaticRelease(t *testing.T) {
	// Skip this test if Docker is not available or in CI environment
	// The Docker build requires network access which may not be available
	if os.Getenv("CI") != "" {
		t.Skip("Skipping Docker-based test in CI environment")
	}

	// Check if Docker is available
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		t.Skip("Docker is not available, skipping build test")
	}

	// Save current working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	// Create temp directory with static files
	tmpDir, err := os.MkdirTemp("", "deployer-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Create static files structure
	cssDir := filepath.Join(tmpDir, "css")
	jsDir := filepath.Join(tmpDir, "js")
	if err := os.MkdirAll(cssDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create sample files
	if err := os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte("<html><head></head><body>Hello</body></html>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cssDir, "style.css"), []byte("body { margin: 0; }"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(jsDir, "app.js"), []byte("console.log('Hello');"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test buildStaticRelease
	d := &Deployer{}
	buildDir := d.buildStaticRelease()
	defer os.RemoveAll(buildDir)

	// Verify that the release directory was created with correct structure
	releaseDir := filepath.Join(buildDir, "release")

	// For Docker multi-stage build, we expect a 'server' binary
	if _, err := os.Stat(filepath.Join(releaseDir, "server")); err != nil {
		t.Errorf("server binary not found in release directory: %v", err)
	}
}

func TestCopyStaticFiles_SkipsHiddenFiles(t *testing.T) {
	// Save current working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	// Create temp directories
	srcDir, err := os.MkdirTemp("", "deployer-src-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(srcDir)

	dstDir, err := os.MkdirTemp("", "deployer-dst-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dstDir)

	// Create files including hidden ones
	if err := os.WriteFile(filepath.Join(srcDir, "index.html"), []byte("<html></html>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, ".gitignore"), []byte("node_modules/"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create hidden directory
	hiddenDir := filepath.Join(srcDir, ".git")
	if err := os.MkdirAll(hiddenDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "config"), []byte("git config"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test
	d := &Deployer{}
	if err := d.copyStaticFiles(srcDir, dstDir); err != nil {
		t.Fatalf("copyStaticFiles failed: %v", err)
	}

	// Verify index.html was copied
	if _, err := os.Stat(filepath.Join(dstDir, "index.html")); err != nil {
		t.Error("index.html should have been copied")
	}

	// Verify .gitignore was NOT copied
	if _, err := os.Stat(filepath.Join(dstDir, ".gitignore")); !os.IsNotExist(err) {
		t.Error(".gitignore should NOT have been copied")
	}

	// Verify .git directory was NOT copied
	if _, err := os.Stat(filepath.Join(dstDir, ".git")); !os.IsNotExist(err) {
		t.Error(".git directory should NOT have been copied")
	}
}
