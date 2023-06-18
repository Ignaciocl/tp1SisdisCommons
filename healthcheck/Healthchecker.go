package healthcheck

type HealthChecker interface {
	Run() error
}
