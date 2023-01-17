package maestro

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/midbel/maestro/internal/help"
	"github.com/midbel/maestro/internal/validate"
	"golang.org/x/crypto/ssh"
)

const DefaultSSHPort = 22

type CommandSettings struct {
	Visible bool

	Name       string
	Alias      []string
	Short      string
	Desc       string
	Categories []string

	Retry   int64
	WorkDir string
	Timeout time.Duration

	Hosts   []CommandTarget
	Deps    []CommandDep
	Options []CommandOption
	Args    []CommandArg
	Lines   CommandScript

	Ev nameset

	locals *Env
}

func NewCommmandSettings(name string) (CommandSettings, error) {
	return NewCommandSettingsWithLocals(name, EmptyEnv())
}

func NewCommandSettingsWithLocals(name string, locals *Env) (CommandSettings, error) {
	cmd := CommandSettings{
		Name:   name,
		locals: locals,
		Ev:     make(nameset),
	}
	if cmd.locals == nil {
		cmd.locals = EmptyEnv()
	}
	return cmd, nil
}

func (s CommandSettings) Command() string {
	return s.Name
}

func (s CommandSettings) About() string {
	return s.Short
}

func (s CommandSettings) Help() (string, error) {
	return help.Command(s)
}

func (s CommandSettings) Tags() []string {
	if len(s.Categories) == 0 {
		return []string{"default"}
	}
	return s.Categories
}

func (s CommandSettings) Usage() string {
	var str strings.Builder
	str.WriteString(s.Name)
	for _, o := range s.Options {
		str.WriteString(" ")
		str.WriteString("[")
		if o.Short != "" {
			str.WriteString("-")
			str.WriteString(o.Short)
		}
		if o.Short != "" && o.Long != "" {
			str.WriteString("/")
		}
		if o.Long != "" {
			str.WriteString("--")
			str.WriteString(o.Long)
		}
		str.WriteString("]")
	}
	for _, a := range s.Args {
		str.WriteString(" ")
		str.WriteString("<")
		str.WriteString(a.Name)
		str.WriteString(">")
	}
	return str.String()
}

func (s CommandSettings) Blocked() bool {
	return !s.Visible
}

func (s CommandSettings) Remote() bool {
	return len(s.Hosts) > 0
}

type CommandTarget struct {
	Addr       string
	User       string
	Pass       string
	Key        ssh.Signer
	KnownHosts ssh.HostKeyCallback
}

func (c CommandTarget) Config(top *ssh.ClientConfig) *ssh.ClientConfig {
	conf := &ssh.ClientConfig{
		User:            top.User,
		HostKeyCallback: top.HostKeyCallback,
	}
	if c.KnownHosts != nil {
		conf.HostKeyCallback = c.KnownHosts
	}
	if c.User != "" {
		conf.User = c.User
	}
	if c.Pass != "" {
		conf.Auth = append(conf.Auth, ssh.Password(c.Pass))
	}
	if c.Key != nil {
		conf.Auth = append(conf.Auth, ssh.PublicKeys(c.Key))
	}
	if len(conf.Auth) == 0 {
		conf.Auth = append(conf.Auth, top.Auth...)
	}
	return conf
}

type CommandScript []string

func (c CommandScript) Reader() io.Reader {
	var str bytes.Buffer
	for i := range c {
		if i > 0 {
			str.WriteString("\n")
		}
		str.WriteString(c[i])
	}
	return &str
}

type CommandDep struct {
	Name string
	Args []string
	// Bg        bool
	// Optional  bool
	// Mandatory bool
}

func (c CommandDep) Key() string {
	return c.Name
}

type CommandOption struct {
	Short    string
	Long     string
	Help     string
	Required bool
	Flag     bool

	Default     string
	DefaultFlag bool

	Valid validate.ValidateFunc
}

type CommandArg struct {
	Name  string
	Valid validate.ValidateFunc
}

func (a CommandArg) Validate(arg string) error {
	if a.Valid == nil {
		return nil
	}
	return a.Valid(arg)
}
