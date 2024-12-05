package query_priv

import (
	"context"
	"dbm-services/common/go-pubpkg/errno"
	"dbm-services/mysql/priv-service/service"
	v2 "dbm-services/mysql/priv-service/service/v2/internal"
	"fmt"
	"slices"
	"strings"
	"sync"
)

func (m *GetPrivPara) GetUserList() ([]string, int, error) {
	err := m.validate()
	if err != nil {
		return nil, 0, err
	}

	var errChan = make(chan error)
	var retChan = make(chan []string)

	go func() {
		wg := sync.WaitGroup{}

		for _, item := range m.ImmuteDomains {
			wg.Add(1)

			err := m.limiter.Wait(context.Background())
			if err != nil {
				errChan <- err
				wg.Done()
				return
			}

			go func(item string) {
				defer func() {
					wg.Done()
				}()

				domainInfo, err := service.GetCluster(*m.ClusterType, service.Domain{EntryName: item})
				if err != nil {
					errChan <- err
					return
				}

				var users []string
				switch domainInfo.ClusterType {
				case v2.ClusterTypeTenDBSingle:
					users, err = usersFromTenDBSingle(&domainInfo, m.Ips)
				case v2.ClusterTypeTenDBHA:
					users, err = usersFromTenDBHA(&domainInfo, m.Ips)
				case v2.ClusterTypeTenDBCluster:
					users, err = usersFromTenDBCluster(&domainInfo, m.Ips)
				default:
					errChan <- fmt.Errorf("unknown cluster type: %s", domainInfo.ClusterType)
					return
				}

				if err != nil {
					errChan <- err
					return
				}

				retChan <- users
				return
			}(item)
		}

		wg.Wait()
	}()

	var userList []string
	var errList []string
	count := 0
	for {
		select {
		case err := <-errChan:
			count++
			errList = append(errList, err.Error())
		case users := <-retChan:
			count++
			userList = append(userList, users...)
		}
		if count == len(m.ImmuteDomains) {
			break
		}
	}

	close(errChan)
	close(retChan)

	if len(errList) > 0 {
		return userList, len(userList), errno.QueryPrivilegesFail.Add("\n" + strings.Join(errList, "\n"))
	}
	return userList, count, nil
}

func usersFromTenDBHA(domainInfo *service.Instance, clientIps []string) ([]string, error) {
	if domainInfo.EntryRole == v2.EntryRoleMasterEntry && !domainInfo.PaddingProxy {
		// ToDo Proxies 可能为空
		queryAddr := fmt.Sprintf("%s:%d", domainInfo.Proxies[0].IP, domainInfo.Proxies[0].Port)
		_, users, _, err := service.ProxyWhiteList(
			queryAddr,
			domainInfo.BkCloudId,
			clientIps,
			nil,
			"",
		)
		if err != nil {
			return nil, err
		}
		return users, nil
	} else {
		instanceRole := v2.InstanceRoleBackendSlave
		if domainInfo.EntryRole == v2.EntryRoleMasterEntry {
			instanceRole = v2.InstanceRoleBackendMaster
		}

		idx := slices.IndexFunc(domainInfo.Storages, func(s service.Storage) bool {
			return s.InstanceRole == instanceRole
		})
		// ToDo idx 可能 < 0
		queryAddr := fmt.Sprintf("%s:%d", domainInfo.Storages[idx].IP, domainInfo.Storages[idx].Port)
		_, users, _, err := service.MysqlUserList(
			queryAddr,
			domainInfo.BkCloudId,
			clientIps,
			nil,
			"",
		)
		if err != nil {
			return nil, err
		}
		return users, nil
	}
}

func usersFromTenDBSingle(domainInfo *service.Instance, clientIps []string) ([]string, error) {
	idx := slices.IndexFunc(domainInfo.Storages, func(s service.Storage) bool {
		return s.InstanceRole == v2.InstanceRoleOrphan
	})
	queryAddr := fmt.Sprintf("%s:%d", domainInfo.Storages[idx].IP, domainInfo.Storages[idx].Port)
	_, users, _, err := service.MysqlUserList(
		queryAddr,
		domainInfo.BkCloudId,
		clientIps,
		nil,
		"",
	)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func usersFromTenDBCluster(domainInfo *service.Instance, clientIps []string) ([]string, error) {
	var queryAddr string
	if domainInfo.EntryRole == v2.EntryRoleMasterEntry {
		queryAddr = fmt.Sprintf(domainInfo.SpiderMaster[0].IP, domainInfo.SpiderMaster[0].Port)
	} else {
		queryAddr = fmt.Sprintf(domainInfo.SpiderSlave[0].IP, domainInfo.SpiderSlave[0].Port)
	}

	_, users, _, err := service.MysqlUserList(
		queryAddr,
		domainInfo.BkCloudId,
		clientIps,
		nil,
		"")
	if err != nil {
		return nil, err
	}
	return users, nil
}
