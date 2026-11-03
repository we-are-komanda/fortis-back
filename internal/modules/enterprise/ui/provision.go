package ui

import (
	"context"
	"encoding/json"
	"flag"
	"io"

	"github.com/fortis/backend/internal/modules/enterprise/application"
)

// WriteProvision is a local command transport. It is intentionally not an HTTP handler.
func WriteProvision(ctx context.Context, service *application.ProvisionService, out io.Writer, args []string) error {
	flags := flag.NewFlagSet("enterprise-provision", flag.ContinueOnError)
	// flag's default errors echo input values, which may contain operator data.
	flags.SetOutput(io.Discard)
	var command application.ProvisionEnterpriseCommand
	flags.StringVar(&command.EnterpriseID, "enterprise-id", "", "stable enterprise UUID")
	flags.StringVar(&command.UserID, "user-id", "", "existing user UUID")
	flags.StringVar(&command.Name, "name", "", "enterprise name")
	flags.StringVar(&command.Address, "address", "", "enterprise address")
	latitude, longitude := flags.Float64("latitude", 0, "explicit latitude"), flags.Float64("longitude", 0, "explicit longitude")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return application.ErrInvalidProvisionInput
	}
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "latitude" {
			command.Latitude = latitude
		}
		if f.Name == "longitude" {
			command.Longitude = longitude
		}
	})
	result, err := service.Provision(ctx, command)
	if err != nil {
		return err
	}
	return json.NewEncoder(out).Encode(provisionResultDTO{EnterpriseID: result.EnterpriseID, UserID: result.UserID})
}
