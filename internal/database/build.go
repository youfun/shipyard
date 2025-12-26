package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"youfun/shipyard/internal/models"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// --- build_artifacts Table Operations ---

// AddBuildArtifact adds a build artifact record.
func AddBuildArtifact(artifact *models.BuildArtifact) error {
	artifact.ID = uuid.New()
	now := time.Now()
	artifact.CreatedAt = models.NullableTime{Time: &now}
	query := `INSERT INTO build_artifacts (id, application_id, version, git_commit_sha, md5_hash, local_path, created_at) VALUES (:id, :application_id, :version, :git_commit_sha, :md5_hash, :local_path, :created_at)`
	_, err := DB.NamedExec(query, artifact)
	return err
}

// GetBuildArtifactByMD5 retrieves a build artifact by MD5 checksum.
func GetBuildArtifactByMD5(appID uuid.UUID, md5Hash string) (*models.BuildArtifact, error) {
	var artifact models.BuildArtifact
	query := Rebind(`SELECT * FROM build_artifacts WHERE application_id = ? AND md5_hash = ?`)
	err := DB.Get(&artifact, query, appID, md5Hash)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no build artifact found with MD5 '%s'", md5Hash)
	}
	return &artifact, err
}

// GetBuildArtifactByMD5Prefix retrieves a build artifact by MD5 prefix (supports short MD5 hashes).
// Returns error if no match or multiple matches found.
func GetBuildArtifactByMD5Prefix(appID uuid.UUID, md5Prefix string) (*models.BuildArtifact, error) {
	var artifacts []models.BuildArtifact
	query := Rebind(`SELECT * FROM build_artifacts WHERE application_id = ? AND md5_hash LIKE ? ORDER BY created_at DESC`)
	err := DB.Select(&artifacts, query, appID, md5Prefix+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query build artifacts: %w", err)
	}
	if len(artifacts) == 0 {
		return nil, fmt.Errorf("no build artifact found with MD5 prefix '%s'", md5Prefix)
	}
	if len(artifacts) > 1 {
		return nil, fmt.Errorf("multiple build artifacts found with MD5 prefix '%s', please use a longer prefix", md5Prefix)
	}
	return &artifacts[0], nil
}

// GetLatestBuildArtifactByVersion retrieves the latest build artifact by version number.
func GetLatestBuildArtifactByVersion(appID uuid.UUID, version string) (*models.BuildArtifact, error) {
	var artifact models.BuildArtifact
	query := Rebind(`SELECT * FROM build_artifacts WHERE application_id = ? AND version = ? ORDER BY created_at DESC LIMIT 1`)
	err := DB.Get(&artifact, query, appID, version)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no build artifact found for version '%s'", version)
	}
	return &artifact, err
}

// GetBuildArtifactByGitSHA retrieves the latest build artifact for a given app and git commit SHA.
func GetBuildArtifactByGitSHA(appID uuid.UUID, gitSHA string) (*models.BuildArtifact, error) {
	var artifact models.BuildArtifact
	query := Rebind(`SELECT * FROM build_artifacts WHERE application_id = ? AND git_commit_sha = ? ORDER BY created_at DESC LIMIT 1`)
	err := DB.Get(&artifact, query, appID, gitSHA)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no build artifact found for Git SHA '%s'", gitSHA)
	}
	return &artifact, err
}

// GetAllBuildArtifactsForApp retrieves all build artifacts for an application.
func GetAllBuildArtifactsForApp(appID uuid.UUID) ([]models.BuildArtifact, error) {
	var artifacts []models.BuildArtifact
	query := Rebind(`SELECT * FROM build_artifacts WHERE application_id = ? ORDER BY created_at DESC`)
	err := DB.Select(&artifacts, query, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to query build artifacts for application '%s': %w", appID, err)
	}
	return artifacts, nil
}

func GetBuildCacheDir() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	buildCacheDir := filepath.Join(configDir, "build_cache")
	if err := os.MkdirAll(buildCacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create build cache directory: %w", err)
	}
	return buildCacheDir, nil
}
