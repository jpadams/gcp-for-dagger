// Code generated by dagger. DO NOT EDIT.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"

	"main/internal/dagger"
	"main/internal/telemetry"
)

var dag = dagger.Connect()

func Tracer() trace.Tracer {
	return otel.Tracer("dagger.io/sdk.go")
}

// used for local MarshalJSON implementations
var marshalCtx = context.Background()

// called by main()
func setMarshalContext(ctx context.Context) {
	marshalCtx = ctx
	dagger.SetMarshalContext(ctx)
}

type DaggerObject = dagger.DaggerObject

type ExecError = dagger.ExecError

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}

// convertSlice converts a slice of one type to a slice of another type using a
// converter function
func convertSlice[I any, O any](in []I, f func(I) O) []O {
	out := make([]O, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

func (r Gcp) MarshalJSON() ([]byte, error) {
	var concrete struct{}
	return json.Marshal(&concrete)
}

func (r *Gcp) UnmarshalJSON(bs []byte) error {
	var concrete struct{}
	err := json.Unmarshal(bs, &concrete)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	ctx := context.Background()

	// Direct slog to the new stderr. This is only for dev time debugging, and
	// runtime errors/warnings.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})))

	if err := dispatch(ctx); err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
}

func dispatch(ctx context.Context) error {
	ctx = telemetry.InitEmbedded(ctx, resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("dagger-go-sdk"),
		// TODO version?
	))
	defer telemetry.Close()

	// A lot of the "work" actually happens when we're marshalling the return
	// value, which entails getting object IDs, which happens in MarshalJSON,
	// which has no ctx argument, so we use this lovely global variable.
	setMarshalContext(ctx)

	fnCall := dag.CurrentFunctionCall()
	parentName, err := fnCall.ParentName(ctx)
	if err != nil {
		return fmt.Errorf("get parent name: %w", err)
	}
	fnName, err := fnCall.Name(ctx)
	if err != nil {
		return fmt.Errorf("get fn name: %w", err)
	}
	parentJson, err := fnCall.Parent(ctx)
	if err != nil {
		return fmt.Errorf("get fn parent: %w", err)
	}
	fnArgs, err := fnCall.InputArgs(ctx)
	if err != nil {
		return fmt.Errorf("get fn args: %w", err)
	}

	inputArgs := map[string][]byte{}
	for _, fnArg := range fnArgs {
		argName, err := fnArg.Name(ctx)
		if err != nil {
			return fmt.Errorf("get fn arg name: %w", err)
		}
		argValue, err := fnArg.Value(ctx)
		if err != nil {
			return fmt.Errorf("get fn arg value: %w", err)
		}
		inputArgs[argName] = []byte(argValue)
	}

	result, err := invoke(ctx, []byte(parentJson), parentName, fnName, inputArgs)
	if err != nil {
		return fmt.Errorf("invoke: %w", err)
	}
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err = fnCall.ReturnValue(ctx, dagger.JSON(resultBytes)); err != nil {
		return fmt.Errorf("store return value: %w", err)
	}
	return nil
}
func invoke(ctx context.Context, parentJSON []byte, parentName string, fnName string, inputArgs map[string][]byte) (_ any, err error) {
	_ = inputArgs
	switch parentName {
	case "Gcp":
		switch fnName {
		case "GetSecret":
			var parent Gcp
			err = json.Unmarshal(parentJSON, &parent)
			if err != nil {
				panic(fmt.Errorf("%s: %w", "failed to unmarshal parent object", err))
			}
			var gcpCredentials *dagger.File
			if inputArgs["gcpCredentials"] != nil {
				err = json.Unmarshal([]byte(inputArgs["gcpCredentials"]), &gcpCredentials)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg gcpCredentials", err))
				}
			}
			return (*Gcp).GetSecret(&parent, ctx, gcpCredentials)
		case "WithGcpSecret":
			var parent Gcp
			err = json.Unmarshal(parentJSON, &parent)
			if err != nil {
				panic(fmt.Errorf("%s: %w", "failed to unmarshal parent object", err))
			}
			var ctr *dagger.Container
			if inputArgs["ctr"] != nil {
				err = json.Unmarshal([]byte(inputArgs["ctr"]), &ctr)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg ctr", err))
				}
			}
			var gcpCredentials *dagger.File
			if inputArgs["gcpCredentials"] != nil {
				err = json.Unmarshal([]byte(inputArgs["gcpCredentials"]), &gcpCredentials)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg gcpCredentials", err))
				}
			}
			return (*Gcp).WithGcpSecret(&parent, ctx, ctr, gcpCredentials)
		case "GcloudCli":
			var parent Gcp
			err = json.Unmarshal(parentJSON, &parent)
			if err != nil {
				panic(fmt.Errorf("%s: %w", "failed to unmarshal parent object", err))
			}
			var project string
			if inputArgs["project"] != nil {
				err = json.Unmarshal([]byte(inputArgs["project"]), &project)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg project", err))
				}
			}
			var gcpCredentials *dagger.File
			if inputArgs["gcpCredentials"] != nil {
				err = json.Unmarshal([]byte(inputArgs["gcpCredentials"]), &gcpCredentials)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg gcpCredentials", err))
				}
			}
			return (*Gcp).GcloudCli(&parent, ctx, project, gcpCredentials)
		case "List":
			var parent Gcp
			err = json.Unmarshal(parentJSON, &parent)
			if err != nil {
				panic(fmt.Errorf("%s: %w", "failed to unmarshal parent object", err))
			}
			var account string
			if inputArgs["account"] != nil {
				err = json.Unmarshal([]byte(inputArgs["account"]), &account)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg account", err))
				}
			}
			var project string
			if inputArgs["project"] != nil {
				err = json.Unmarshal([]byte(inputArgs["project"]), &project)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg project", err))
				}
			}
			var gcpCredentials *dagger.File
			if inputArgs["gcpCredentials"] != nil {
				err = json.Unmarshal([]byte(inputArgs["gcpCredentials"]), &gcpCredentials)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg gcpCredentials", err))
				}
			}
			return (*Gcp).List(&parent, ctx, account, project, gcpCredentials)
		case "GarEnsureServiceAccountKey":
			var parent Gcp
			err = json.Unmarshal(parentJSON, &parent)
			if err != nil {
				panic(fmt.Errorf("%s: %w", "failed to unmarshal parent object", err))
			}
			var account string
			if inputArgs["account"] != nil {
				err = json.Unmarshal([]byte(inputArgs["account"]), &account)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg account", err))
				}
			}
			var region string
			if inputArgs["region"] != nil {
				err = json.Unmarshal([]byte(inputArgs["region"]), &region)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg region", err))
				}
			}
			var project string
			if inputArgs["project"] != nil {
				err = json.Unmarshal([]byte(inputArgs["project"]), &project)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg project", err))
				}
			}
			var gcpCredentials *dagger.File
			if inputArgs["gcpCredentials"] != nil {
				err = json.Unmarshal([]byte(inputArgs["gcpCredentials"]), &gcpCredentials)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg gcpCredentials", err))
				}
			}
			return (*Gcp).GarEnsureServiceAccountKey(&parent, ctx, account, region, project, gcpCredentials)
		case "GarPushExample":
			var parent Gcp
			err = json.Unmarshal(parentJSON, &parent)
			if err != nil {
				panic(fmt.Errorf("%s: %w", "failed to unmarshal parent object", err))
			}
			var account string
			if inputArgs["account"] != nil {
				err = json.Unmarshal([]byte(inputArgs["account"]), &account)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg account", err))
				}
			}
			var region string
			if inputArgs["region"] != nil {
				err = json.Unmarshal([]byte(inputArgs["region"]), &region)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg region", err))
				}
			}
			var project string
			if inputArgs["project"] != nil {
				err = json.Unmarshal([]byte(inputArgs["project"]), &project)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg project", err))
				}
			}
			var repo string
			if inputArgs["repo"] != nil {
				err = json.Unmarshal([]byte(inputArgs["repo"]), &repo)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg repo", err))
				}
			}
			var image string
			if inputArgs["image"] != nil {
				err = json.Unmarshal([]byte(inputArgs["image"]), &image)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg image", err))
				}
			}
			var gcpCredentials *dagger.File
			if inputArgs["gcpCredentials"] != nil {
				err = json.Unmarshal([]byte(inputArgs["gcpCredentials"]), &gcpCredentials)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg gcpCredentials", err))
				}
			}
			return (*Gcp).GarPushExample(&parent, ctx, account, region, project, repo, image, gcpCredentials)
		case "CleanupServiceAccountKey":
			var parent Gcp
			err = json.Unmarshal(parentJSON, &parent)
			if err != nil {
				panic(fmt.Errorf("%s: %w", "failed to unmarshal parent object", err))
			}
			var account string
			if inputArgs["account"] != nil {
				err = json.Unmarshal([]byte(inputArgs["account"]), &account)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg account", err))
				}
			}
			var region string
			if inputArgs["region"] != nil {
				err = json.Unmarshal([]byte(inputArgs["region"]), &region)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg region", err))
				}
			}
			var project string
			if inputArgs["project"] != nil {
				err = json.Unmarshal([]byte(inputArgs["project"]), &project)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg project", err))
				}
			}
			var gcpCredentials *dagger.File
			if inputArgs["gcpCredentials"] != nil {
				err = json.Unmarshal([]byte(inputArgs["gcpCredentials"]), &gcpCredentials)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg gcpCredentials", err))
				}
			}
			var keyId string
			if inputArgs["keyId"] != nil {
				err = json.Unmarshal([]byte(inputArgs["keyId"]), &keyId)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg keyId", err))
				}
			}
			return nil, (*Gcp).CleanupServiceAccountKey(&parent, ctx, account, region, project, gcpCredentials, keyId)
		case "GarPush":
			var parent Gcp
			err = json.Unmarshal(parentJSON, &parent)
			if err != nil {
				panic(fmt.Errorf("%s: %w", "failed to unmarshal parent object", err))
			}
			var pushCtr *dagger.Container
			if inputArgs["pushCtr"] != nil {
				err = json.Unmarshal([]byte(inputArgs["pushCtr"]), &pushCtr)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg pushCtr", err))
				}
			}
			var account string
			if inputArgs["account"] != nil {
				err = json.Unmarshal([]byte(inputArgs["account"]), &account)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg account", err))
				}
			}
			var region string
			if inputArgs["region"] != nil {
				err = json.Unmarshal([]byte(inputArgs["region"]), &region)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg region", err))
				}
			}
			var project string
			if inputArgs["project"] != nil {
				err = json.Unmarshal([]byte(inputArgs["project"]), &project)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg project", err))
				}
			}
			var repo string
			if inputArgs["repo"] != nil {
				err = json.Unmarshal([]byte(inputArgs["repo"]), &repo)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg repo", err))
				}
			}
			var image string
			if inputArgs["image"] != nil {
				err = json.Unmarshal([]byte(inputArgs["image"]), &image)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg image", err))
				}
			}
			var gcpCredentials *dagger.File
			if inputArgs["gcpCredentials"] != nil {
				err = json.Unmarshal([]byte(inputArgs["gcpCredentials"]), &gcpCredentials)
				if err != nil {
					panic(fmt.Errorf("%s: %w", "failed to unmarshal input arg gcpCredentials", err))
				}
			}
			return (*Gcp).GarPush(&parent, ctx, pushCtr, account, region, project, repo, image, gcpCredentials)
		default:
			return nil, fmt.Errorf("unknown function %s", fnName)
		}
	case "":
		return dag.Module().
			WithDescription("Push a container image into Google Artifact Registry\n\nThis module lets you push a container into Google Artifact Registry, automating the tedious manual steps of setting up a service account for the docker credential\n").
			WithObject(
				dag.TypeDef().WithObject("Gcp").
					WithFunction(
						dag.Function("GetSecret",
							dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("gcpCredentials", dag.TypeDef().WithObject("File"))).
					WithFunction(
						dag.Function("WithGcpSecret",
							dag.TypeDef().WithObject("Container")).
							WithArg("ctr", dag.TypeDef().WithObject("Container")).
							WithArg("gcpCredentials", dag.TypeDef().WithObject("File"))).
					WithFunction(
						dag.Function("GcloudCli",
							dag.TypeDef().WithObject("Container")).
							WithArg("project", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("gcpCredentials", dag.TypeDef().WithObject("File"))).
					WithFunction(
						dag.Function("List",
							dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("account", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("project", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("gcpCredentials", dag.TypeDef().WithObject("File"))).
					WithFunction(
						dag.Function("GarEnsureServiceAccountKey",
							dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("account", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("region", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("project", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("gcpCredentials", dag.TypeDef().WithObject("File"))).
					WithFunction(
						dag.Function("GarPushExample",
							dag.TypeDef().WithKind(dagger.StringKind)).
							WithDescription("Push ubuntu:latest to GAR under existing repo").
							WithArg("account", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("region", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("project", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("repo", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("image", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("gcpCredentials", dag.TypeDef().WithObject("File"))).
					WithFunction(
						dag.Function("CleanupServiceAccountKey",
							dag.TypeDef().WithKind(dagger.VoidKind).WithOptional(true)).
							WithArg("account", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("region", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("project", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("gcpCredentials", dag.TypeDef().WithObject("File")).
							WithArg("keyId", dag.TypeDef().WithKind(dagger.StringKind))).
					WithFunction(
						dag.Function("GarPush",
							dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("pushCtr", dag.TypeDef().WithObject("Container")).
							WithArg("account", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("region", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("project", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("repo", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("image", dag.TypeDef().WithKind(dagger.StringKind)).
							WithArg("gcpCredentials", dag.TypeDef().WithObject("File")))), nil
	default:
		return nil, fmt.Errorf("unknown object %s", parentName)
	}
}
