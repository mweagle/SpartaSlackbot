package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Sirupsen/logrus"
	sparta "github.com/mweagle/Sparta"
)

// slackLambdaJSONEvent provides a pass through mapping
// of all whitelisted Parameters.  The transformation is defined
// by the resources/gateway/inputmapping_json.vtl template.
type slackLambdaJSONEvent struct {
	// HTTPMethod
	Method string `json:"method"`
	// Body, if available.  This is going to be an interface s.t. we can support
	// testing through APIGateway, which by default sends 'application/json'
	Body interface{} `json:"body"`
	// Whitelisted HTTP headers
	Headers map[string]string `json:"headers"`
	// Whitelisted HTTP query params
	QueryParams map[string]string `json:"queryParams"`
	// Whitelisted path parameters
	PathParams map[string]string `json:"pathParams"`
	// Context information - http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-mapping-template-reference.html#context-variable-reference
	Context sparta.APIGatewayContext `json:"context"`
}

////////////////////////////////////////////////////////////////////////////////
// Hello world event handler
//
func helloSlackbot(event *json.RawMessage,
	context *sparta.LambdaContext,
	w http.ResponseWriter,
	logger *logrus.Logger) {

	// 1. Unmarshal the primary event
	var lambdaEvent slackLambdaJSONEvent
	err := json.Unmarshal([]byte(*event), &lambdaEvent)
	if err != nil {
		logger.Error("Failed to unmarshal event data: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// 2. Conditionally unmarshal to get the Slack text.  See
	// https://api.slack.com/slash-commands
	// for the value name list
	requestParams := url.Values{}
	if bodyData, ok := lambdaEvent.Body.(string); ok {
		requestParams, err = url.ParseQuery(bodyData)
		if err != nil {
			logger.Error("Failed to parse query: ", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		logger.WithFields(logrus.Fields{
			"Values": requestParams,
		}).Info("Slack slashcommand values")
	} else {
		logger.Info("Event body empty")
	}

	// 3. Create the response
	// Slack formatting:
	// https://api.slack.com/docs/formatting
	responseText := "You talkin to me?"
	for _, eachLine := range requestParams["text"] {
		responseText += fmt.Sprintf("\n>>> %s", eachLine)
	}

	// 4. Setup the response object:
	// https://api.slack.com/slash-commands, "Responding to a command"
	responseData := sparta.ArbitraryJSONObject{
		"response_type": "in_channel",
		"text":          responseText,
	}
	// 5. Send it off
	responseBody, err := json.Marshal(responseData)
	if err != nil {
		logger.Error("Failed to marshal response: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fmt.Fprint(w, string(responseBody))
}

func spartaLambdaFunctions(api *sparta.API) []*sparta.LambdaAWSInfo {
	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFn := sparta.NewLambda(sparta.IAMRoleDefinition{}, helloSlackbot, nil)

	if nil != api {
		apiGatewayResource, _ := api.NewResource("/slack", lambdaFn)
		_, err := apiGatewayResource.NewMethod("POST")
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
