package healthcheck

type HealthCheckerReplier interface {
	Run() error
}
