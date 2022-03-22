package proxy

type Config struct {
	MinPort           int
	MaxPort           int
	BaseDomain        string
	MaxConnsPerClient int

	InactiveHoursTimeout          int
	NoActiveSocketsMinutesTimeout int
	NoActiveSocketsChecks         int
}

func (pc *Config) MaxClients() int {
	return pc.MaxPort - pc.MinPort
}
