package jid

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type workflowDoc struct {
	Name        string                 `yaml:"name"`
	On          yaml.Node              `yaml:"on"`
	Permissions yaml.Node              `yaml:"permissions"`
	Jobs        map[string]workflowJob `yaml:"jobs"`
}

type workflowJob struct {
	Name        string              `yaml:"name"`
	RunsOn      string              `yaml:"runs-on"`
	Permissions yaml.Node           `yaml:"permissions"`
	Steps       []map[string]any    `yaml:"steps"`
}

func TestVirusTotalWorkflowStructure(t *testing.T) {
	path := filepath.Join(".github", "workflows", "virustotal.yml")
	content, err := os.ReadFile(path)
	require.NoError(t, err, "virustotal.yml must exist")
	require.NotEmpty(t, content, "virustotal.yml must not be empty")

	var doc workflowDoc
	err = yaml.Unmarshal(content, &doc)
	require.NoError(t, err, "virustotal.yml must be valid YAML")

	assert.Equal(t, "VirusTotal", doc.Name)

	raw := string(content)
	assert.True(t, strings.Contains(raw, "release:"), "must trigger on release event")
	assert.True(t, strings.Contains(raw, "types: [published]"), "must trigger on published release")
	assert.True(t, strings.Contains(raw, "workflow_dispatch:"), "must support workflow_dispatch")

	job, exists := doc.Jobs["virustotal"]
	require.True(t, exists, "virustotal job must exist")
	assert.Equal(t, "ubuntu-latest", job.RunsOn)

	var hasKeyCheck, hasVTAction bool
	for _, step := range job.Steps {
		name, _ := step["name"].(string)
		uses, _ := step["uses"].(string)

		if strings.Contains(strings.ToLower(name), "check virustotal api key") {
			hasKeyCheck = true
		}
		if strings.HasPrefix(uses, "cssnr/virustotal-action") {
			hasVTAction = true
			with, ok := step["with"].(map[string]any)
			require.True(t, ok, "virustotal-action step must specify inputs")
			assert.Contains(t, with, "vt_api_key")
			assert.Contains(t, with, "github_token")
			assert.Contains(t, with, "release_id")
		}
	}

	assert.True(t, hasKeyCheck, "must include key check step for graceful handling when VT_API_KEY is not set")
	assert.True(t, hasVTAction, "must use cssnr/virustotal-action")
}

func TestReleaseWorkflowWithVirusTotal(t *testing.T) {
	path := filepath.Join(".github", "workflows", "release.yml")
	content, err := os.ReadFile(path)
	require.NoError(t, err, "release.yml must exist")

	var doc workflowDoc
	err = yaml.Unmarshal(content, &doc)
	require.NoError(t, err, "release.yml must be valid YAML")

	assert.Equal(t, "Release", doc.Name)

	job, exists := doc.Jobs["release"]
	require.True(t, exists, "release job must exist")
	assert.Equal(t, "ubuntu-latest", job.RunsOn)

	var hasGoReleaser, hasVTAction bool
	for _, step := range job.Steps {
		uses, _ := step["uses"].(string)

		if strings.HasPrefix(uses, "goreleaser/goreleaser-action") {
			hasGoReleaser = true
		}
		if strings.HasPrefix(uses, "cssnr/virustotal-action") {
			hasVTAction = true
			with, ok := step["with"].(map[string]any)
			require.True(t, ok, "virustotal-action step must specify inputs")
			assert.Contains(t, with, "vt_api_key")
			assert.Contains(t, with, "github_token")
			assert.Contains(t, with, "release_id")
			assert.Contains(t, with, "file_globs")

			globs, _ := with["file_globs"].(string)
			assert.True(t, strings.Contains(globs, "dist/*.zip"), "must scan dist/*.zip archives")

			cond, _ := step["if"].(string)
			assert.True(t, strings.Contains(cond, "VT_API_KEY"), "must conditionally run only when VT_API_KEY is present")
		}
	}

	assert.True(t, hasGoReleaser, "must include goreleaser-action step")
	assert.True(t, hasVTAction, "must include virustotal-action step in release workflow")
}

func TestAllWorkflowsValidYAML(t *testing.T) {
	workflowDir := filepath.Join(".github", "workflows")
	entries, err := os.ReadDir(workflowDir)
	require.NoError(t, err, "must read workflow directory")

	yamlCount := 0
	for _, entry := range entries {
		if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".yml") && !strings.HasSuffix(entry.Name(), ".yaml")) {
			continue
		}
		yamlCount++
		filePath := filepath.Join(workflowDir, entry.Name())
		data, err := os.ReadFile(filePath)
		require.NoError(t, err, "failed to read %s", entry.Name())

		var parsed map[string]any
		err = yaml.Unmarshal(data, &parsed)
		require.NoError(t, err, "workflow %s must be valid YAML", entry.Name())
		assert.NotEmpty(t, parsed["name"], "workflow %s must have a name", entry.Name())
		assert.NotEmpty(t, parsed["jobs"], "workflow %s must have jobs", entry.Name())
	}

	assert.GreaterOrEqual(t, yamlCount, 4, "must find at least 4 workflow files")
}

func TestWorkflowEdgeCases(t *testing.T) {
	_, err := os.ReadFile(filepath.Join(".github", "workflows", "non_existent_file.yml"))
	assert.Error(t, err, "reading non-existent workflow should error")

	invalidYAML := []byte("name: invalid\njobs: [\nincomplete")
	var parsed map[string]any
	err = yaml.Unmarshal(invalidYAML, &parsed)
	assert.Error(t, err, "unmarshaling malformed YAML should error")
}
