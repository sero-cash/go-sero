package app

import (
	"runtime"
	"os/user"
	"os"
	"bytes"
	"os/exec"
	"strings"
	"errors"
	"fmt"
)

var (
	app_home_path     string
	app_keystore_path string
	app_log_path      string
	app_balance_path  string
	app_data_path     string
)

type App struct {
}

func (app *App) Init() error {

	if err := initDataPath(); err != nil {
		return err
	}

	return nil
}

func initDataPath() (err error) {
	if home, err := Home(); err != nil {
		return fmt.Errorf("Current operating system is not supported，err=[%v] ", err)
	} else {
		switch runtime.GOOS {
		case "darwin":
			app_home_path = home + "/Library/SeroLight"
			break
		case "windows":
			app_home_path = home + "\\AppData\\Roaming\\SeroLight"
			break
		case "linux":
			app_home_path = home + "/.config/SeroLight"
			break
		}
	}
	if app_home_path == "" {
		return fmt.Errorf("Current operating system is not supported ")
	}
	fmt.Println("app_home_path: ", app_home_path)

	app_keystore_path = app_home_path + "/keystore"
	app_data_path = app_home_path + "/data"
	app_log_path = app_home_path + "/log"
	app_balance_path = app_home_path + "/balance"

	subdirectory := []string{app_keystore_path, app_data_path, app_log_path, app_balance_path}

	fmt.Println( "os.ModePerm:: ",os.ModePerm)
	if _, err := os.Stat(app_home_path); os.IsNotExist(err) {
		if err = os.MkdirAll(app_home_path, os.ModePerm); err != nil {
			return fmt.Errorf("Application folder initialization failed，err=[%v] ", err)
		}
		for _, dir := range subdirectory {
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return fmt.Errorf("Application folder initialization failed，err=[%v] ", err)
			}
		}
	} else {
		for _, dir := range subdirectory {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err = os.MkdirAll(dir, os.ModePerm); err != nil {
					return fmt.Errorf("Application folder initialization failed，err=[%v] ", err)
				}
			}
		}
	}
	return nil
}

func GetPath(folder string) string {
	return app_home_path + folder
}

func GetLogPath() string {
	return app_log_path
}

func GetKeystorePath() string {
	return app_keystore_path
}

func GetDataPath() string {
	return app_data_path
}

func GetBalancePath() string {
	return app_balance_path
}

func (app *App) GetHomePath() string {
	return app_home_path
}

// Home returns the home directory for the executing user.
//
// This uses an OS-specific method for discovering the home directory.
// An error is returned if a home directory cannot be detected.
func Home() (string, error) {
	user, err := user.Current()
	if nil == err {
		return user.HomeDir, nil
	}
	// cross compile support
	if "windows" == runtime.GOOS {
		return homeWindows()
	}

	// Unix-like system, so just assume Unix
	return homeUnix()
}

func homeUnix() (string, error) {
	// First prefer the HOME environmental variable
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	// If that fails, try the shell
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}
