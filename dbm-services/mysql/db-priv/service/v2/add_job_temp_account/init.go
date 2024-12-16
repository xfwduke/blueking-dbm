package add_job_temp_account

type AddJobTempAccount struct {
	User      string   `json:"user"`
	Psw       string   `json:"psw"`
	BkCloudId int64    `json:"bk_cloud_id"`
	Addresses []string `json:"addresses"`
}
