package action

type Action interface {
	Run(params interface{}) error
}

type Params interface {
	ValidateAndFill(args []string) error
	Get() interface{}
}
