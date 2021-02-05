package action

type Action interface {
	Run(params interface{}) error
}

type ActionParams interface {
	ValidateAndFill(args []string) error
	Get() interface{}
}
