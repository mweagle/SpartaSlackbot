package main

import (
	"context"
	"fmt"
	"net/http"

	sparta "github.com/mweagle/Sparta"
	"github.com/mweagle/Sparta/aws/events"
	"github.com/sirupsen/logrus"
)

////////////////////////////////////////////////////////////////////////////////
// Hello world event handler
//
func helloSlackbot(ctx context.Context, apiRequest events.APIGatewayRequest) (map[string]interface{}, error) {
	logger, _ := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)

	bodyParams, bodyParamsOk := apiRequest.Body.(map[string]interface{})
	if !bodyParamsOk {
		return nil, fmt.Errorf("Failed to type convert body. Type: %T", apiRequest.Body)
	}

	logger.WithFields(logrus.Fields{
		"BodyType":  fmt.Sprintf("%T", bodyParams),
		"BodyValue": fmt.Sprintf("%+v", bodyParams),
	}).Info("Slack slashcommand values")

	// 2. Create the response
	// Slack formatting:
	// https://api.slack.com/docs/formatting
	responseText := "Here's what I understood"
	for eachKey, eachParam := range bodyParams {
		responseText += fmt.Sprintf("\n*%s*: %+v", eachKey, eachParam)
	}

	// 4. Setup the response object:
	// https://api.slack.com/slash-commands, "Responding to a command"
	responseData := map[string]interface{}{
		"response_type": "in_channel",
		"text":          responseText,
		"mrkdwn":        true,
	}
	return responseData, nil
}

func spartaLambdaFunctions(api *sparta.API) []*sparta.LambdaAWSInfo {
	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFn := sparta.HandleAWSLambda(sparta.LambdaName(helloSlackbot),
		helloSlackbot,
		sparta.IAMRoleDefinition{})

	if nil != api {
		apiGatewayResource, _ := api.NewResource("/slack", lambdaFn)
		_, err := apiGatewayResource.NewMethod("POST", http.StatusCreated, http.StatusCreated)
		if nil != err {
			panic("Failed to create /hello resource")
		}
	}
	return append(lambdaFunctions, lambdaFn)
}

////////////////////////////////////////////////////////////////////////////////
// Main
func main() {
	// Register the function with the API Gateway
	apiStage := sparta.NewStage("v1")
	apiGateway := sparta.NewAPIGateway("SpartaSlackbot", apiStage)

	// Deploy it
	sparta.Main("SpartaSlackbot",
		fmt.Sprintf("Sparta app that responds to Slack commands"),
		spartaLambdaFunctions(apiGateway),
		apiGateway,
		nil)
}
