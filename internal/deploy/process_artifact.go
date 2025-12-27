package deploy

import (
	"youfun/shipyard/internal/database"
	"youfun/shipyard/internal/models"
	"youfun/shipyard/pkg/types"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// ProcessArtifact handles the build artifact lifecycle: validation, reuse, or creation.
// It populates the Deployer's metadata fields (Version, GitCommitSHA, tarballPath, md5Hash).
func (d *Deployer) ProcessArtifact() error {
	// 1. Try to reuse explicitly requested build (MD5 or Version)
	if d.useBuild != "" {
		log.Printf("--- Attempting to reuse build artifact: %s ---", d.useBuild)
		if err := d.findAndReuseArtifact(d.useBuild, true); err != nil {
			// Logic matches original: if not found, log and continue.
			// The logging inside findAndReuseArtifact handles the user feedback.
		} else if d.tarballPath != "" {
			return nil
		}
	}

	// 2. Try to reuse based on current Git commit
	gitVersion, err := d.getGitVersion()
	if err != nil {
		log.Printf("⚠️ Failed to get Git version info: %v. Proceeding with build.", err)
		gitVersion = "unknown"
	}

	d.GitCommitSHA = gitVersion

	if d.tarballPath == "" && gitVersion != "unknown" && !strings.HasSuffix(gitVersion, "-dirty") {
		log.Printf("Git version: %s (clean workspace)", gitVersion)
		if err := d.findAndReuseArtifact(gitVersion, false); err != nil {
			// Just log
		} else if d.tarballPath != "" {
			return nil
		}
	} else if strings.HasSuffix(gitVersion, "-dirty") {
		log.Printf("Git version: %s (uncommitted changes in workspace)", gitVersion)
	}

	// 3. Perform a new build if no artifact has been reused
	if d.tarballPath == "" {
		log.Println("--- Performing new build ---")
		return d.performNewBuild(gitVersion)
	}

	return nil
}

// findAndReuseArtifact attempts to find and validate an existing artifact by identifier (MD5, Version or GitSHA).
func (d *Deployer) findAndReuseArtifact(query string, isExplicit bool) error {
	var version, tarballPath, md5Hash, gitSha string

	if d.APIClient != nil {
		// API Mode
		artifact, artErr := d.APIClient.CheckArtifact(d.Application.ID.String(), query)
		if artErr == nil && artifact != nil {
			version = artifact.Version
			tarballPath = artifact.LocalPath
			md5Hash = artifact.MD5Hash
			gitSha = artifact.GitCommitSHA
		} else if artErr != nil {
			// Log checking error if needed, but we essentially proceed to not found
			// log.Printf("Debug: API artifact check error: %v", artErr)
		}
	} else {
		// DB Mode
		if isExplicit {
			buildArtifact, dbErr := database.GetBuildArtifactByMD5(d.Application.ID, query)
			if dbErr != nil {
				buildArtifact, dbErr = database.GetLatestBuildArtifactByVersion(d.Application.ID, query)
			}
			if dbErr == nil && buildArtifact != nil {
				version = buildArtifact.Version
				tarballPath = buildArtifact.LocalPath
				md5Hash = buildArtifact.MD5Hash
				gitSha = buildArtifact.GitCommitSHA
			}
		} else {
			// Implicit by Git SHA
			buildArtifact, dbErr := database.GetBuildArtifactByGitSHA(d.Application.ID, query)
			if dbErr == nil && buildArtifact != nil {
				version = buildArtifact.Version
				tarballPath = buildArtifact.LocalPath
				md5Hash = buildArtifact.MD5Hash
				gitSha = buildArtifact.GitCommitSHA
			}
		}
	}

	if tarballPath != "" {
		// Validate Local File
		actualMD5, md5Err := calculateMD5(tarballPath)
		if md5Err == nil && actualMD5 == md5Hash {
			log.Printf("✅ Found and reusing build artifact (Version: %s, MD5: %s, Git: %s)", version, md5Hash, gitSha)
			d.Version = version
			d.tarballPath = tarballPath
			d.md5Hash = md5Hash
			d.GitCommitSHA = gitSha
			return nil
		}
		log.Printf("⚠️ Cached build artifact '%s' is corrupted or MD5 mismatch.", tarballPath)
	} else {
		if isExplicit {
			log.Printf("⚠️ No matching build artifact found for '%s'", query)
		}
	}
	return fmt.Errorf("artifact not found or invalid")
}

// performNewBuild handles the creation of a new build artifact.
func (d *Deployer) performNewBuild(gitVersion string) error {
	var version string
	var err error

	// Determine version
	if d.Runtime == "static" {
		version, err = d.getVersionForStatic()
	} else {
		version, err = d.getVersionFromMix()
	}
	if err != nil {
		return err
	}
	log.Printf("Got project version: %s", version)

	// Build
	var buildDir string
	if d.Runtime == "static" {
		buildDir = d.buildStaticRelease()
	} else {
		buildDir = d.buildRelease()
	}
	// Note: buildDir is a temp dir that contains the release structure
	defer os.RemoveAll(buildDir)

	// Create Tarball
	tempTarballPath, err := d.createTarball(filepath.Join(buildDir, "release"), d.AppName)
	if err != nil {
		return err
	}
	defer os.Remove(tempTarballPath)

	// Calculate MD5
	md5Hash, err := calculateMD5(tempTarballPath)
	if err != nil {
		return fmt.Errorf("MD5 calculation failed: %w", err)
	}

	// Move to Cache
	buildCacheDir, err := database.GetBuildCacheDir()
	if err != nil {
		return fmt.Errorf("failed to get cache directory: %w", err)
	}
	cachedTarballPath := path.Join(buildCacheDir, fmt.Sprintf("%s-%s.tar.gz", d.AppName, md5Hash))

	if err := moveFile(tempTarballPath, cachedTarballPath); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	// Register Metadata
	if d.APIClient != nil {
		artifactDTO := &types.BuildArtifactDTO{
			ApplicationID: d.Application.ID.String(),
			Version:       version,
			GitCommitSHA:  gitVersion,
			MD5Hash:       md5Hash,
			LocalPath:     cachedTarballPath,
		}
		if err := d.APIClient.RegisterArtifact(artifactDTO); err != nil {
			log.Printf("⚠️ Warning: Failed to register build artifact to API: %v", err)
		}
	} else {
		// DB Mode
		now := time.Now()
		artifact := &models.BuildArtifact{
			ApplicationID: d.Application.ID,
			Version:       version,
			GitCommitSHA:  gitVersion,
			MD5Hash:       md5Hash,
			LocalPath:     cachedTarballPath,
			CreatedAt:     models.NullableTime{Time: &now},
		}
		if err := database.AddBuildArtifact(artifact); err != nil {
			log.Printf("⚠️ Warning: Failed to save build artifact metadata: %v", err)
		} else {
			log.Printf("✅ Build artifact cached (Version: %s, Git: %s, MD5: %s)", version, gitVersion, md5Hash)
		}
	}

	// Update Deployer State
	d.Version = version
	d.tarballPath = cachedTarballPath
	d.md5Hash = md5Hash
	// d.GitCommitSHA is typically updated to match the build env, which is passed in.

	return nil
}
