package nginx_updater

import (
	"bufio"
	"os"
)

const nginxAddrFile = "/home/mysql/nginx_conf/address.list"

func Updater() {
	addrs, err := readAddr()
	if err != nil {

	}

	newAddrs, err := queryNewAddr(addrs)

	f, err := os.OpenFile(nginxAddrFile, os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
	}
	defer func() {
		_ = f.Close()
	}()

	for _, ad := range newAddrs {
		_, _ = f.WriteString(ad + "\n")
	}
}

func readAddr() (res []string, err error) {
	f, err := os.Open(nginxAddrFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		res = append(res, line)
	}
	err = scanner.Err()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func queryNewAddr(addrs []string) (res []string, err error) {
	return
}
