package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	sparta "github.com/mweagle/Sparta"
)

// slackLambdaJSONEvent provides a pass through mapping
// of all whitelisted Parameters.  The transformation is defined
// by the resources/gateway/inputmapping_json.vtl template.
type slackLambdaJSONEvent struct {
	// HTTPMethod
	Method string `json:"method"`
	// BodyParams, if available.  This is going to be an interface s.t. we can support
	// testing through APIGateway, which by default sends 'application/json'
	BodyParams map[string]interface{} `json:"body"`
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
func helloSlackbot(w http.ResponseWriter, r *http.Request) {
	logger, _ := r.Context().Value(sparta.ContextKeyLogger).(*logrus.Logger)

	// 1. Unmarshal the primary event
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var lambdaEvent slackLambdaJSONEvent
	err := decoder.Decode(&lambdaEvent)
	if err != nil {
		logger.Error("Failed to unmarshal event data: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.WithFields(logrus.Fields{
		"BodyType":  fmt.Sprintf("%T", lambdaEvent.BodyParams),
		"BodyValue": fmt.Sprintf("%+v", lambdaEvent.BodyParams),
	}).Info("Slack slashcommand values")

	// 2. Conditionally unmarshal to get the Slack text.  See
	// https://api.slack.com/slash-commands
	// for the value name list

	// 3. Create the response
	// Slack formatting:
	// https://api.slack.com/docs/formatting
	responseText := "Here's what I understood"
	for eachKey, eachParam := range lambdaEvent.BodyParams {
		responseText += fmt.Sprintf("\n*%s*: %+v", eachKey, eachParam)
	}

	// 4. Setup the response object:
	// https://api.slack.com/slash-commands, "Responding to a command"
	responseData := sparta.ArbitraryJSONObject{
		"response_type": "in_channel",
		"text":          responseText,
		"mrkdwn":        true,
	}
	// 5. Send it off
	responseBody, err := json.Marshal(responseData)
	if err != nil {
		logger.Error("Failed to marshal response: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(responseBody)
}

func spartaLambdaFunctions(api *sparta.API) []*sparta.LambdaAWSInfo {
	var lambdaFunctions []*sparta.LambdaAWSInfo
	lambdaFn := sparta.HandleAWSLambda(sparta.LambdaName(helloSlackbot),
		http.HandlerFunc(helloSlackbot),
		sparta.IAMRoleDefinition{})

	if nil != api {
		apiGatewayResource, _ := api.NewResource("/slack", lambdaFn)
		_, err := apiGatewayResource.NewMethod("POST", http.StatusCreated)
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
