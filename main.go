package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

type ProxyPool struct {
	proxies []string
	urls    []string
	mutex   sync.Mutex
	stop    chan bool
	ticker  *time.Ticker
}

func (p *ProxyPool) Start() {
	p.stop = make(chan bool)
	p.ticker = time.NewTicker(30 * time.Second) //change '30' to any second u want. remind that if u run script 4 the fist time u need to wait as same as time that u set before it gonna show in browser

	go func() {
		for {
			select {
			case <-p.ticker.C:
				p.updateProxies()
			case <-p.stop:
				return
			}
		}
	}()
}

func (p *ProxyPool) GetProxies() []string {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.proxies
}

func (p *ProxyPool) Stop() {
	p.ticker.Stop()
	p.stop <- true
}

func (p *ProxyPool) updateProxies() {
	proxies, err := getProxies(p.urls)
	if err != nil {
		fmt.Println(err)
		return
	}

	p.mutex.Lock()
	p.proxies = removeDuplicateProxies(proxies)
	p.mutex.Unlock()
}
func getProxies(urls []string) ([]string, error) {
	var proxies []string
	for _, url := range urls {
		if url == "https://free-proxy-list.net/" {
			response, err := http.Get(url)
			if err != nil {
				return nil, err
			}
			defer response.Body.Close()
			html, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return nil, err
			}
			re := regexp.MustCompile(`<tr><td>(\d+.\d+.\d+.\d+)</td><td>(\d+)</td>`)
			matches := re.FindAllStringSubmatch(string(html), -1)
			for _, match := range matches {
				proxies = append(proxies, fmt.Sprintf("%s:%s", match[1], match[2]))
			}
		} else {
			response, err := http.Get(url)
			if err != nil {
				return nil, err
			}
			defer response.Body.Close()
			html, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return nil, err
			}
			lines := strings.Split(string(html), "\n")
			proxies = append(proxies, lines...)
		}
	}

	return proxies, nil
}

func removeDuplicateProxies(proxies []string) []string {
	uniqueProxies := make(map[string]struct{})
	for _, proxy := range proxies {
		uniqueProxies[proxy] = struct{}{}
	}
	var result []string
	for proxy := range uniqueProxies {
		result = append(result, proxy)
	}
	return result
}
func main() {
	proxyPool := &ProxyPool{
		urls: []string{
			"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt",
			"https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=10000&country=all&ssl=all&anonymity=all",
			"https://free-proxy-list.net/",
			"https://raw.githubusercontent.com/rdavydov/proxy-list/main/proxies_anonymous/http.txt",
			"https://raw.githubusercontent.com/MuRongPIG/Proxy-Master/main/http.txt",
			"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/master/http.txt",
			"https://raw.githubusercontent.com/MuRongPIG/Proxy-Master/main/http.txt",
			"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/http.txt",
			"https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/txt/proxies-http.txt",
		},
	}
	proxyPool.Start()

	http.HandleFunc("/proxies", func(w http.ResponseWriter, r *http.Request) {
		proxies := proxyPool.GetProxies()
		for _, proxy := range proxies {
			fmt.Fprintln(w, proxy)
		}
	})

	http.ListenAndServe(":8080", nil)
}
