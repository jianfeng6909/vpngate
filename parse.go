package vpngate

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// parseList parses vpngate CSV.
func parseList(f io.Reader) ([]*VPN, error) {
	// First read to ignore invalid lines (first "*vpnservers", last "*").
	r := bufio.NewReader(f)
	var buf bytes.Buffer
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if strings.HasPrefix(line, "*") {
			continue
		}
		buf.WriteString(line)
	}

	c := csv.NewReader(&buf)
	_, err := c.Read() // Ignore CSV header.
	if err != nil {
		return nil, err
	}
	var s []*VPN
	for {
		vpn, err := parseRecord(c)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		s = append(s, vpn)
	}
	return s, nil
}

// parseRecord parses a CSV record into a VPN.
func parseRecord(c *csv.Reader) (*VPN, error) {
	d, err := c.Read()
	if err != nil {
		return nil, err
	}
	if len(d) != 15 {
		return nil, fmt.Errorf("got %d columns, want 15", len(d))
	}

	v := &VPN{}
	v.Hostname = d[0]
	v.IP = d[1]
	v.Score, _ = strconv.Atoi(d[2])
	v.Ping, _ = strconv.Atoi(d[3])
	v.Speed, _ = strconv.Atoi(d[4])
	v.Country = d[5]
	v.CountryShort = d[6]
	v.Sessions, _ = strconv.Atoi(d[7])
	v.Uptime, _ = strconv.Atoi(d[8])
	v.Users, _ = strconv.Atoi(d[9])
	v.Traffic, _ = strconv.Atoi(d[10])
	v.LogType = d[11]
	v.Operator = d[12]
	v.Message = d[13]

	b, err := base64.StdEncoding.DecodeString(d[14])
	if err != nil {
		return nil, err
	}
	r := bytes.NewBuffer(b)
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		words := strings.Split(strings.TrimSpace(line), " ")
		switch {
		case len(words) < 2:
			continue
		case words[0] == "proto":
			v.Proto = words[1]
		case words[0] == "cipher":
			v.Cipher = words[1]
		case words[0] == "auth":
			v.Auth = words[1]
		case len(words) < 3:
			continue
		case words[0] == "remote":
			if v.IP != words[1] {
				return nil, fmt.Errorf("inconsistent IP: got %s, want %s", words[1], v.IP)
			}
			v.Port, _ = strconv.Atoi(words[2])
		}
	}

	if v.Proto == "" || v.IP == "" || v.Port == 0 || v.Cipher == "" || v.Auth == "" {
		return nil, fmt.Errorf("invalid config")
	}

	return v, nil
}
