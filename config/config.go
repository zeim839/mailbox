package config

import "github.com/spf13/viper"

// Config defines the configuration parameters for a Mailbox server.
type Config struct {
	MongoURI      string `mapstructure:"MONGO_URI"`
	GinMode       string `mapstructure:"GIN_MODE"`
	Port          string `mapstructure:"PORT"`
	Username      string `mapstructure:"USERNAME"`
	Password      string `mapstructure:"PASSWORD"`
	CaptchaSecret string `mapstructure:"CAPTCHA_SECRET"`
}

// LoadConfig fetches a configuration from the given directory path.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName(".env")
	viper.AutomaticEnv()
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("GIN_MODE", "debug")
	viper.SetDefault("CAPTCHA_SECRET", "")
	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
