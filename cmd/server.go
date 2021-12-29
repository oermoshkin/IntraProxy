package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

type RegexpHandler struct {
	routes []*route
}

func (h *RegexpHandler) Handler(pattern *regexp.Regexp, handler http.Handler) {
	h.routes = append(h.routes, &route{pattern, handler})
}

func (h *RegexpHandler) HandleFunc(pattern *regexp.Regexp, handler func(http.ResponseWriter, *http.Request)) {
	h.routes = append(h.routes, &route{pattern, http.HandlerFunc(handler)})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if route.pattern.MatchString(r.URL.Path) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

//MyServer Запуск веб сервера
func MyServer() {
	handler := &RegexpHandler{}

	allReq, _ := regexp.Compile(".*")
	handler.HandleFunc(allReq, MyHandler)
	ServerIP := strings.Join([]string{Config.Server.Host, Config.Server.Port}, ":")
	log.Println("Started server", ServerIP)
	log.Fatal(http.ListenAndServeTLS(ServerIP, Config.Server.Cert, Config.Server.PrivKey, handler))
}

//MyHandler Обработчик запросов. Вся магия тут
func MyHandler(w http.ResponseWriter, r *http.Request) {
	var Host string

	if r.Header.Get("Upgrade") == "websocket" {
		NewWS(w, r)
		return
	}

	//Меняем Host для запроса на сервера IntraDesk
	switch r.Host {
	case Config.Proxy.Server:
		Host = Config.Origin.Server
	case Config.Proxy.Login:
		Host = Config.Origin.Login
	case Config.Proxy.ApiGW:
		Host = Config.Origin.ApiGW
	case Config.Proxy.Doc:
		Host = Config.Origin.Doc
	default:
		return
	}

	var RemoteIP string

	if r.Header.Get("X-Real-IP") != "" {
		RemoteIP = r.Header.Get("X-Real-IP")
	} else {
		RemoteIP = r.RemoteAddr
	}

	Header := r.Header

	//Мне лень было распаковывать gzip, поэтому я просто удаляю этот заголовок.
	Header.Del("Accept-Encoding")

	//Готовим URL для запроса на сервера вендора
	var url string

	if r.URL.RawQuery != "" {
		//Меняем прокси host на оригинальный в параметрах запроса
		query := strings.Replace(r.URL.RawQuery, Config.Proxy.Server, Config.Origin.Server, -1)
		url = strings.Join([]string{"https://", Host, r.URL.Path, "?", query}, "")
	} else {
		url = strings.Join([]string{"https://", Host, r.URL.Path}, "")
	}

	client := &http.Client{}
	req, _ := http.NewRequest(r.Method, url, r.Body)

	req.Header = Header.Clone()
	redirect := false

	//Останавливаем редирект, чтобы запрос не продолжал выполнение
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		redirect = true
		return errors.New("Detected Redirect")
	}

	var NewLoc string
	resp, err := client.Do(req)
	if err != nil {
		if !redirect {
			log.Println("Request error: ", err)
			return
		} else {
			//Произошел редирект. Меняем Location для пользователя.
			replacer := strings.NewReplacer(Config.Origin.Server, Config.Proxy.Server,
				Config.Origin.Login, Config.Proxy.Login)
			NewLoc = replacer.Replace(resp.Header.Get("Location"))
		}
	}

	//Header. Выставляем заголовки для пользователя
	RespHeader := resp.Header
	RespHeader.Del("Content-Length")
	RespHeader.Del("Location")
	for name, values := range RespHeader {
		w.Header()[name] = values
	}

	if redirect {
		w.Header().Set("Location", NewLoc)
	}

	defer resp.Body.Close()

	respByte, _ := ioutil.ReadAll(resp.Body)

	var data []byte

	//Если ответ это JSON. То в нем нам нужно изменить оригинальные URL на прокси URL
	if (strings.ToLower(resp.Header.Get("content-type")) == "application/json; charset=utf-8") ||
		(strings.ToLower(resp.Header.Get("content-type")) == "application/jwk-set+json; charset=utf-8") {
		respString := string(respByte)
		replacer := strings.NewReplacer(Config.Origin.Login, Config.Proxy.Login,
			Config.Origin.ApiGW, Config.Proxy.ApiGW,
			Config.Origin.Doc, Config.Proxy.Doc)
		newResp := replacer.Replace(respString)
		data = []byte(newResp)
	} else {
		data = respByte
	}

	//Меняем в js параметр this.skipIssuerCheck на True
	matched, _ := regexp.MatchString(`vendor-.*\.js$`, r.URL.Path)
	if matched {
		data = bytes.Replace(data, []byte("this.skipIssuerCheck=!1"), []byte("this.skipIssuerCheck=1"), 1)
	}

	w.WriteHeader(resp.StatusCode)
	if len(data) != 0 {
		_, err = w.Write(data)
		if err != nil {
			log.Println("Error send data:", err)
			log.Println("StatusCode:", resp.StatusCode)
			log.Println("Data:", string(data))
		}
	}

	log.Printf("%s: %s %s %s%s %d", RemoteIP, r.Proto, r.Method, r.Host, r.URL.Path, resp.StatusCode)
}
