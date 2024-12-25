package rpc_core

type tableDataType []map[string]interface{}

type CmdResultType struct {
	Cmd          string        `json:"cmd"`
	TableData    tableDataType `json:"table_data"`
	RowsAffected int64         `json:"rows_affected"`
	ErrorMsg     string        `json:"error_msg"`
}

type OneAddressResultType struct {
	Address    string          `json:"address"`
	CmdResults []CmdResultType `json:"cmd_results"`
	ErrorMsg   string          `json:"error_msg"`
}
