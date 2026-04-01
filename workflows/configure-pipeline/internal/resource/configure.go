package resource

import (
	"fmt"
	"log"
	"os"
	"strconv"

	kratix "github.com/syntasso/kratix-go"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func Configure(sdk *kratix.KratixSDK) {
	resource, err := sdk.ReadResourceInput()
	if err != nil {
		log.Fatalf("failed to read resource input: %v", err)
	}

	// Read base manifest — its values act as defaults
	refManifest, err := os.ReadFile("/resources/references/minimal-postgres-manifest.yaml")
	if err != nil {
		log.Fatalf("failed to read base manifest: %v", err)
	}

	var pgManifest unstructured.Unstructured
	if err := yaml.Unmarshal(refManifest, &pgManifest.Object); err != nil {
		log.Fatalf("failed to parse base manifest: %v", err)
	}
	pgManifest.SetNamespace("default")

	// Read current values from the provided resource request and update them in the CRD
	resourceRequest := resource.ToUnstructured()

	updateSpecIfPresent(resourceRequest, unstructured.NestedString, "namespace", func(namespace string) {
		pgManifest.SetNamespace(namespace)
	})

	updateSpecIfPresent(resourceRequest, unstructured.NestedString, "teamId", func(team string) {
		users := map[string]any{team: []any{"superuser", "createdb"}}
		unstructured.SetNestedField(pgManifest.Object, team, "spec", "teamId")
		unstructured.SetNestedField(pgManifest.Object, users, "spec", "users")
	})

	updateSpecIfPresent(resourceRequest, unstructured.NestedString, "dbName", func(dbName string) {
		team, _, _ := unstructured.NestedString(pgManifest.Object, "spec", "teamId")
		databases := map[string]string{dbName: team}
		unstructured.SetNestedStringMap(pgManifest.Object, databases, "spec", "databases")
	})

	updateSpecIfPresent(resourceRequest, unstructured.NestedBool, "backupEnabled", func(backup bool) {
		unstructured.SetNestedField(pgManifest.Object, backup, "spec", "enableLogicalBackup")
	})

	// Always set name (derived from resource) and env-driven values
	minReplicas, _ := strconv.Atoi(getEnv("MINIMUM_REPLICAS", "1"))
	version := getEnv("PG_VERSION", "16")
	team, _, _ := unstructured.NestedString(pgManifest.Object, "spec", "teamId")
	instanceName := fmt.Sprintf("%s-%s-postgresql", team, resource.GetName())

	pgManifest.SetName(instanceName)
	unstructured.SetNestedField(pgManifest.Object, version, "spec", "postgresql", "version")

	// Add overrides for production environments
	envType, _, _ := unstructured.NestedString(resourceRequest.Object, "spec", "env")
	size := "1Gi"
	instances := int64(minReplicas)
	if envType == "prod" {
		size = "10Gi"
		instances += 2
	}
	unstructured.SetNestedField(pgManifest.Object, size, "spec", "volume", "size")
	unstructured.SetNestedField(pgManifest.Object, instances, "spec", "numberOfInstances")
	unstructured.RemoveNestedField(pgManifest.Object, "spec", "preparedDatabases")

	out, err := yaml.Marshal(pgManifest.Object)
	if err != nil {
		log.Fatalf("failed to marshal manifest: %v", err)
	}
	if err := sdk.WriteOutput("postgres-instance.yaml", out); err != nil {
		log.Fatalf("failed to write output: %v", err)
	}

	// Write status
	backup, _, _ := unstructured.NestedBool(pgManifest.Object, "spec", "enableLogicalBackup")
	message := fmt.Sprintf("%s instance v%s deployed successfully", size, version)
	if backup {
		message += " with backups enabled"
	} else {
		message += " without backups"
	}

	namespace := pgManifest.GetNamespace()

	status := kratix.NewStatus()
	if err := status.Set("message", message); err != nil {
		log.Fatalf("failed to set status: %v", err)
	}
	if err := status.Set("instanceName", instanceName); err != nil {
		log.Fatalf("failed to set status: %v", err)
	}
	if err := status.Set("pgVersion", version); err != nil {
		log.Fatalf("failed to set status: %v", err)
	}
	if err := status.Set("connectionDetails", map[string]string{
		"host":        fmt.Sprintf("%s.%s.svc.cluster.local", instanceName, namespace),
		"credentials": fmt.Sprintf("Username and Password available in Secret: \"%s/postgres.%s.credentials.postgresql.acid.zalan.do\"", namespace, instanceName),
	}); err != nil {
		log.Fatalf("failed to set status: %v", err)
	}
	if err := sdk.WriteStatus(status); err != nil {
		log.Fatalf("failed to write status: %v", err)
	}

	fmt.Println(message)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// updateSpecIfPresent reads a spec field from src and calls fn with its value only if found.
func updateSpecIfPresent[T any](
	src unstructured.Unstructured,
	get func(map[string]any, ...string) (T, bool, error),
	field string,
	fn func(T),
) {
	if v, ok, _ := get(src.Object, "spec", field); ok {
		fn(v)
	}
}
