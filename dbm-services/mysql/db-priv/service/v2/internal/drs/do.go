package drs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type APIServerResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *drsClient) do(method string, path string, body []byte) (*APIServerResponse, error) {
	c.baseURL = viper.GetString("dbRemoteService")

	slog.Info(
		"drs do",
		slog.String("method", method),
		slog.String("path", path),
		slog.String("body", string(body)),
		slog.String("base url", c.baseURL),
	)

	endPoint, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		slog.Error("drs do", slog.String("err", err.Error()))
		return nil, errors.Wrapf(err, "drs do")
	}
	slog.Info("drs do", slog.String("url", endPoint))

	request, err := http.NewRequest(method, endPoint, bytes.NewBuffer(body))
	if err != nil {
		slog.Error("drs new request", slog.String("err", err.Error()))
		return nil, errors.Wrapf(err, "drs new request")
	}
	request.Header.Set("Content-Type", "application/json")
	bkAuth := fmt.Sprintf(
		`{"bk_app_code": %s, "bk_app_secret": %s}`,
		viper.GetString("bk_app_code"),
		viper.GetString("bk_app_secret"),
	)
	request.Header.Set("x-bkapi-authorization", bkAuth)

	cookieAppCode := http.Cookie{
		Name:   "bk_app_code",
		Path:   "/",
		Value:  viper.GetString("bk_app_code"),
		MaxAge: 86400,
	}
	cookieAppSecret := http.Cookie{
		Name:   "bk_app_secret",
		Path:   "/",
		Value:  viper.GetString("bk_app_secret"),
		MaxAge: 86400,
	}
	request.AddCookie(&cookieAppCode)
	request.AddCookie(&cookieAppSecret)

	var resp *http.Response
	//for i := 0; i < 5; i++ {
	//if resp != nil && resp.Body != nil {
	//	_ = resp.Body.Close()
	//}

	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		slog.Error("drs do", slog.String("err", err.Error()))
		return nil, errors.Wrapf(err, "drs do")
		//if i == 4 {
		//	return nil, errors.Wrapf(err, "drs do")
		//}
		//slog.Info("drs do retry")
		//if resp != nil && resp.Body != nil {
		//	_ = resp.Body.Close()
		//}
		//time.Sleep(time.Second)
		//continue
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error(
			"drs do",
			slog.String("http status", resp.Status),
			slog.Int("status_code", resp.StatusCode),
		)
		return nil, errors.Errorf("%s: %d", resp.Status, resp.StatusCode)
		//if i == 4 {
		//	return nil, errors.Errorf("%s: %d", resp.Status, resp.StatusCode)
		//}
		//slog.Info("drs do retry")
		//if resp.Body != nil {
		//	_ = resp.Body.Close()
		//}
		//time.Sleep(time.Second)
		//continue
	}
	//}
	defer func() {
		_ = resp.Body.Close()
	}()

	slog.Info(
		"drs do",
		slog.String("http status", resp.Status),
		slog.Int("http code", resp.StatusCode),
	)

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("drs do", slog.String("err", err.Error()))
		return nil, errors.Wrapf(err, "drs do")
	}

	res := APIServerResponse{}
	err = json.Unmarshal(b, &res)
	if err != nil {
		slog.Error("drs do", slog.String("err", err.Error()))
		return nil, errors.Wrapf(err, "drs do")
	}

	if res.Code != 0 {
		err = errors.Errorf("%d: %s", res.Code, res.Message)
		slog.Error(
			"drs do",
			slog.String("err", err.Error()),
		)
		return nil, errors.Wrapf(err, "drs do")
	}

	return &res, nil
}
