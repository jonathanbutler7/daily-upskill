package main

import (
	"time"

	"weavelab.xyz/devx/pkg/daemon"
	"weavelab.xyz/monorail/shared/wlib/werror"
	"weavelab.xyz/monorail/shared/wlib/wgrpc/wgrpcserver"
	"weavelab.xyz/schema-gen-go/pkg/wgateway"
	willtheypay "weavelab.xyz/schema-gen-go/schemas/payments-platform/will-they-pay/v1"
	"weavelab.xyz/wcore"
	"weavelab.xyz/will-they-pay-service/internal/server"
)

func main() {
	daemon.Main(func(ctx wcore.Context, d *daemon.Daemon) error {
		willTheyPayServer := server.NewWillTheyPayServer("v0.1.0")

		ops := []daemon.GRPCOption{
			daemon.WithUnaryInterceptors(wgrpcserver.UnaryErrorInterceptor),
		}
		d.AddGRPCServer(
			func(s wgrpcserver.Server) error {
				willtheypay.RegisterWillTheyPayAPIServer(s, willTheyPayServer)
				return nil
			},
			ops...,
		)

		gateway, err := willtheypay.NewGateway(ctx,
			wgateway.WithRequestTimeoutDuration(60*time.Second),
		)
		if err != nil {
			return werror.Wrap(err, "failed to create gateway")
		}
		d.AddSchemaGatewayWithoutGrpc(gateway)

		return d.Run(ctx)
	})
}
