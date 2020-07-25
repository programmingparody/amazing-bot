package scrapers

import (
	"net/http"
)

type HTTPStepParameters struct {
	Client   *http.Client
	Request  *http.Request
	Response *http.Response
}

// HTTPStep represents a single process in the scraping flow
// IE: Recaptcha Handling (Utilizing the response object and sending a request using the http.Client)
// IE: HTML Parsing
type HTTPStep interface {
	Step(HTTPStepParameters) (HTTPStepParameters, error)
}

type RoutineStep struct {
	HTTPStep   HTTPStep
	BeforeStep func(parameters HTTPStepParameters)
	AfterStep  func(parameters HTTPStepParameters, e error)
}

func RoutineStepFromHTTPStep(step HTTPStep) RoutineStep {
	return RoutineStep{
		HTTPStep: step,
	}
}

type HTTPRoutine struct {
	Steps []RoutineStep
}

func NewHTTPRoutine(steps []RoutineStep) HTTPRoutine {
	return HTTPRoutine{
		Steps: steps,
	}
}
func NewEmtpyHTTPRoutine() HTTPRoutine {
	return HTTPRoutine{
		Steps: []RoutineStep{},
	}
}

func (hr *HTTPRoutine) AddHTTPStep(step HTTPStep) {
	hr.Steps = append(hr.Steps, RoutineStepFromHTTPStep(step))
}

func (hr *HTTPRoutine) Run(initialParameters HTTPStepParameters) (error, *RoutineStep) {
	currentParameters := initialParameters
	var currentError error = nil
	for _, step := range hr.Steps {
		if step.BeforeStep != nil {
			step.BeforeStep(currentParameters)
		}

		currentParameters, currentError = step.HTTPStep.Step(currentParameters)

		if step.AfterStep != nil {
			step.AfterStep(currentParameters, currentError)
		}

		if currentError != nil {
			return currentError, &step
		}
	}
	return nil, nil
}
