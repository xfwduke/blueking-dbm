package add_priv_without_account_rule

type AddPrivWithoutAccountRule struct {
	User      string   `json:"user"`
	Psw       string   `json:"psw"`
	BkCloudId int64    `json:"bk_cloud_id"`
	Addresses []string `json:"addresses"`
}
