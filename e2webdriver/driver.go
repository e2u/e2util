package e2webdriver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/e2u/e2util/e2exec"
	"github.com/e2u/e2util/e2http"
	"github.com/e2u/e2util/e2os"
	"github.com/klauspost/compress/zip"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

var (
	fileNameMap = map[string]string{
		"darwin_amd64": "chromedriver_mac64.zip",
		"darwin_arm64": "chromedriver_mac_arm64.zip",
		"linux_amd64":  "chromedriver_linux64.zip",
		// "linux_arm":     "",
		// "linux_arm64":   "",
		"windows_i386":  "chromedriver_win32.zip",
		"windows_amd64": "chromedriver_win32.zip",
	}
)

// const (
//	VersionStable = "Stable"
//	VersionBeta   = "Beta"
//	VersionDev    = "Dev"
//	VersionCanary = "Canary"
//)

// https://storage.googleapis.com/storage/v1/b/chromedriver/o?maxResults=50&fields=items(id,selfLink,mediaLink,name,timeCreated)&chromedriver=chromedriver&matchGlob=**/114.0.5735.90/chromedriver_*.zip
// https://googlechromelabs.github.io/chrome-for-testing/latest-versions-per-milestone-with-downloads.json

func Install(ctx context.Context, binDir ...string) (string, error) {
	installDir := os.TempDir()
	if len(binDir) > 0 {
		installDir = binDir[0]
	}
	var extName string
	if runtime.GOOS == "windows" {
		extName = ".exe"
	}
	checkPath := filepath.Clean(filepath.Join(installDir, strings.ReplaceAll(e2exec.Must(getCurrentOSFileName()), ".zip", ""), "chromedriver"+extName))
	if e2os.FileExists(checkPath) {
		return checkPath, nil
	}
	logrus.Infof("chromedriver do not exists, start install")

	ver := e2exec.Must(getLatestVersions(ctx))
	u, err := buildDownloadUrl(ctx, ver)
	if err != nil {
		return "", err
	}

	path, err := downloadAndUnzip(ctx, u, filepath.Clean(installDir))
	if err != nil {
		return "", err
	}
	return path, nil
}

// https://googlechromelabs.github.io/chrome-for-testing/latest-versions-per-milestone-with-downloads.json

// https://googlechromelabs.github.io/chrome-for-testing/113.0.5672.63.json
// func downloadDriver(ctx context.Context, localDir string) error {
//	e2http.Builder(ctx).URL("https://googlechromelabs.github.io/chrome-for-testing/latest-versions-per-milestone-with-downloads.json")
//	return nil
//}

// getVersions
// url: https://googlechromelabs.github.io/chrome-for-testing/last-known-good-versions.json
// response example:
// {"timestamp":"2024-06-21T20:09:39.687Z","channels":{"Stable":{"channel":"Stable","version":"126.0.6478.63","revision":"1300313"},"Beta":{"channel":"Beta","version":"127.0.6533.17","revision":"1313161"},"Dev":{"channel":"Dev","version":"127.0.6523.4","revision":"1310990"},"Canary":{"channel":"Canary","version":"128.0.6550.0","revision":"1317892"}}}
func getLatestVersions(ctx context.Context) (string, error) {
	h := e2http.Builder(ctx).URL("https://chromedriver.storage.googleapis.com/LATEST_RELEASE").Do()
	if len(h.Errors()) > 0 || h.StatusCode() != http.StatusOK {
		return "", errors.Join(h.Errors()...)
	}
	if h.BodyString() != "" {
		return h.BodyString(), nil
	}
	return "", fmt.Errorf("cannot find latest versions")
}

func getCurrentOSFileName() (string, error) {
	fileName, ok := fileNameMap[runtime.GOOS+"_"+runtime.GOARCH]
	if !ok {
		return "", fmt.Errorf("os %s arch %s not supported", runtime.GOOS, runtime.GOARCH)
	}
	return fileName, nil
}

func buildDownloadUrl(ctx context.Context, version string) (string, error) {
	fileName, err := getCurrentOSFileName()
	if err != nil {
		return "", err
	}

	apiUrl := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/chromedriver/o?fields=items(id,selfLink,mediaLink,name,timeCreated)&matchGlob=**/%s/%s", version, fileName)
	h := e2http.Builder(ctx).URL(apiUrl).Do()
	if len(h.Errors()) > 0 || h.StatusCode() != http.StatusOK {
		return "", errors.Join(h.Errors()...)
	}
	var downloadUrl string
	gjson.GetBytes(h.Body(), "items").ForEach(func(key, val gjson.Result) bool {
		if v := val.Get("mediaLink"); v.Exists() && v.String() != "" {
			downloadUrl = v.String()
			return false
		}
		return true
	})
	return downloadUrl, nil
}

func downloadAndUnzip(ctx context.Context, url string, localDir string) (string, error) {
	fileName, err := getCurrentOSFileName()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(localDir), 0755); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer e2exec.MustClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	zipFile := filepath.Join(localDir, fileName)
	out, err := os.OpenFile(zipFile, os.O_CREATE|os.O_SYNC|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		logrus.Errorf("open file error=%v", err)
		return "", err
	}
	defer e2exec.MustClose(out)

	if _, err := out.ReadFrom(resp.Body); err != nil {
		logrus.Errorf("read file error=%v", err)
		return "", err
	}
	if err := out.Sync(); err != nil {
		logrus.Errorf("sync file error=%v", err)
		return "", err
	}

	zr, err := zip.OpenReader(filepath.Join(localDir, fileName))
	if err != nil {
		logrus.Errorf("new zip reader error=%v", err)
		return "", err
	}
	defer e2exec.MustClose(zr)
	var execFilePath string

	for _, f := range zr.File {
		extractDir := filepath.Join(localDir, strings.ReplaceAll(filepath.Base(zipFile), ".zip", ""))
		fPath := filepath.Join(extractDir, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(extractDir, os.ModePerm); err != nil {
				return "", err
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(fPath), os.ModePerm); err != nil {
				return "", err
			}
			outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return "", err
			}
			fr, err := f.Open()
			if err != nil {
				return "", err
			}
			e2exec.SilentError(io.Copy(outFile, fr))
			e2exec.MustClose(fr)
			e2exec.MustClose(outFile)
			if filepath.Base(fPath) == "chromedriver" || filepath.Base(fPath) == "chromedriver.exe" {
				execFilePath = fPath
			}
		}
	}
	return execFilePath, nil
}
