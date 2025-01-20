/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-DB管理系统(BlueKing-BK-DBM) available.
 * Copyright (C) 2017-2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package mysqlutil

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync"

	"dbm-services/common/go-pubpkg/mysqlcomm"

	"github.com/samber/lo"

	"dbm-services/common/go-pubpkg/cmutil"
	"dbm-services/common/go-pubpkg/logger"
	"dbm-services/mysql/db-tools/dbactuator/pkg/core/cst"
	"dbm-services/mysql/db-tools/dbactuator/pkg/util"
	"dbm-services/mysql/db-tools/dbactuator/pkg/util/osutil"
)

var dumpCompleteReg = regexp.MustCompile("Dump completed on")

// Dumper TODO
type Dumper interface {
	Dump() error
}

// MySQLDumpOption TODO
type MySQLDumpOption struct {
	DumpSchema bool
	DumpData   bool
	// NoData       bool
	AddDropTable            bool // 默认 false 代表添加 --skip-add-drop-table 选项
	NoUseDbAndWirteCreateDb bool // 备份的sql文件中不会打印 "CREATE DATABASE ..." 语句和 "USE db_name;" 语句
	NoCreateDb              bool
	NoCreateTb              bool
	DumpRoutine             bool // 默认 false 代表添加不导出存储过程,True导出存储过程
	DumpTrigger             bool // 默认 false 代表添加不导出触发器
	DumpEvent               bool // 默认 false 导出 event
	GtidPurgedOff           bool // --set-gtid-purged=OFF
	Quick                   bool
	ExtendedInsert          bool
	Force                   bool
}

type runtimectx struct {
	maxConcurrency        int
	maxResourceUsePercent int
}

// MySQLDumper  use mysqldump param
type MySQLDumper struct {
	MySQLDumpOption
	DumpDir      string // 备份到哪个目录
	DbBackupUser string
	DbBackupPwd  string
	Ip           string
	Port         int
	Charset      string
	DumpCmdFile  string // mysqldump 的绝对路径
	DbNames      []string
	Tables       []string
	IgnoreTables []string
	IsMaster     bool
	Where        string
	runtimectx
}

// GetDumpFileInfo 获取dump db 对应输出的文件信息
func (m MySQLDumper) GetDumpFileInfo() map[string]string {
	dumpMap := make(map[string]string)
	for _, db := range m.DbNames {
		dumpMap[db] = fmt.Sprintf("%s.sql", db)
	}
	return dumpMap
}

// Dump SplitByDb OneByOne 按照每个db 分别导出不同的文件，可控制并发
//
//	@receiver m
//	@return err
func (m MySQLDumper) Dump() (err error) {
	m.init()
	var wg sync.WaitGroup
	var errs []error
	concurrencyControl := make(chan struct{}, m.maxConcurrency)
	dumpMap := m.GetDumpFileInfo()
	errChan := make(chan error)
	for db, outputFileName := range dumpMap {
		dumper := m
		dumper.DbNames = []string{db}
		outputFile := path.Join(m.DumpDir, outputFileName)
		concurrencyControl <- struct{}{}
		wg.Add(1)
		go func(dump MySQLDumper, db string, outputFile string) {
			defer func() {
				wg.Done()
				<-concurrencyControl
			}()
			errFile := path.Join(dump.DumpDir, fmt.Sprintf("%s.err", db))
			dumpCmd := dump.getDumpCmd(outputFile, errFile, "", false)
			logger.Info("mysqldump cmd:%s", mysqlcomm.RemovePassword(dumpCmd))
			output, err := osutil.StandardShellCommand(false, dumpCmd)
			if err != nil {
				errContent, _ := os.ReadFile(errFile)
				errChan <- fmt.Errorf("execte %s get an error:%s,%w\n errfile content:%s", dumpCmd, output, err,
					string(errContent))
				return
			}
			if err := checkDumpComplete(outputFile); err != nil {
				errContent, _ := os.ReadFile(errFile)
				errChan <- fmt.Errorf("%w\n errfile content:%s", err, string(errContent))
				return
			}
		}(dumper, db, outputFile)
	}
	go func() {
		wg.Wait()
		close(errChan)
	}()
	for err := range errChan {
		logger.Error("dump db failed: %s", err.Error())
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// MySQLDumperTogether TODO
type MySQLDumperTogether struct {
	MySQLDumper
	OutputfileName string
	UseTMySQLDump  bool // 是否使用的是自研的mysqldump,一般介质在备份目录下
	TDBCTLDump     bool // 中控专有参数
}

// checkDumpComplete  检查导出结果是否Ok
//
//	@receiver file  导出SQL文件的绝对路径
//	@return err
func checkDumpComplete(file string) (err error) {
	// 倒着读取10行
	res, err := util.ReverseRead(file, 10)
	if err != nil {
		return err
	}
	for _, l := range res {
		// 如果匹配到了，表示备份的文件oK
		if dumpCompleteReg.MatchString(l) {
			return nil
		}
	}
	return fmt.Errorf("备份文件没有匹配到Dump completed on")
}

// init 初始化运行时参数
//
//	@receiver m
func (m *MySQLDumper) init() {
	m.maxConcurrency = runtime.NumCPU() / 2
	m.maxResourceUsePercent = 50
	if m.IsMaster || m.maxConcurrency == 0 {
		// 如果是在Master Dump的话不允许开启并发
		m.maxConcurrency = 1
	}
}

// Dump Together 后面指定的 db 名字空格分隔，例如 --databases db1 db2 > just_one_file.sql
//
//	@receiver m
//	@return err
func (m *MySQLDumperTogether) Dump() (err error) {
	m.init()
	outputFile := path.Join(m.DumpDir, m.OutputfileName)
	errFile := path.Join(m.DumpDir, m.OutputfileName+".err")
	dumpOption := ""
	if m.UseTMySQLDump {
		dumpOption = m.getTMySQLDumpOption()
	}
	defer func() {
		if err != nil {
			errFileContext, e1 := osutil.ReadFileString(errFile)
			if e1 != nil {
				logger.Error("read errFile failed %s", e1.Error())
			}
			logger.Error("errFile:%s", errFileContext)
		}
	}()
	dumpCmd := m.getDumpCmd(outputFile, errFile, dumpOption, false)
	logger.Info("mysqldump cmd:%s", mysqlcomm.ClearSensitiveInformation(dumpCmd))
	output, err := osutil.StandardShellCommand(false, dumpCmd)
	if err != nil {
		return fmt.Errorf("execte %s get an error:%s,%w", dumpCmd, output, err)
	}
	if err := checkDumpComplete(outputFile); err != nil {
		logger.Error("checkDumpComplete failed %s", err.Error())
		return err
	}
	return
}

// MySQLDumperAppend 不同库表导出到同一个文件
type MySQLDumperAppend struct {
	MySQLDumper
	OutputfileName string
	DumpMap        map[string][]string
}

// Dump do dump
func (m *MySQLDumperAppend) Dump() (err error) {
	outputFile := path.Join(m.DumpDir, m.OutputfileName)
	errFile := path.Join(m.DumpDir, m.OutputfileName+".err")
	fd, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file failed %s", err.Error())
	}
	_, err = fd.WriteString("-- dump schema for dbm simulation\n")
	if err != nil {
		logger.Error("write file failed %s", err.Error())
		return err
	}
	defer func() {
		if err != nil {
			if cmutil.FileExists(errFile) {
				errMsg, errx := osutil.ReadFileString(errFile)
				if errx != nil {
					logger.Error("read errFile failed %s", errx.Error())
				}
				logger.Error("errFile contenxt:%s", errMsg)
			}
		}
	}()
	defer fd.Close()
	inputdbs := m.DbNames
	for db, tables := range m.DumpMap {
		var realdbs []string
		if lo.IsNotEmpty(db) {
			// inputdbs是实际存在的库
			// 如果dumpMap中的库不在inputdbs中，直接跳过
			// 让错误在模拟执行中体现
			if !slices.Contains(inputdbs, db) {
				logger.Warn("db %s not in inputdbs %v", db, inputdbs)
				continue
			}
			realdbs = []string{db}
		} else {
			realdbs = inputdbs
		}
		for _, realdb := range realdbs {
			_, err = fd.WriteString(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`;\n USE `%s`;\n", realdb, realdb))
			if err != nil {
				return fmt.Errorf("write file failed %s", err.Error())
			}
			m.Tables = lo.Uniq(tables)
			m.DbNames = []string{realdb}
			dumpCmd := m.getDumpCmd(outputFile, errFile, "", true)
			logger.Info("mysqldump cmd:%s", mysqlcomm.ClearSensitiveInformation(dumpCmd))
			output, errx := osutil.StandardShellCommand(false, dumpCmd)
			if errx != nil {
				if err = dumpIsOk(errFile); err == nil {
					continue
				}
				return fmt.Errorf("execte %s get an error:%s,%w", dumpCmd, output, errx)
			}
			if err = checkDumpComplete(outputFile); err != nil {
				logger.Error("checkDumpComplete failed %s", err.Error())
				return err
			}
		}
	}
	return err
}

func dumpIsOk(errLog string) (err error) {
	fd, err := os.Open(errLog)
	if err != nil {
		return err
	}
	// ignore warning
	w := regexp.MustCompile(`[Warning] Using a password on the command line interface can be insecure`)
	r := regexp.MustCompile(`Couldn't find table:`)
	var lines []string
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		l := scanner.Text()
		if !r.MatchString(l) && lo.IsNotEmpty(l) && !w.MatchString(l) {
			lines = append(lines, l)
		}
	}
	if len(lines) > 0 {
		return fmt.Errorf("%s", strings.Join(lines, "\n"))
	}
	return
}

/*
mysqldump 参数说明：
-B --databases ：后面指定的 db 名字空格分隔，例如 --databases db1 db2 >> aaa.sql

-d, --no-data：不导出 row information，也就是不导出行数据。 只导出 schema 的时候比较常用，例如： --databases testdb -d > testdb_d.sql 。
需要注意的是带上 -B，sql 文件里面就会多上 create database 相关语句：
CREATE DATABASE testdb ...
USE `testdb`;
--skip-add-drop-table：导出的时候不带上  DROP TABLE IF EXISTS table_name;
 提示：默认是--add-drop-table (Add a DROP TABLE before each create)
这个一般建议带上这个选项， 不然很容易由于dump 没有用好，导致drop了正确的 table 。
*/

// getDumpCmd TODO
/*
mysqldump --skip-add-drop-table -d testdb > testdb.sql

DumpSchema 功能概述：
1. 一个 DB 一个 schema 文件
2. 文件名 DumpDir/$dump_file.$old_db_name.$SUBJOB_ID
3. $mysqldump_file
-h$SOURCE_IP
-P $SOURCE_PORT
-u$dbbackup_user
-p$dbbackup_pass $dump_schema_opt
--skip-foreign-key-check
--skip-opt
--create-option
--single-transaction
-q
--no-autocommit
--default-character-set=$charset_server
-R $create_db_opt $old_db_name
>/data/dbbak/$dump_file.$old_db_name 2>/data/dbbak/$dump_file.$old_db_name.$SUBJOB_ID.err;
*/
// nolint
func (m *MySQLDumper) getDumpCmd(outputFile, errFile, dumpOption string, appendOutput bool) (dumpCmd string) {

	switch {
	case m.DumpData && m.DumpSchema:
		dumpOption += " --hex-blob --create-options "
		if m.ExtendedInsert {
			dumpOption += " --extended-insert "
		}
		// no options represents backup library table data
	case m.DumpData:
		dumpOption += " --no-create-info --no-create-db --hex-blob "
		if m.ExtendedInsert {
			dumpOption += " --extended-insert "
		}
	case m.DumpSchema:
		dumpOption += " -d "
	case !m.DumpData && !m.DumpSchema:
		dumpOption += " --no-create-info --no-data --no-create-db "
	}

	if m.AddDropTable {
		dumpOption += " --add-drop-table "
	} else {
		dumpOption += "--skip-add-drop-table"
	}
	if m.NoCreateDb {
		dumpOption += " -n "
	}
	if m.NoCreateTb {
		dumpOption += " -t "
	}
	if m.DumpRoutine {
		dumpOption += " -R "
	}
	if m.DumpTrigger {
		dumpOption += " --triggers "
	} else {
		dumpOption += " --skip-triggers "
	}
	if m.DumpEvent {
		dumpOption += " --events"
	}
	if m.GtidPurgedOff {
		dumpOption += " --set-gtid-purged=OFF"
	}
	if m.Quick {
		dumpOption += " --quick "
	}
	if m.Charset != "" { // charset 可能为空
		dumpOption += " --default-character-set=" + m.Charset
	}
	if m.Force {
		dumpOption += " -f "
	}
	dumpCmd = fmt.Sprintf(
		`%s -h%s -P%d  -u%s  -p%s --skip-opt --create-options --single-transaction --max-allowed-packet=1G -q --no-autocommit %s`,
		m.DumpCmdFile,
		m.Ip,
		m.Port,
		m.DbBackupUser,
		m.DbBackupPwd,
		dumpOption,
	)

	if cmutil.IsNotEmpty(m.Where) {
		dumpCmd += ` --where='` + m.Where + `'`
	}

	if m.NoUseDbAndWirteCreateDb {
		dumpCmd += fmt.Sprintf(" %s", strings.Join(m.DbNames, " "))
	} else {
		if len(m.DbNames) > 0 {
			dumpCmd += fmt.Sprintf(" --databases  %s", strings.Join(m.DbNames, " "))
		}
	}

	if len(m.Tables) > 0 {
		dumpCmd += fmt.Sprintf(" --tables  %s", strings.Join(m.Tables, " "))
	}

	if len(m.IgnoreTables) > 0 {
		for _, igTb := range m.IgnoreTables {
			dumpCmd += fmt.Sprintf(" --ignore-table=%s", igTb)
		}
	}
	mysqlDumpCmd := fmt.Sprintf("%s > %s 2>%s", dumpCmd, outputFile, errFile)
	if appendOutput {
		mysqlDumpCmd = fmt.Sprintf("%s >> %s 2>>%s", dumpCmd, outputFile, errFile)
	}
	return strings.ReplaceAll(mysqlDumpCmd, "\n", " ")
}

// getTMySQLDumpOption  自研mysqldump
//
//	@receiver m
//	@return dumpCmd
func (m *MySQLDumper) getTMySQLDumpOption() (dumpOption string) {
	return fmt.Sprintf(
		`
	--ignore-show-create-table-error
	--skip-foreign-key-check
	--flush-wait-timeout=0
	--max-concurrency=%d 
	--max-resource-use-percent=%d
	`, m.maxConcurrency, m.maxResourceUsePercent,
	)
}

func (m *MySQLDumper) getTDBCTLDumpOption() (dumpOption string) {
	// 默认false 即不带有SET tc_admin=0
	// 如果不需要下发spider，可添加此参数
	return " --print-tc-admin-info "
}

// MyDumper Options mydumper options
type MyDumper struct {
	Options MyDumperOptions
	Host    string
	Port    int
	User    string
	Pwd     string
	Charset string
	DumpDir string // 备份到哪个目录
	BinPath string // mydumper 的绝对路径
}

// MyDumperOptions mydumper options
type MyDumperOptions struct {
	NoData    bool
	Threads   int
	UseStream bool
	Regex     string
	Db        string
}

// buildCommand build command
func (m *MyDumper) buildCommand() (command string) {
	command = fmt.Sprintf(`%s -h %s -P %d -u %s -p '%s' --set-names=%s`, m.BinPath, m.Host,
		m.Port, m.User, m.Pwd, m.Charset)
	if m.Options.UseStream {
		command += " --stream "
	} else {
		command += fmt.Sprintf(" -o %s ", m.DumpDir)
	}
	command += " --events --routines --triggers --verbose=2 "
	command += " --trx-consistency-only --long-query-retry-interval=10 "
	if m.Options.NoData {
		command += " --no-data "
	}
	if m.Options.Threads > 0 {
		command += fmt.Sprintf(" --threads=%d ", m.Options.Threads)
	}
	if lo.IsNotEmpty(m.Options.Regex) {
		command += fmt.Sprintf(` -x '%s'`, m.Options.Regex)
	}
	if lo.IsNotEmpty(m.Options.Db) {
		command += fmt.Sprintf(" --database=%s ", m.Options.Db)
	}
	// logger.Info("mydumper command: %s", command)
	return
}

// MyLoader Options myloader options
type MyLoader struct {
	Options     MyLoaderOptions
	Host        string
	Port        int
	User        string
	Pwd         string
	Charset     string
	BinPath     string
	LoadDataDir string
}

// MyLoaderOptions TODO
type MyLoaderOptions struct {
	NoData         bool
	UseStream      bool
	Threads        int
	DefaultsFile   string
	SourceDb       string
	TargetDb       string
	OverWriteTable bool // Drop tables if they already exist
}

func (m *MyLoader) buildCommand() (command string) {
	command = fmt.Sprintf(`%s -h %s -P %d -u %s -p '%s' --set-names=%s `, m.BinPath, m.Host,
		m.Port, m.User, m.Pwd, m.Charset)
	command += " --enable-binlog --verbose=2 "
	if lo.IsNotEmpty(m.Options.SourceDb) {
		command += fmt.Sprintf(" -s %s ", m.Options.SourceDb)
	}
	if lo.IsNotEmpty(m.Options.TargetDb) {
		command += fmt.Sprintf(" -B %s ", m.Options.TargetDb)
	}
	if m.Options.UseStream {
		command += " --stream "
	} else {
		command += fmt.Sprintf(" -d %s ", m.LoadDataDir)
	}
	if m.Options.Threads > 0 {
		command += fmt.Sprintf(" --threads=%d ", m.Options.Threads)
	}
	if cmutil.IsNotEmpty(m.Options.DefaultsFile) {
		command += fmt.Sprintf(" --defaults-file=%s ", m.Options.DefaultsFile)
	}
	if m.Options.NoData {
		command += " --no-data "
	}
	if m.Options.OverWriteTable {
		command += " -o "
	}
	return
}

// Loader do myloader load data
func (m *MyLoader) Loader() (err error) {
	m.BinPath = filepath.Join(cst.DbbackupGoInstallPath, "bin/myloader")
	if err = setEnv(); err != nil {
		logger.Error("set env failed %s", err.Error())
		return
	}
	var stderr string
	loadcmd := m.buildCommand()
	stderr, err = osutil.StandardShellCommand(false, loadcmd)
	if err != nil {
		logger.Error("stderr %s", stderr)
		return fmt.Errorf("stderr:%s,err:%w", stderr, err)
	}
	return nil
}

// Dumper do mydumper dump data
func (m *MyDumper) Dumper() (err error) {
	m.BinPath = filepath.Join(cst.DbbackupGoInstallPath, "bin/mydumper")
	if err = setEnv(); err != nil {
		logger.Error("set env failed %s", err.Error())
		return
	}
	var stderr string
	stderr, err = osutil.StandardShellCommand(false, m.buildCommand())
	if err != nil {
		logger.Error("stderr %s", stderr)
		return fmt.Errorf("stderr:%s,err:%w", stderr, err)
	}
	return nil
}

// DumperByEachDb do mydumper dump data by each db
func (m *MyDumper) DumperByEachDb() (err error) {
	m.BinPath = filepath.Join(cst.DbbackupGoInstallPath, "bin/mydumper")
	if err = setEnv(); err != nil {
		logger.Error("set env failed %s", err.Error())
		return
	}
	var stderr string
	stderr, err = osutil.StandardShellCommand(false, m.buildCommand())
	if err != nil {
		logger.Error("stderr %s", stderr)
		return fmt.Errorf("stderr:%s,err:%w", stderr, err)
	}
	return nil
}

// MyStreamDumpLoad  stream dumper loader
type MyStreamDumpLoad struct {
	Dumper *MyDumper
	Loader *MyLoader
}

func (s *MyStreamDumpLoad) buildCommand() (command string) {
	s.Dumper.Options.UseStream = true
	dumpCmd := s.Dumper.buildCommand()
	loadCmd := s.Loader.buildCommand()
	return fmt.Sprintf("%s|%s", dumpCmd, loadCmd)
}

// setEnv mydumper or myloader lib path
func setEnv() (err error) {
	var libPath []string
	libPath = append(libPath, filepath.Join(cst.DbbackupGoInstallPath, "lib/libmydumper"))
	oldLibs := strings.Split(os.Getenv("LD_LIBRARY_PATH"), ":")
	oldLibs = append(oldLibs, libPath...)
	return os.Setenv("LD_LIBRARY_PATH", strings.Join(oldLibs, ":"))
}

// Run Command Run
func (s *MyStreamDumpLoad) Run() (err error) {
	if err = setEnv(); err != nil {
		logger.Error("set env failed %s", err.Error())
		return
	}
	s.Dumper.BinPath = filepath.Join(cst.DbbackupGoInstallPath, "bin/mydumper")
	s.Loader.BinPath = filepath.Join(cst.DbbackupGoInstallPath, "bin/myloader")
	var stderr string
	command := s.buildCommand()
	logger.Info("the stream dump load command is %s", command)
	stderr, err = osutil.StandardShellCommand(false, command)
	if err != nil {
		logger.Error("stderr %s", stderr)
		return fmt.Errorf("stderr:%s,err:%w", stderr, err)
	}
	return nil
}

// OADumper TODO
type OADumper interface {
	OpenAreaDump() error
}

// OpenAreaDumperTogether TODO
type OpenAreaDumperTogether struct {
	OpenAreaDumper
	OutputfileName string
	UseTMySQLDump  bool // 是否使用的是自研的mysqldump,一般介质在备份目录下
}

// OpenAreaDumpOption TODO
type OpenAreaDumpOption struct {
	/* 	DumpSchema   bool
	   	DumpData     bool */
	NoData        bool
	AddDropTable  bool // 默认 false 代表添加 --skip-add-drop-table 选项
	NeedUseDb     bool
	NoCreateDb    bool
	NoCreateTb    bool
	DumpRoutine   bool // 默认 false 代表添加不导出存储过程,True导出存储过程
	DumpTrigger   bool // 默认 false 代表添加不导出触发器
	DumpEvent     bool // 默认 false 导出 event
	GtidPurgedOff bool // --set-gtid-purged=OFF
}

// OpenAreaDumper TODO
type OpenAreaDumper struct {
	OpenAreaDumpOption
	DumpDir      string // 备份到哪个目录
	DbBackupUser string
	DbBackupPwd  string
	Ip           string
	Port         int
	Charset      string
	DumpCmdFile  string // mysqldump 的绝对路径
	DbNames      []string
	IsMaster     bool
	// Todo
	// SelfDefineArgs []string  ...
	// Precheck ...
	runtimectx
}

func (m *OpenAreaDumper) init() {
	m.maxConcurrency = runtime.NumCPU() / 2
	m.maxResourceUsePercent = 50
	if m.IsMaster || m.maxConcurrency == 0 {
		// 如果是在Master Dump的话不允许开启并发
		m.maxConcurrency = 1
	}
}

func (m *OpenAreaDumper) getTMySQLDumpOption() (dumpOption string) {
	return fmt.Sprintf(
		`
	--ignore-show-create-table-error
	--skip-foreign-key-check
	--max-concurrency=%d 
	--max-resource-use-percent=%d
	`, m.maxConcurrency, m.maxResourceUsePercent,
	)
}

// OpenAreaDump TODO
func (m *OpenAreaDumperTogether) OpenAreaDump() (err error) {
	m.init()
	outputFile := path.Join(m.DumpDir, m.OutputfileName)
	errFile := path.Join(m.DumpDir, m.OutputfileName+".err")
	dumpOption := ""
	if m.UseTMySQLDump {
		dumpOption = m.getTMySQLDumpOption()
	}
	dumpCmd := m.getOpenAreaDumpCmd(strings.Join(m.DbNames, " "), outputFile, errFile, dumpOption)
	logger.Info("mysqldump cmd:%s", mysqlcomm.ClearSensitiveInformation(dumpCmd))
	output, err := osutil.StandardShellCommand(false, dumpCmd)
	if err != nil {
		return fmt.Errorf("execte %s get an error:%s,%w", dumpCmd, output, err)
	}
	if err := checkDumpComplete(outputFile); err != nil {
		logger.Error("checkDumpComplete failed %s", err.Error())
		return err
	}
	return
}

func (m *OpenAreaDumper) getOpenAreaDumpCmd(dbName, outputFile, errFile, dumpOption string) (dumpCmd string) {
	if m.NoData {
		dumpOption += " -d "
	}
	if m.AddDropTable {
		dumpOption += " --add-drop-table "
	} else {
		dumpOption += "--skip-add-drop-table"
	}
	if m.NeedUseDb {
		dumpOption += " -B "
	}
	if m.NoCreateDb {
		dumpOption += " -n "
	}
	if m.NoCreateTb {
		dumpOption += " -t "
	}
	if m.DumpRoutine {
		dumpOption += " -R "
	}
	if m.DumpTrigger {
		dumpOption += " --triggers "
	} else {
		dumpOption += " --skip-triggers "
	}
	if m.DumpEvent {
		dumpOption += " --events"
	}
	if m.GtidPurgedOff {
		dumpOption += " --set-gtid-purged=OFF"
	}
	dumpCmd = fmt.Sprintf(
		//nolint
		`%s -h%s -P%d -u%s -p%s --skip-opt --create-options --single-transaction --max-allowed-packet=1G -q --no-autocommit --default-character-set=%s %s %s > %s 2>%s`,
		m.DumpCmdFile,
		m.Ip,
		m.Port,
		m.DbBackupUser,
		m.DbBackupPwd,
		m.Charset,
		dumpOption,
		dbName,
		outputFile,
		errFile,
	)
	return strings.ReplaceAll(dumpCmd, "\n", " ")
}

// DbbackupDumper 使用备份工具进行数据导出
type DbbackupDumper interface {
	DumpbackupLogical() error
	// PhysicalBackup() error
}

// DbMigrateDumper TODO
type DbMigrateDumper struct {
	DumpDir         string
	DbBackupUser    string
	DbBackupPwd     string
	Ip              string
	Port            int
	BackupCmdPath   string
	DbNames         string
	DataSchemaGrant string
	// LogDir mydumper 执行日志所放目录。为了解决目录权限问题
	LogDir string `json:"-"`
}

// DumpbackupLogical 使用备份工具进行逻辑备份
func (d *DbMigrateDumper) DumpbackupLogical() (err error) {
	backupCmd := d.getBackupCmd()
	logger.Info("backupcmd is %s", mysqlcomm.ClearSensitiveInformation(backupCmd))
	output, err := osutil.StandardShellCommand(false, backupCmd)
	if err != nil {
		return fmt.Errorf("execte %s get an error:%s,%w", backupCmd, output, err)
	}
	return nil
}

// getBackupCmd 拼接数据备份命令
func (d *DbMigrateDumper) getBackupCmd() (backupCmd string) {
	backupCmd = fmt.Sprintf("%s dumpbackup logical -u%s -p%s -h%s -P%d %s",
		d.BackupCmdPath, d.DbBackupUser, d.DbBackupPwd, d.Ip, d.Port, d.getBackupCmdOption())
	if d.LogDir != "" {
		backupCmd += fmt.Sprintf(" --log-dir %s", d.LogDir)
	}
	return backupCmd
}

// getBackupCmdOption 拼接数据备份参数选项
func (d *DbMigrateDumper) getBackupCmdOption() (opt string) {
	opt = fmt.Sprintf(" --nocheck-diskspace --data-schema-grant=%s "+
		"--backup-dir=%s --cluster-domain=xx.xx --databases %s --tables \\*",
		d.DataSchemaGrant, d.DumpDir, d.DbNames)
	return opt
}
