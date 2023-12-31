package main

import (
	"context"
	adminsample "sst-go-template/functions/api/admin/samples"
	"sst-go-template/internal/db"
	"sst-go-template/internal/model/sample"
	"sst-go-template/internal/request"
	"sst-go-template/internal/response"
	"sst-go-template/internal/storage"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var ssmClient = storage.LoadSSMClient()
var conn = db.Connect(ssmClient)
var repo = sample.NewRepository(conn)
var service = sample.NewService(repo)

func handler(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	id, err := request.GetPathInt64(event.PathParameters, "id")
	if err != nil {
		return response.ErrorJSON(err)
	}

	s, err := service.Get(ctx, id)
	if err != nil {
		return response.ErrorJSON(err)
	}

	res := adminsample.SampleResponse{
		ID:             s.ID,
		Name:           s.Name,
		Description:    s.Description,
		Amount:         s.Amount,
		Version:        s.Version,
		Translations:   adminsample.ToTranslationsResponse(s.Translations),
		CreatedAt:      s.CreatedAt,
		CreatedBy:      s.CreatedBy,
		LastModifiedAt: s.LastModifiedAt,
		LastModifiedBy: s.LastModifiedBy,
	}
	return response.JSON(res, 200)
}

func main() {
	lambda.Start(handler)
}
