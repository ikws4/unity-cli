package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	connectorPackageName       = "com.ikws4.unity-cli-connector"
	legacyConnectorPackageName = "com.youngwoocho02.unity-cli-connector"
	connectorPackageURL        = "https://github.com/ikws4/unity-cli.git?path=unity-connector"
)

func setupCmd(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("setup does not accept arguments; run it from the Unity project root")
	}

	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}
	return runSetup(projectRoot, Version, os.Stdout)
}

func runSetup(projectRoot, version string, out io.Writer) error {
	manifestPath, err := unityManifestPath(projectRoot)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("read Packages/manifest.json: %w", err)
	}

	var manifest map[string]json.RawMessage
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("parse Packages/manifest.json: %w", err)
	}
	if manifest == nil {
		return fmt.Errorf("parse Packages/manifest.json: expected a JSON object")
	}

	dependencies := make(map[string]string)
	if raw, ok := manifest["dependencies"]; ok {
		if err := json.Unmarshal(raw, &dependencies); err != nil {
			return fmt.Errorf("parse Packages/manifest.json dependencies: %w", err)
		}
		if dependencies == nil {
			dependencies = make(map[string]string)
		}
	}

	desired := connectorURLForVersion(version)
	current, exists := dependencies[connectorPackageName]
	_, legacyExists := dependencies[legacyConnectorPackageName]
	if exists && current == desired && !legacyExists {
		_, err := fmt.Fprintf(out, "Unity CLI Connector is already configured in Packages/manifest.json:\n  %s\n", desired)
		return err
	}

	delete(dependencies, legacyConnectorPackageName)
	dependencies[connectorPackageName] = desired
	encodedDependencies, err := json.Marshal(dependencies)
	if err != nil {
		return fmt.Errorf("encode connector dependency: %w", err)
	}
	manifest["dependencies"] = encodedDependencies

	encodedManifest, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("encode Packages/manifest.json: %w", err)
	}
	encodedManifest = append(encodedManifest, '\n')
	if err := writeFileAtomically(manifestPath, encodedManifest); err != nil {
		return fmt.Errorf("write Packages/manifest.json: %w", err)
	}

	action := "Installed"
	if exists || legacyExists {
		action = "Updated"
	}
	if _, err := fmt.Fprintf(out, "%s Unity CLI Connector in Packages/manifest.json:\n  %s\n", action, desired); err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, "Open the project in Unity to resolve the package and start the Connector.")
	return err
}

func unityManifestPath(projectRoot string) (string, error) {
	requiredDirs := []string{"Assets", "Packages", "ProjectSettings"}
	for _, relative := range requiredDirs {
		path := filepath.Join(projectRoot, relative)
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			return "", fmt.Errorf("setup must be run from a Unity project root (missing %s directory)", relative)
		}
	}

	projectVersionPath := filepath.Join(projectRoot, "ProjectSettings", "ProjectVersion.txt")
	if info, err := os.Stat(projectVersionPath); err != nil || info.IsDir() {
		return "", fmt.Errorf("setup must be run from a Unity project root (missing ProjectSettings/ProjectVersion.txt)")
	}

	manifestPath := filepath.Join(projectRoot, "Packages", "manifest.json")
	if info, err := os.Stat(manifestPath); err != nil || info.IsDir() {
		return "", fmt.Errorf("setup must be run from a Unity project root (missing Packages/manifest.json)")
	}
	return manifestPath, nil
}

func connectorURLForVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" || version == "dev" {
		return connectorPackageURL
	}
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return connectorPackageURL + "#" + version
}

func writeFileAtomically(path string, data []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	temp, err := os.CreateTemp(filepath.Dir(path), ".manifest-*.json")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() {
		_ = temp.Close()
		_ = os.Remove(tempPath)
	}()

	if err := temp.Chmod(info.Mode().Perm()); err != nil {
		return err
	}
	if _, err := temp.Write(data); err != nil {
		return err
	}
	if err := temp.Sync(); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	return nil
}
