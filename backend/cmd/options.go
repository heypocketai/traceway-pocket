package cmd

type options struct {
	sqlitePath      string
	clickhousePath  string
	port            int
	serverURL       string
	defaultUser     *defaultUserOpts
	defaultProjects []defaultProjectOpts
}

type defaultUserOpts struct {
	email    string
	password string
}

type defaultProjectOpts struct {
	name      string
	framework string
	token     string
}

type Option func(*options)

func WithSQLitePath(path string) Option {
	return func(o *options) {
		o.sqlitePath = path
	}
}

func WithClickhousePath(path string) Option {
	return func(o *options) {
		o.clickhousePath = path
	}
}

func WithPort(port int) Option {
	return func(o *options) {
		o.port = port
	}
}

func WithServerURL(url string) Option {
	return func(o *options) {
		o.serverURL = url
	}
}

func WithDefaultUser(email, password string) Option {
	return func(o *options) {
		o.defaultUser = &defaultUserOpts{email: email, password: password}
	}
}

func WithDefaultProject(name, framework, token string) Option {
	return func(o *options) {
		o.defaultProjects = append(o.defaultProjects, defaultProjectOpts{
			name: name, framework: framework, token: token,
		})
	}
}
