package scrapers

//MakeRequestStep executes the request by the httpClient
type MakeRequestStep struct{}

func (mr *MakeRequestStep) Step(params HTTPStepParameters) (HTTPStepParameters, error) {
	response, error := params.Client.Do(params.Request)
	params.Response = response
	return params, error
}
