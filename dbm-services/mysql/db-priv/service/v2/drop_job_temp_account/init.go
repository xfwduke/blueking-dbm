package drop_job_temp_account

type DropJobTempAccount struct {
	BkCloudId int64    `db:"bk_cloud_id"`
	User      string   `json:"user"`
	Addresses []string `json:"addresses"`
}
