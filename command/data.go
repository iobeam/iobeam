package command

// Interface fo data that is posted to API, and generated from command-line input.
type Data interface {
	IsValid() bool
}
