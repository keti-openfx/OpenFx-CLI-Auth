package grpc

import (
	"context"
	"errors"
	"strings"

	"github.com/keti-openfx/openfx-cli/pb"
	grpcgo "google.golang.org/grpc"
)

func Delete(fxGateway, functionName string, token string) error {

	gateway := strings.TrimRight(fxGateway, "/")

	conn, err := grpcgo.Dial(gateway, grpcgo.WithInsecure())
	if err != nil {
		return errors.New("did not connect: " + err.Error())
	}
	client := pb.NewFxGatewayClient(conn)

	_, statusErr := client.Delete(context.Background(), &pb.DeleteFunctionRequest{FunctionName: functionName, Token: token})
	if statusErr != nil {
		return errors.New("did not delete: " + statusErr.Error())
	}

	return nil
}
