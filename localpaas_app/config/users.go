package config

type Users struct {
	Admin UserAdmin `toml:"admin"`
	Demo  UserDemo  `toml:"demo"`
}

type UserAdmin struct {
	Email    string `toml:"email" env:"LP_USER_ADMIN_EMAIL"`
	Username string `toml:"username" env:"LP_USER_ADMIN_USERNAME"`
	Password string `toml:"password" env:"LP_USER_ADMIN_PASSWORD"`
}

type UserDemo struct {
	UserID string `toml:"user_id" env:"LP_USER_DEMO_USER_ID"`
}
