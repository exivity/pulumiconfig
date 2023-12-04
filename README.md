# PulumiConfig

> **Experimental Project**: This project is in an experimental phase and may undergo significant changes. Use at your own risk.

[![GoTemplate](https://img.shields.io/badge/go/template-black?logo=go)](https://github.com/SchwarzIT/go-template)

PulumiConfig is a Golang library designed to improve the way developers manage configuration in Pulumi. By leveraging Golang structs, it simplifies the process of tracking and validating configuration keys, ensuring a more efficient and error-free deployment process in cloud infrastructure projects.

## Features

- **Seamless Integration**: Effortlessly integrates with Pulumi and Golang projects.
- **Automated Key Tracking**: Automatically tracks configuration keys using Golang structs.
- **JSON Tagging**: Supports JSON tagging for Pulumi configuration keys, including nested structs.
- **Validation**: Integrates with the Go Playground Validator for custom validation logic, allowing required values and complex validations.

## Installation

To integrate PulumiConfig into your Golang project, follow these steps:

```bash
go get -u github.com/exivity/pulumiconfig
```

## Usage

### Basic Usage

```go
import (
    "github.com/exivity/EaaS-Pulumi-Deployment/pkg/providers/config"
)

// Example of defining a PulumiConfig struct
type PulumiConfig struct {
    Name string `pulumi:"name" validate:"default=john-doe"`
}

// Example deployment function using PulumiConfig
func main() error {
    ...
    cfg := &PulumiConfig{}
    err = pulumiconfig.GetConfig(ctx, cfg, config.GetCustomValidations(ctx)...)
    ...
}
```

### Advanced Features

- **Custom Validation Logic**: Implement the `Validator` interface to create custom validation types. This is useful for scenarios that require specific validation rules beyond standard checks.

### Example Snippets

```go
// Custom Validator Implementation
type MyValidator struct {
    // Implementation details
}

func (v *MyValidator) Register(validate *validator.Validate) error {
    // Register custom validation logic
}

// Using type conversions in PulumiConfig
type MyConfig struct {
    MyNumber string `pulumi:"myNumber" validate:"int64"`
}
```

```go
// Example Test Case
func TestMyConfigValidation(t *testing.T) {
    // Test setup and assertions
}
```

## Contributing

We welcome contributions! Please refer to the `CODEOWNERS` file for guidelines on contributing to PulumiConfig.

## License

PulumiConfig is released under [LICENSE TYPE]. See the [LICENSE](./LICENSE) file for more details.

## Support

For support, questions, or contributions, please contact [CONTACT INFORMATION].
