package deploy

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"youfun/shipyard/internal/static"
	"strings"
	"time"
)

// Regex patterns compiled once for performance
var (
	versionFromMixRegex    = regexp.MustCompile(`version:\s*"(.*?)"`)
	appNameFromMixRegex    = regexp.MustCompile(`app:\s*:(\w+)`)
	versionFromPackageJSON = regexp.MustCompile(`"version"\s*:\s*"([^"]+)"`)
)

func (d *Deployer) buildRelease() string {
	log.Println("Building using Docker...")

	dockerfilePath := "Dockerfile.shipyard"

	// Check if the agreed Dockerfile exists in project root
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		// If not exists, write built-in Dockerfile to current directory based on runtime
		log.Printf("Did not find '%s' in project root, using built-in preset Dockerfile and outputting to project root", dockerfilePath)

		// Choose Dockerfile based on runtime type
		var dockerfileContent string
		if d.Runtime == "elixir" {
			log.Println("Using pure Elixir Dockerfile (without Phoenix assets)")
			dockerfileContent = static.DockerfileBuildElixir
		} else {
			// Default to Phoenix Dockerfile (phoenix runtime or fallback)
			log.Println("Using Phoenix Dockerfile (with assets build)")
			dockerfileContent = static.DockerfileBuildPhoenix
		}

		err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644)
		if err != nil {
			log.Fatalf("failed to write preset Dockerfile: %v", err)
		}
		// Note: We do not delete it here so user can see and modify it after build
	} else {
		log.Printf("Detected '%s' in project root, using this file for build.", dockerfilePath)
	}

	// Get app name from mix.exs
	appName, err := d.getAppNameFromMix()
	if err != nil {
		log.Fatalf("failed to get app name from mix.exs: %v", err)
	}
	log.Printf("Got app name from mix.exs: %s", appName)

	// For pure Elixir projects, ensure priv directory exists (even if empty)
	// This is because Docker COPY will fail if the source doesn't exist
	if d.Runtime == "elixir" {
		if _, err := os.Stat("priv"); os.IsNotExist(err) {
			log.Println("Creating empty priv directory for Elixir project...")
			if err := os.MkdirAll("priv", 0755); err != nil {
				log.Printf("Warning: Failed to create priv directory: %v", err)
			}
		}
	}

	// Create a temp directory to store build artifacts
	buildOutputDir, err := os.MkdirTemp("", "deployer-build-")
	if err != nil {
		log.Fatalf("failed to create temp build directory: %v", err)
	}
	log.Printf("Build artifacts will be output to: %s", buildOutputDir)

	// Execute docker build
	cmd := exec.Command("docker", "build", "--output", fmt.Sprintf("type=local,dest=%s", buildOutputDir), "-f", dockerfilePath, "--build-arg", fmt.Sprintf("APP_NAME=%s", appName), ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Docker build failed, please check docker status or build file: %v", err)
	}

	log.Println("✅ Docker build version success.")
	return buildOutputDir
}

func (d *Deployer) createTarball(source, prefix string) (string, error) {
	tarballPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%d.tar.gz", prefix, time.Now().Unix()))
	tarballFile, err := os.Create(tarballPath)
	if err != nil {
		return "", err
	}
	defer tarballFile.Close()

	gzipWriter := gzip.NewWriter(tarballFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return tarballPath, filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(source, path)
		header.Name = strings.ReplaceAll(relPath, "\\", "/")

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if !info.IsDir() {
			file, _ := os.Open(path)
			defer file.Close()
			_, _ = io.Copy(tarWriter, file)
		}
		return nil
	})
}

func (d *Deployer) getVersionFromMix() (string, error) {
	content, err := os.ReadFile("mix.exs")
	if err != nil {
		log.Printf("DEBUG: Failed to read mix.exs file: %v", err)
		return "", fmt.Errorf("failed to read mix.exs: %w", err)
	}
	// log.Printf("DEBUG: Read mix.exs content:\n%s", string(content))
	matches := versionFromMixRegex.FindStringSubmatch(string(content))
	log.Printf("DEBUG: Regex match result: %v (Length: %d)", matches, len(matches))
	if len(matches) < 2 {
		return "", fmt.Errorf("version not found in mix.exs")
	}
	return matches[1], nil
}

func (d *Deployer) getAppNameFromMix() (string, error) {
	content, err := os.ReadFile("mix.exs")
	if err != nil {
		return "", fmt.Errorf("failed to read mix.exs: %w", err)
	}
	matches := appNameFromMixRegex.FindStringSubmatch(string(content))
	if len(matches) < 2 {
		return "", fmt.Errorf("app name not found in mix.exs")
	}
	return matches[1], nil
}

func (d *Deployer) getGitVersion() (string, error) {
	// Check if it's a git repository
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return "", fmt.Errorf("not a git repository")
	}

	// Get the latest commit hash
	hashCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	hashBytes, err := hashCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git commit hash: %w", err)
	}
	hash := strings.TrimSpace(string(hashBytes))

	// Check for dirty working tree
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusBytes, err := statusCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git status: %w", err)
	}

	if len(statusBytes) > 0 {
		log.Println("⚠️ Warning: Uncommitted changes in current Git workspace, build artifact will be marked as '-dirty'.")
		return hash + "-dirty", nil
	}

	return hash, nil
}

// buildStaticRelease builds a static site into a single Go binary using Docker multi-stage build.
// For projects with package.json: Node.js builds frontend -> Go embeds static files
// For pure static files: Go directly embeds static files
// The output is a single Linux binary that serves static files with SPA routing support.
//
// Requirements:
// - Docker with BuildKit enabled (Docker 18.09+ with DOCKER_BUILDKIT=1 or Docker 23.0+ by default)
// - The --output flag requires Docker BuildKit support
func (d *Deployer) buildStaticRelease() string {
	log.Println("Using Docker multi-stage build for static site...")

	// Determine which Dockerfile to use based on project type
	dockerfilePath := "Dockerfile.shipyard.static"
	var dockerfileContent string

	// Check if project has package.json (needs frontend build)
	if _, err := os.Stat("package.json"); err == nil {
		log.Println("Detected package.json, will use Node.js to build frontend then embed into Go binary")
		dockerfileContent = static.DockerfileStatic
	} else {
		log.Println("package.json not detected, will embed static files directly into Go binary")
		dockerfileContent = static.DockerfileStaticSimple
	}

	// Check if user has their own Dockerfile.shipyard.static
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		log.Printf("Did not find '%s' in project root, using built-in preset Dockerfile and outputting to project root", dockerfilePath)
		err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644)
		if err != nil {
			log.Fatalf("failed to write preset Dockerfile: %v", err)
		}
	} else {
		log.Printf("Detected '%s' in project root, using this file for build.", dockerfilePath)
	}

	// Create a temporary directory to store build output
	buildOutputDir, err := os.MkdirTemp("", "deployer-static-build-")
	if err != nil {
		log.Fatalf("failed to create temp build directory: %v", err)
	}
	log.Printf("Build artifacts will be output to: %s", buildOutputDir)

	// Prepare docker build command
	// Note: --output flag requires Docker BuildKit (Docker 18.09+ with DOCKER_BUILDKIT=1 or Docker 23.0+ by default)
	args := []string{"build", "--output", fmt.Sprintf("type=local,dest=%s", buildOutputDir), "-f", dockerfilePath}

	// For simple static (no package.json), pass STATIC_DIR build arg
	if _, err := os.Stat("package.json"); os.IsNotExist(err) {
		staticDir := d.findStaticSourceDir()
		args = append(args, "--build-arg", fmt.Sprintf("STATIC_DIR=%s", staticDir))
	}

	args = append(args, ".")

	// Execute docker build
	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Docker build failed, please check Docker status or build file: %v", err)
	}

	log.Println("✅ Docker multi-stage build completed, single file Linux binary generated.")
	return buildOutputDir
}

// findStaticSourceDir determines the source directory for static files.
// Priority: dist/ > build/ > public/ > root (with index.html)
func (d *Deployer) findStaticSourceDir() string {
	candidates := []string{"dist", "build", "public"}
	for _, dir := range candidates {
		indexPath := filepath.Join(dir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			return dir
		}
	}
	// Check if index.html exists in root
	if _, err := os.Stat("index.html"); err == nil {
		return "."
	}
	log.Fatal("Static file directory not found. Please ensure index.html or dist/build/public directory exists.")
	return "" // unreachable, but required for Go compiler
}

// copyStaticFiles copies all static files from source to destination directory.
func (d *Deployer) copyStaticFiles(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path first
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip hidden files and directories (like .git), but not the root directory "."
		if strings.HasPrefix(info.Name(), ".") && relPath != "." {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		return copyFile(path, dstPath)
	})
}

// getVersionForStatic returns a version string for static projects.
// It tries to read from package.json first, then falls back to timestamp-based version.
func (d *Deployer) getVersionForStatic() (string, error) {
	// Try to get version from package.json if it exists
	if content, err := os.ReadFile("package.json"); err == nil {
		matches := versionFromPackageJSON.FindStringSubmatch(string(content))
		if len(matches) >= 2 {
			return matches[1], nil
		}
	}

	// Fallback to timestamp-based version
	return time.Now().Format("20060102.150405"), nil
}
