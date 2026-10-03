package application

type Config struct {
	AppName string
}

func NewConfig() Config {
	return Config{
		AppName: "Jungle Gaming Backend Challenge",
	}
}
