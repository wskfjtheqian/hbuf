package hbuf

import (
	"os"
	"os/exec"
	"testing"
)

// 获得Git的版本号
func gitVersion() string {
	version := "v1.0.0"

	cmd := exec.Command("git", "describe", "--tags", "--always")
	out, err := cmd.Output()
	if err == nil {
		version = string(out)
	}
	for i := 0; i < len(version); i++ {
		if version[i] == '\n' {
			version = version[:i]
			break
		}
	}

	return version
}

func build(t *testing.T, out string, env ...string) error {
	version := gitVersion()

	cmd := exec.Command("go", "build", "-ldflags", "-X main.version="+version, "-o", out, "./cmd/main.go")
	env = append(os.Environ(), env...)
	cmd.Env = append(cmd.Env, env...)
	cmd.Dir = "./"
	t.Log("Executing command: " + cmd.String())

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// 编译测试
func TestBuild(t *testing.T) {
	t.Run("Build linux", func(t *testing.T) {
		err := build(t, "./bin/hbuf.linux", "GOOS=linux", "GOARCH=amd64")
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("Build macos", func(t *testing.T) {
		err := build(t, "./bin/hbuf.macos", "GOOS=darwin", "GOARCH=amd64")
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("Build windows", func(t *testing.T) {
		err := build(t, "./bin/hbuf.exe", "GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=0")
		if err != nil {
			t.Error(err)
		}
	})

}
