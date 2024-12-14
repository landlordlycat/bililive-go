package xiaohongshu

import (
	"fmt"
	"github.com/hr3lxphr6j/bililive-go/src/live"
	"github.com/hr3lxphr6j/bililive-go/src/live/internal"
	"github.com/hr3lxphr6j/requests"
	"github.com/tidwall/gjson"
	"net/http"
	"net/url"
	"strings"
)

const (
	domain = "www.xiaohongshu.com"
	cnName = "小红书"

	roomApiUrl = "https://www.xiaohongshu.com/api/sns/red/live/app/v1/ecology/outside/share_info"
	streamUrl  = "http://live-play.xhscdn.com/live"

	userAgent = "Mozilla/5.0 (Linux; Android 11; SAMSUNG SM-G973U) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/14.2 Chrome/87.0.4280.141 Mobile Safari/537.36"
)

func init() {
	live.Register(domain, new(builder))
}

type builder struct{}

func (b *builder) Build(url *url.URL, opt ...live.Option) (live.Live, error) {
	return &Live{
		BaseLive: internal.NewBaseLive(url, opt...),
	}, nil
}

type Live struct {
	internal.BaseLive
}

func (l *Live) GetInfo() (info *live.Info, err error) {
	headers := map[string]interface{}{
		"User-Agent":      userAgent,
		"Accept":          "application/json, text/plain, */*",
		"Accept-Language": "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2",
		"Referer":         "https://www.xiaohongshu.com/hina/livestream/568979931846654360",
	}
	cookies := l.Options.Cookies.Cookies(l.Url)
	cookieKVs := make(map[string]string)
	for _, item := range cookies {
		cookieKVs[item.Name] = item.Value
	}

	pathParts := strings.Split(l.Url.Path, "/")
	roomId := pathParts[len(pathParts)-1]

	resp, err := requests.Get(
		roomApiUrl,
		requests.Query("room_id", roomId),
		requests.Cookies(cookieKVs),
		requests.Headers(headers),
	)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, live.ErrRoomNotExist
	}

	body, err := resp.Bytes()
	if err != nil {
		return nil, err
	}
	if gjson.GetBytes(body, "code").Int() != 0 {
		return nil, live.ErrRoomNotExist
	}

	mobileUrl := fmt.Sprintf("%s/%s.flv", streamUrl, roomId)
	response, err := requests.Head(mobileUrl,
		requests.Cookies(cookieKVs),
		requests.Headers(headers),
	)
	if err != nil {
		return nil, err
	}

	info = &live.Info{
		Live:     l,
		HostName: gjson.GetBytes(body, "data.host_info.nickname").String(),
		RoomName: gjson.GetBytes(body, "data.room.name").String(),
		// 小红书直播间开没开播，status都为0
		//Status:   gjson.GetBytes(body, "data.room.status").Int() == 0
		Status: response.StatusCode == http.StatusOK,
	}

	return
}

func (l *Live) GetStreamUrls() (us []*url.URL, err error) {
	roomUrl := l.Url.String()
	urlParts := strings.Split(roomUrl, "?")
	pathParts := strings.Split(urlParts[0], "/")
	roomId := pathParts[len(pathParts)-1]

	// 不用appuid也能播
	//appUIDRegex := regexp.MustCompile(`appuid=(.*?)&`)
	//appUIDMatch := appUIDRegex.FindStringSubmatch(roomUrl)
	//if len(appUIDMatch) < 2 {
	//	return nil, live.ErrRoomUrlIncorrect
	//}
	//appuid := appUIDMatch[1]
	//flvURL := fmt.Sprintf("http://live-play.xhscdn.com/live/%s.flv?uid=%s", roomId, appuid)
	mobileUrl := fmt.Sprintf("%s/%s.flv", streamUrl, roomId)
	u, err := url.Parse(mobileUrl)
	if err != nil {
		return nil, err
	}

	return []*url.URL{u}, nil
}

func (l *Live) GetPlatformCNName() string {
	return cnName
}
