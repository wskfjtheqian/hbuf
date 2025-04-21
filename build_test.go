package hbuf

import (
	"archive/zip"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

// 编译程序
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

// 复制文件
func copyFile(t *testing.T, src, dst string) {
	t.Log("Copying file from " + src + " to " + dst)
	srcFile, err := os.Open(src)
	if err != nil {
		t.Error(err)
		return
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		t.Error(err)
		return
	}
	defer dstFile.Close()
	_, err = dstFile.ReadFrom(srcFile)
	if err != nil {
		t.Error(err)
		return
	}
}

// 使用GO代码打包ZIP文件
func buildZip(t *testing.T, src, dst string) {
	zipFile, err := os.Create(dst)
	if err != nil {
		t.Error(err)
		return
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	srcFile, err := os.Open(src)
	if err != nil {
		t.Error(err)
		return
	}

	defer srcFile.Close()
	f, err := w.Create(filepath.Base(src))
	if err != nil {
		t.Error(err)
		return
	}

	_, err = io.Copy(f, srcFile)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Building zip file from " + src + " to " + dst)

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
		err := build(t, "./bin/hbuf.darwin", "GOOS=darwin", "GOARCH=amd64")
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
	t.Run("Copy config file", func(t *testing.T) {
		copyFile(t, "./bin/hbuf.exe", "E:\\develop\\hanber\\hbuf.exe")
		copyFile(t, "./bin/hbuf.darwin", "E:\\develop\\hanber\\hbuf.darwin")
		copyFile(t, "./bin/hbuf.linux", "E:\\develop\\hanber\\hbuf.linux")

	})
}

type BuildConfig struct {
	GOOS   string
	GOARCH string
	Ext    string
}

func TestBuildAll(t *testing.T) {

	list := []BuildConfig{
		{"linux", "amd64", ""},
		{"windows", "amd64", ".exe"},
		{"darwin", "amd64", ""},
		{"linux", "386", ""},
		{"windows", "386", ".exe"},
		{"linux", "arm64", ""},
		{"windows", "arm64", ".exe"},
		{"darwin", "arm64", ""},
		{"linux", "arm", ""},
		{"windows", "arm", ".exe"},
	}

	version := gitVersion()
	for _, config := range list {
		t.Log("Building server for " + config.GOOS + "/" + config.GOARCH)

		bin := "./bin/hbuf" + config.Ext
		err := build(t, bin, "GOOS="+config.GOOS, "GOARCH="+config.GOARCH, "CGO_ENABLED=0")
		if err != nil {
			t.Error(err)
		}

		buildZip(t, bin, "./bin/"+config.GOOS+"_"+config.GOARCH+"_"+version+".zip")
		defer os.Remove(bin)
	}
}
