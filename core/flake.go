package core

import (
	"embed"
	"errors"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
)

type Data struct {
	Config          *Config
	UserName        string
	Home            string
	LowPackages     []string
	DefaultPackages []string
	HighPackages    []string
	LowPrograms     []string
	DefaultPrograms []string
	HighPrograms    []string
}

// InitFlake writes the first flake configuration
func InitFlake(force bool) error {
	t, err := template.ParseFS(content, "*.tmpl")
	if err != nil {
		return err
	}
	conf, err := ReadConfig()
	if err != nil {
		return err
	}
	err = conf.Validate()
	if err != nil {
		return err
	}

	data := Data{
		Config:          conf,
		LowPackages:     LowPackages,
		DefaultPackages: DefaultPackages,
		HighPackages:    HighPackages,
		LowPrograms:     LowPrograms,
		DefaultPrograms: DefaultPrograms,
		HighPrograms:    HighPrograms,
	}

	err = writeFile("flake.nix", t, data, force)
	if err != nil {
		return err
	}

	err = writeFile("home.nix", t, data, force)
	if err != nil {
		return err
	}
	err = writeFile("aliases.nix", t, data, force)
	if err != nil {
		return err
	}
	err = writeFile("path.nix", t, data, force)
	if err != nil {
		return err
	}
	err = writeFile("programs.nix", t, data, force)
	if err != nil {
		return err
	}
	err = writeFile("shell.nix", t, data, force)
	if err != nil {
		return err
	}
	for _, sys := range data.Config.Systems {
		err = writeSystem(sys, t, force)
		if err != nil {
			return err
		}
	}
	err = writeFile("user.nix", t, data, force)
	if err != nil {
		return err
	}

	err = CreateRepo()
	if err != nil {
		return err
	}
	err = Commit()
	if err != nil {
		return err
	}
	return nil
}

// WriteFlake writes the applied flake configuration
func WriteFlake() error {
	t, err := template.ParseFS(content, "*.tmpl")
	if err != nil {
		return err
	}
	conf, err := ReadConfig()
	if err != nil {
		return err
	}
	err = conf.Validate()
	if err != nil {
		return err
	}

	data := Data{
		Config:          conf,
		LowPackages:     LowPackages,
		DefaultPackages: DefaultPackages,
		HighPackages:    HighPackages,
		LowPrograms:     LowPrograms,
		DefaultPrograms: DefaultPrograms,
		HighPrograms:    HighPrograms,
	}
	err = writeFile("flake.nix", t, data, true)
	if err != nil {
		return err
	}
	err = writeFile("home.nix", t, data, true)
	if err != nil {
		return err
	}
	err = writeFile("aliases.nix", t, data, true)
	if err != nil {
		return err
	}
	err = writeFile("path.nix", t, data, true)
	if err != nil {
		return err
	}
	err = writeFile("programs.nix", t, data, true)
	if err != nil {
		return err
	}
	for _, sys := range data.Config.Systems {
		err = writeSystem(sys, t, true)
		if err != nil {
			return err
		}
	}
	err = writeFile("shell.nix", t, data, true)
	if err != nil {
		return err
	}
	return Commit()

}

func ApplyFlake() error {
	conf, err := ReadConfig()
	if err != nil {
		return err
	}
	err = conf.Validate()
	if err != nil {
		return err
	}
	workdir, err := FlakeLocation()
	if err != nil {
		return err
	}
	user, err := Username()
	if err != nil {
		return err
	}
	host, err := Hostname()
	if err != nil {
		return err
	}
	apply := exec.Command("nix", "run", "--impure", "home-manager/master", "--", "-b", "bak", "switch", "--flake", ".#"+user+"@"+host)
	apply.Stderr = os.Stderr
	apply.Stdin = os.Stdin
	apply.Stdout = os.Stdout
	apply.Dir = workdir
	apply.Env = os.Environ()

	if conf.Unfree {
		apply.Env = append(apply.Env, "NIXPKGS_ALLOW_UNFREE=1")
	}

	err = apply.Run()
	if err != nil {
		return err
	}
	return nil
}
func CheckFlake() error {
	conf, err := ReadConfig()
	if err != nil {
		return err
	}
	workdir, err := FlakeLocation()
	if err != nil {
		return err
	}
	apply := exec.Command("nix", "run", "--impure", "home-manager/master", "build", "--impure", "--", "--flake", ".")
	apply.Stderr = os.Stderr
	apply.Stdin = os.Stdin
	apply.Stdout = os.Stdout
	apply.Dir = workdir
	apply.Env = os.Environ()
	if conf.Unfree {
		apply.Env = append(apply.Env, "NIXPKGS_ALLOW_UNFREE=1")
	}

	err = apply.Run()
	if err != nil {
		return err
	}
	return nil
}
func writeFile(fname string, t *template.Template, d Data, force bool) error {
	fleekPath, err := FlakeLocation()
	if err != nil {
		return err
	}
	fpath := filepath.Join(fleekPath, fname)
	_, err = os.Stat(fpath)
	if force || os.IsNotExist(err) {

		f, err := os.Create(fpath)
		if err != nil {
			return err
		}
		defer f.Close()
		tmplName := fname + ".tmpl"
		if err = t.ExecuteTemplate(f, tmplName, d); err != nil {
			return err
		}
	} else {
		return errors.New("cowardly refusing to overwrite existing file without --force flag")
	}
	return nil
}
func writeSystem(sys System, t *template.Template, force bool) error {
	fleekPath, err := FlakeLocation()
	if err != nil {
		return err
	}
	hostPath := filepath.Join(fleekPath, sys.Hostname)
	err = os.MkdirAll(hostPath, 0755)
	if err != nil {
		return err
	}
	fpath := filepath.Join(hostPath, sys.Hostname+".nix")
	_, err = os.Stat(fpath)
	if force || os.IsNotExist(err) {

		f, err := os.Create(fpath)
		if err != nil {
			return err
		}
		defer f.Close()
		tmplName := "host.nix.tmpl"
		if err = t.ExecuteTemplate(f, tmplName, sys); err != nil {
			return err
		}
	} else {
		return errors.New("cowardly refusing to overwrite existing file without --force flag")
	}
	upath := filepath.Join(hostPath, "user.nix")
	_, err = os.Stat(upath)
	if force || os.IsNotExist(err) {

		f, err := os.Create(upath)
		if err != nil {
			return err
		}
		defer f.Close()
		tmplName := "user.nix.tmpl"
		if err = t.ExecuteTemplate(f, tmplName, sys); err != nil {
			return err
		}
	} else {
		return errors.New("cowardly refusing to overwrite existing file without --force flag")
	}
	return nil
}

var (
	//go:embed *.tmpl
	content embed.FS
)
