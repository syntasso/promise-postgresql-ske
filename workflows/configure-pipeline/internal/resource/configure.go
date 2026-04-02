package resource

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	kratix "github.com/syntasso/kratix-go"
	"github.com/syntasso/promise-postgresql/postgresql-configure-pipeline/internal/logger"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

const (
	referenceManifestPath = "/resources/references/minimal-postgres-manifest.yaml"
	outputFile            = "postgres-instance.yaml"
	statusFilePath        = "/kratix/metadata/status.yaml"
)

func Configure(sdk *kratix.KratixSDK) {
	logger := logger.New("resource-configure")
	start := time.Now()

	logger.Info("PostgreSQL resource configure started")

	resource, logger, err := readResourceInput(sdk, logger)
	if err != nil {
		fatal(logger, err, "failed to read resource input")
	}

	pgManifest, err := loadBaseManifest(referenceManifestPath)
	if err != nil {
		fatal(logger, err, "failed to load base manifest",
			"reference_path", referenceManifestPath,
		)
	}
	logger.Info("loaded PostgreSQL reference manifest",
		"reference_path", referenceManifestPath,
	)

	updateBaseManifest(resource, logger, &pgManifest)

	applyWorkflowStepOverrides(resource, logger, &pgManifest)

	applyEnvironmentProfileOverrides(resource, logger, &pgManifest)

	err = publishPostgresInstanceManifest(sdk, logger, outputFile, pgManifest)
	if err != nil {
		fatal(logger, err, "failed to render and publish PostgreSQL instance manifest",
			"output_path", filepath.Join("/kratix/output", outputFile),
		)
	}

	message, err := writeResourceStatus(sdk, logger, pgManifest)
	if err != nil {
		fatal(logger, err, "failed to write resource status",
			"output_path", statusFilePath,
		)
	}

	version, _, _ := unstructured.NestedString(pgManifest.Object, "spec", "postgresql", "version")
	instanceCount, _, _ := unstructured.NestedInt64(pgManifest.Object, "spec", "numberOfInstances")
	storageSize, _, _ := unstructured.NestedString(pgManifest.Object, "spec", "volume", "size")
	logger.Info("PostgreSQL resource configure completed",
		"instance_name", pgManifest.GetName(),
		"namespace", pgManifest.GetNamespace(),
		"pg_version", version,
		"instance_count", instanceCount,
		"storage_size", storageSize,
		"backup_enabled", envBool(pgManifest.Object, "spec", "enableLogicalBackup"),
		"duration_ms", time.Since(start).Milliseconds(),
		"result", message,
	)
}

func readResourceInput(sdk *kratix.KratixSDK, logger logr.Logger) (kratix.Resource, logr.Logger, error) {
	resource, err := sdk.ReadResourceInput()
	if err != nil {
		return nil, logger, err
	}

	logger = logger.WithValues("resource_name", resource.GetName())
	if namespace := resource.GetNamespace(); namespace != "" {
		logger = logger.WithValues("resource_namespace", namespace)
	}

	return resource, logger, nil
}

func loadBaseManifest(path string) (unstructured.Unstructured, error) {
	refManifest, err := os.ReadFile(path)
	if err != nil {
		return unstructured.Unstructured{}, fmt.Errorf("read base manifest: %w", err)
	}

	var pgManifest unstructured.Unstructured
	if err := yaml.Unmarshal(refManifest, &pgManifest.Object); err != nil {
		return unstructured.Unstructured{}, fmt.Errorf("parse base manifest: %w", err)
	}

	pgManifest.SetNamespace("default")
	unstructured.RemoveNestedField(pgManifest.Object, "spec", "preparedDatabases")
	return pgManifest, nil
}

func updateBaseManifest(resource kratix.Resource, logger logr.Logger, pgManifest *unstructured.Unstructured) {
	overrides := map[string]any{}

	if namespace, ok := resourceSpecString(resource, "namespace"); ok {
		pgManifest.SetNamespace(namespace)
		overrides["metadata.namespace"] = namespace
	}

	if team, ok := resourceSpecString(resource, "teamId"); ok {
		users := map[string]any{team: []any{"superuser", "createdb"}}
		unstructured.SetNestedField(pgManifest.Object, team, "spec", "teamId")
		unstructured.SetNestedField(pgManifest.Object, users, "spec", "users")
		overrides["spec.teamId"] = team
		overrides["spec.users"] = users
	}

	if dbName, ok := resourceSpecString(resource, "dbName"); ok {
		team, _, _ := unstructured.NestedString(pgManifest.Object, "spec", "teamId")
		databases := map[string]string{dbName: team}
		unstructured.SetNestedStringMap(pgManifest.Object, databases, "spec", "databases")
		overrides["spec.databases"] = databases
	}

	if backup, ok := resourceSpecBool(resource, "backupEnabled"); ok {
		unstructured.SetNestedField(pgManifest.Object, backup, "spec", "enableLogicalBackup")
		overrides["spec.enableLogicalBackup"] = backup
	}

	logger.Info("updated PostgreSQL manifest from resource input",
		"overrides", overrides,
	)
}

func applyWorkflowStepOverrides(resource kratix.Resource, logger logr.Logger, pgManifest *unstructured.Unstructured) {
	version := getEnv("PG_VERSION", "16")
	team, _, _ := unstructured.NestedString(pgManifest.Object, "spec", "teamId")
	instanceName := fmt.Sprintf("%s-%s-postgresql", team, resource.GetName())

	pgManifest.SetName(instanceName)
	unstructured.SetNestedField(pgManifest.Object, version, "spec", "postgresql", "version")

	logger.Info("applied PostgreSQL workflow step overrides",
		"overrides", map[string]any{
			"metadata.name":           pgManifest.GetName(),
			"spec.postgresql.version": version,
		},
	)
}

func applyEnvironmentProfileOverrides(resource kratix.Resource, logger logr.Logger, pgManifest *unstructured.Unstructured) {
	minReplicas, _ := strconv.Atoi(getEnv("MINIMUM_REPLICAS", "1"))

	envType := "dev"
	size := "1Gi"
	instances := int64(minReplicas)
	if requestedEnvType, ok := resourceSpecString(resource, "env"); ok {
		envType = requestedEnvType
	}
	if envType == "prod" {
		size = "10Gi"
		instances += 2
	}
	unstructured.SetNestedField(pgManifest.Object, instances, "spec", "numberOfInstances")
	unstructured.SetNestedField(pgManifest.Object, size, "spec", "volume", "size")

	logger.Info("applied PostgreSQL environment profile overrides",
		"environment", envType,
		"overrides", map[string]any{
			"spec.numberOfInstances": instances,
			"spec.volume.size":       size,
		},
	)
}

func publishPostgresInstanceManifest(
	sdk *kratix.KratixSDK,
	logger logr.Logger,
	outputFile string,
	pgManifest unstructured.Unstructured,
) error {
	out, err := yaml.Marshal(pgManifest.Object)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	if err := sdk.WriteOutput(outputFile, out); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	version, _, _ := unstructured.NestedString(pgManifest.Object, "spec", "postgresql", "version")
	instanceCount, _, _ := unstructured.NestedInt64(pgManifest.Object, "spec", "numberOfInstances")
	storageSize, _, _ := unstructured.NestedString(pgManifest.Object, "spec", "volume", "size")

	logger.Info("published PostgreSQL instance manifest",
		"instance_name", pgManifest.GetName(),
		"namespace", pgManifest.GetNamespace(),
		"output_path", filepath.Join("/kratix/output", outputFile),
		"pg_version", version,
		"instance_count", instanceCount,
		"storage_size", storageSize,
		"backup_enabled", envBool(pgManifest.Object, "spec", "enableLogicalBackup"),
	)

	return nil
}

func writeResourceStatus(
	sdk *kratix.KratixSDK,
	logger logr.Logger,
	pgManifest unstructured.Unstructured,
) (string, error) {
	version, _, _ := unstructured.NestedString(pgManifest.Object, "spec", "postgresql", "version")
	storageSize, _, _ := unstructured.NestedString(pgManifest.Object, "spec", "volume", "size")
	backupEnabled := envBool(pgManifest.Object, "spec", "enableLogicalBackup")
	message := fmt.Sprintf("%s instance v%s deployed successfully", storageSize, version)
	if backupEnabled {
		message += " with backups enabled"
	} else {
		message += " without backups"
	}

	status := kratix.NewStatus()
	if err := status.Set("message", message); err != nil {
		return "", fmt.Errorf("set status message: %w", err)
	}
	if err := status.Set("instanceName", pgManifest.GetName()); err != nil {
		return "", fmt.Errorf("set status instance name: %w", err)
	}
	if err := status.Set("pgVersion", version); err != nil {
		return "", fmt.Errorf("set status pg version: %w", err)
	}
	if err := status.Set("connectionDetails", map[string]string{
		"host":        fmt.Sprintf("%s.%s.svc.cluster.local", pgManifest.GetName(), pgManifest.GetNamespace()),
		"credentials": fmt.Sprintf("Username and Password available in Secret: \"%s/postgres.%s.credentials.postgresql.acid.zalan.do\"", pgManifest.GetNamespace(), pgManifest.GetName()),
	}); err != nil {
		return "", fmt.Errorf("set status connection details: %w", err)
	}
	if err := sdk.WriteStatus(status); err != nil {
		return "", fmt.Errorf("write status: %w", err)
	}

	logger.Info("published PostgreSQL status",
		"instance_name", pgManifest.GetName(),
		"namespace", pgManifest.GetNamespace(),
		"output_path", statusFilePath,
		"status_message", message,
	)

	return message, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func resourceSpecString(resource kratix.Resource, field string) (string, bool) {
	value, err := resource.GetValue("spec." + field)
	if err != nil {
		return "", false
	}

	stringValue, ok := value.(string)
	return stringValue, ok
}

func resourceSpecBool(resource kratix.Resource, field string) (bool, bool) {
	value, err := resource.GetValue("spec." + field)
	if err != nil {
		return false, false
	}

	boolValue, ok := value.(bool)
	return boolValue, ok
}

func fatal(logger interface {
	Error(error, string, ...interface{})
}, err error, message string, keysAndValues ...interface{}) {
	logger.Error(err, message, keysAndValues...)
	os.Exit(1)
}

func envBool(object map[string]any, fields ...string) bool {
	value, found, err := unstructured.NestedBool(object, fields...)
	return found && err == nil && value
}
