package config

type Driver struct {
	Type        int    `yaml:"type"`
	Waitn       int    `yaml:"waitn"`
	Url         string `yaml:"url"`
	AccessToken string `yaml:"access_token"`
}

var Bot struct {
	ID            int64    `yaml:"id"`
	NickName      []string `yaml:"nick_name"`
	CommandPrefix string   `yaml:"command_prefix"`
	SuperUsers    []int64  `yaml:"super_users"`
	Driver        []Driver `yaml:"driver"`
}
