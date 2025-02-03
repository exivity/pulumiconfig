# PulumiConfig

PulumiConfig is a Golang library designed to improve the way developers manage configuration in Pulumi. By leveraging Golang structs, it simplifies the process of tracking and validating configuration keys, ensuring a more efficient and error-free deployment process in cloud infrastructure projects.

## Features

- **Seamless Integration**: Effortlessly integrates with Pulumi and Golang projects.
- **Automated Key Tracking**: Automatically tracks configuration keys using Golang structs.
- **JSON Tagging**: Supports JSON tagging for Pulumi configuration keys, including nested structs.
- **[go-playground/validator](https://github.com/go-playground/validator)**, letting you define both field- and struct-level validations., allowing required values and complex validations.
- **Namespace Overrides**: Use `overrideConfigNamespace` to override specific fields with values from a different namespace.

## Installation

To integrate PulumiConfig into your Golang project, follow these steps:

```bash
go get -u github.com/exivity/pulumiconfig
```

## Usage

### Basic Usage

Pulumi stack configuration is typically stored in a `Pulumi.<stack>.yaml` file. PulumiConfig simplifies the process of reading and validating these configuration values.

```yaml
config:
  pulumiconfig:name: britney
```

```go
package main

import (
    "github.com/exivity/pulumiconfig/pkg/pulumiconfig"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Example of defining a PulumiConfig struct
type PulumiConfig struct {
    Name string `pulumi:"name" validate:"default=john-doe"`
}

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        cfg := &PulumiConfig{}
        err := pulumiconfig.GetConfig(ctx, cfg)
        if err != nil {
            return err
        }

        ctx.Export("name", pulumi.String(cfg.Name))

        return nil
    })
}

```

### Using `pulumiConfigNamespace`

The `pulumiConfigNamespace` tag allows you to specify a custom namespace for a field in your configuration struct. This is useful for grouping related configuration values under a specific namespace. Note that this tag only works on the first level of a configuration struct.

A use case could be adding provider credentials just once, so that it can be used for both the provider and within the user application.

```yaml
config:
  pulumiconfig:provider_credentials:
    token:
      secure: do7ipohcahaiShaupheo5Ooneeghoh
```

```go
package main

import (
    "github.com/exivity/pulumiconfig/pkg/pulumiconfig"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PulumiConfig struct {
    ProviderCredentials *ProviderCredentials `json:"provider_credentials" pulumiConfigNamespace:"provider" validate:"required"`
}

type ProviderCredentials struct {
    Token string `json:"token"`
}

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        cfg := &PulumiConfig{}
        err := pulumiconfig.GetConfig(ctx, cfg)
        if err != nil {
            return err
        }

        ctx.Export("provider_token", pulumi.String(cfg.ProviderCredentials.Token))

        return nil
    })
}

```

### Using `overrideConfigNamespace`

In some cases, you may want to override certain values with a separate namespace. For example, you might have a "global" config in the main namespace, but you wish to override some keys when running specific environments. This can be particularly useful when using Pulumi ESC, allowing you to set configuration once and use it in several stacks. An example could be a multi-stage deployment, where only credentials need to differ, or in development where a backup configuration is not needed.

```yaml
config:
  pulumiconfig:digital_ocean:
    region: AMS3
    project: staging-project
  prod:digital_ocean:
    project: production-project
```

```go
package main

import (
    "github.com/exivity/pulumiconfig/pkg/pulumiconfig"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PulumiConfig struct {
    ProdOverrides DigitalOceanConfig `json:"digital_ocean" overrideConfigNamespace:"prod"`
}

type DigitalOceanConfig struct {
    Region  string `json:"region"`
    Project string `json:"project"`
}

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        cfg := &PulumiConfig{}
        err := pulumiconfig.GetConfig(ctx, cfg)
        if err != nil {
            return err
        }

        ctx.Export("region", pulumi.String(cfg.ProdOverrides.Region))   // -> AMS3
        ctx.Export("project", pulumi.String(cfg.ProdOverrides.Project)) // -> production-project

        return nil
    })
}
```

You can use `overrideConfigNamespace` on any field-level struct tag. PulumiConfig will first load from the main namespace, and then—if `overrideConfigNamespace` is set—load the separate namespace and merge those values in.

### Example: Custom Field and Struct Validators

Below is a more in-depth example illustrating how you can combine PulumiConfig with the Pulumi DigitalOcean provider for domain-specific validation:

```go
package main

import (
    "fmt"

    "github.com/exivity/pulumiconfig/pkg/pulumiconfig"
    "github.com/go-playground/validator/v10"
    do "github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const (
    pulumiDigitalOceanNamespace = "digitalocean"
    pulumiDigitalOceanTokenKey  = "token"
)

type Configuration struct {
    DigitalOceanToken string `json:"digitalOceanToken" validate:"required"`
    KubernetesVersion string `json:"kubernetesVersion" validate:"required"`
    Region            string `json:"region" validate:"required"`
}

// GetCustomValidations returns a slice of Validators that run on a Configuration struct.
func GetCustomValidations(ctx *pulumi.Context) []pulumiconfig.Validator {
    v := &Validation{ctx: ctx}
    return []pulumiconfig.Validator{
        // Struct-level validation (checks if DO Token is set).
        pulumiconfig.StructValidation{
            Struct:   &Configuration{},
            Validate: v.DigitalOceanToken,
        },
        // Field-level validation example: fetch Kubernetes version from DO and region availability.
        pulumiconfig.FieldValidation{
            Tag:      "kubernetesVersion",
            Validate: v.KubernetesVersion,
        },
        pulumiconfig.FieldValidation{
            Tag:      "region",
            Validate: v.Region,
        },
        // Additional field-level validators omitted...
    }
}

type Validation struct {
    ctx *pulumi.Context
}

// DigitalOceanToken checks if the DigitalOcean token is set.
func (v *Validation) DigitalOceanToken(sl validator.StructLevel) {
    cfg := config.New(v.ctx, pulumiDigitalOceanNamespace)
    _, err := cfg.TrySecret(pulumiDigitalOceanTokenKey)
    if err != nil {
        // Log an error and mark validation as failed.
        v.ctx.Log.Error(fmt.Sprintf("Missing DigitalOcean API token: %v", err), nil)
        sl.ReportError(nil, "", "", "", "")
    }
}

// KubernetesVersion looks up the latest DO K8s version that matches the user-supplied prefix.
func (v *Validation) KubernetesVersion(fl validator.FieldLevel) bool {
    versionPrefix := fl.Field().String()

    versions, err := do.GetKubernetesVersions(v.ctx, &do.GetKubernetesVersionsArgs{
        VersionPrefix: pulumi.StringRef(versionPrefix),
    })
    if err != nil {
        v.ctx.Log.Error(fmt.Sprintf("Error fetching Kubernetes versions: %v", err), nil)
        return false
    }
    if len(versions.ValidVersions) == 0 {
        v.ctx.Log.Error(fmt.Sprintf("No matching Kubernetes versions found for prefix: %s", versionPrefix), nil)
        return false
    }

    // Update the struct field with the latest valid version.
    field := fl.Field()
    if field.CanSet() {
        field.SetString(versions.LatestVersion)
        v.ctx.Export("Kubernetes version", pulumi.String(versions.LatestVersion))
    }

    return true
}

// Region checks if the specified region is currently available.
func (v *Validation) Region(fl validator.FieldLevel) bool {
    region := fl.Field().String()

    regions, err := do.GetRegions(v.ctx, &do.GetRegionsArgs{
        Filters: []do.GetRegionsFilter{{Key: "available", Values: []string{"true"}}},
    })
    if err != nil {
        v.ctx.Log.Error(fmt.Sprintf("Error fetching regions: %v", err), nil)
        return false
    }

    for _, r := range regions.Regions {
        if r.Slug == region {
            return true
        }
    }

    v.ctx.Log.Error(fmt.Sprintf("Region '%s' is not available", region), nil)
    return false
}

// etc... (more field-level checks for node sizes, database node sizes, etc.)
```

This snippet demonstrates a **struct-level** validator (`DigitalOceanToken`) ensuring that a DigitalOcean API token is set, and **field-level** validators (`KubernetesVersion`, `Region`, etc.) that fetch data from the provider's API at deployment time.

For instance, after defining these validators, you might integrate them like so:

```go
func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        cfg := &PulumiConfig{}
        err := pulumiconfig.GetConfig(ctx, cfg, GetCustomValidations(ctx)...)
        if err != nil {
            return err
        }

        // Continue with your Pulumi logic.
        // ...

        return nil
    })
}
```

By combining `StructValidation` and `FieldValidation`, you can enforce both global and per-field checks for your Pulumi configurations. Adjust or extend as needed for your own providers or custom logic.

## License

PulumiConfig is released under MIT. See the [LICENSE](./LICENSE) file for more details.
