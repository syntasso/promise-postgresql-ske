package main

import (
	kratix "github.com/syntasso/kratix-go"
	"github.com/syntasso/promise-postgresql/postgresql-configure-pipeline/internal/promise"
	"github.com/syntasso/promise-postgresql/postgresql-configure-pipeline/internal/resource"
)

func main() {
	sdk := kratix.New()
	switch {
	case sdk.IsPromiseWorkflow():
		promise.Configure()
	case sdk.IsResourceWorkflow():
		resource.Configure(sdk)
	}
}
