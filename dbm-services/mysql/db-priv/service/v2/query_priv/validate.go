package query_priv

import "dbm-services/common/go-pubpkg/errno"

// CheckPara 查询权限入参检查
func (m *GetPrivPara) validate() error {
	if m.ClusterType == nil {
		return errno.ClusterTypeIsEmpty
	}
	if len(m.Ips) == 0 {
		return errno.IpRequired
	}
	if len(m.ImmuteDomains) == 0 {
		return errno.DomainRequired
	}
	return nil
}
