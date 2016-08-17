package cmd

var (
	defaultEnv *Environment
)

func init() {
	defaultEnv = NewEnvironment()
}
