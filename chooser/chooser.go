/*
Binary chooser provides a web interface to choose a VPN from the list of
http://www.vpngate.net, create the OpenVPN config and restart OpenVPN daemon.
*/
package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"sort"
	"time"

	"github.com/StalkR/vpngate"
)

var (
	list    []*vpngate.VPN
	refresh = make(chan chan bool)
	current vpngate.VPN
)

func fmtNumber(n int) string {
	f := func(n, unit int) string {
		f := "%.0f"
		if n < 10*unit {
			f = "%.2f"
		} else if n < 100*unit {
			f = "%.1f"
		}
		return fmt.Sprintf(f, float64(n)/float64(unit))
	}
	switch {
	case n >= 1<<50:
		return f(n, 1<<50) + "P"
	case n >= 1<<40:
		return f(n, 1<<40) + "T"
	case n >= 1<<30:
		return f(n, 1<<30) + "G"
	case n >= 1<<20:
		return f(n, 1<<20) + "M"
	case n >= 1<<10:
		return f(n, 1<<10) + "K"
	}
	return fmt.Sprintf("%v", n)
}

func fmtUptime(d time.Duration) string {
	f := func(d time.Duration, unit string) string {
		plural := ""
		if d > 1 {
			plural = "s"
		}
		return fmt.Sprintf("%d %s%s", d, unit, plural)
	}
	day := 24 * time.Hour
	week := 7 * day
	year := 365 * day
	switch {
	case d > year:
		return f(d/year, "year")
	case d > week:
		return f(d/week, "week")
	case d > day:
		return f(d/day, "day")
	case d > time.Hour:
		return f(d/time.Hour, "hour")
	case d > time.Minute:
		return f(d/time.Minute, "minute")
	}
	return f(d/time.Second, "second")
}

var indexTmpl = template.Must(template.New("list.html").Funcs(
	template.FuncMap{
		"fmtNumber": fmtNumber,
		"fmtUptime": fmtUptime,
	}).ParseFiles("templates/list.html"))

func handleIndex(w http.ResponseWriter, req *http.Request) {
	v := struct {
		List    ByScore
		Current vpngate.VPN
	}{
		List:    ByScore(list),
		Current: current,
	}
	sort.Sort(sort.Reverse(v.List))
	err := indexTmpl.Execute(w, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ByScore implements sort.Interface to sort a list of VPNs by their score.
type ByScore []*vpngate.VPN

func (e ByScore) Len() int           { return len(e) }
func (e ByScore) Less(i, j int) bool { return e[i].Score < e[j].Score }
func (e ByScore) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

func refreshEvery(d time.Duration) error {
	t := time.NewTimer(d)
	for {
		var done chan bool
		select {
		case done = <-refresh:
		case <-t.C:
		}
		if r, err := vpngate.Get(http.DefaultClient); err == nil {
			list = r
		} else {
			log.Print("error refresh: ", err)
		}
		if done != nil {
			done <- true
		}
		t.Reset(d)
	}
}

func handleRefresh(w http.ResponseWriter, req *http.Request) {
	done := make(chan bool)
	refresh <- done
	<-done
	http.Redirect(w, req, "/", http.StatusFound)
}

func handleChoose(w http.ResponseWriter, req *http.Request) {
	h := req.URL.Query().Get("hostname")
	var s *vpngate.VPN
	for _, v := range list {
		if v.Hostname == h {
			s = v
			break
		}
	}
	if s == nil {
		http.NotFound(w, req)
		return
	}
	err := ioutil.WriteFile("/etc/openvpn/vpngate.conf", []byte(s.OpenVPN()), 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = exec.Command("/etc/init.d/openvpn", "restart").Output()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	current = *s
	http.Redirect(w, req, "/", http.StatusFound)
}

func main() {
	go refreshEvery(6 * time.Hour)
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/refresh", handleRefresh)
	http.HandleFunc("/choose", handleChoose)
	http.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("static"))))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
