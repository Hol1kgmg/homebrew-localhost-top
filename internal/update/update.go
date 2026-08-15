package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Hol1kgmg/homebrew-localhost-top/internal/i18n"
)

const releasesURL = "https://api.github.com/repos/Hol1kgmg/homebrew-localhost-top/releases/latest"

// Info はアップデートチェックの結果を表す。
type Info struct {
	Available bool
	Current   string
	Latest    string
}

type releaseResponse struct {
	TagName string `json:"tag_name"`
}

// Check はGitHub Releasesの最新タグとcurrentを比較する。
func Check(ctx context.Context, current string) (Info, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releasesURL, nil)
	if err != nil {
		return Info{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "localhost-top")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return Info{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Info{}, fmt.Errorf(i18n.T("github_api_request_failed"), resp.StatusCode)
	}

	var rel releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return Info{}, err
	}

	newer, err := isNewer(rel.TagName, current)
	if err != nil {
		return Info{}, err
	}

	return Info{
		Available: newer,
		Current:   current,
		Latest:    rel.TagName,
	}, nil
}

// isNewer はlatestがcurrentより新しいバージョンかを判定する。
// "vX.Y.Z"形式を想定し、パースに失敗した場合はエラーを返す。
func isNewer(latest, current string) (bool, error) {
	l, err := parseVersion(latest)
	if err != nil {
		return false, err
	}
	c, err := parseVersion(current)
	if err != nil {
		return false, err
	}

	for i := 0; i < 3; i++ {
		if l[i] != c[i] {
			return l[i] > c[i], nil
		}
	}
	return false, nil
}

func parseVersion(v string) ([3]int, error) {
	var out [3]int
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return out, fmt.Errorf(i18n.T("invalid_version_format"), v)
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return out, fmt.Errorf(i18n.T("invalid_version_format"), v)
		}
		out[i] = n
	}
	return out, nil
}
