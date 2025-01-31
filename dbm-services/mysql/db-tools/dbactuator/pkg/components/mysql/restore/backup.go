package restore

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"dbm-services/common/go-pubpkg/cmutil"
	"dbm-services/common/go-pubpkg/logger"
	"dbm-services/mysql/db-tools/dbactuator/pkg/components/mysql/dbbackup"
)

// BackupInfo backup file info
type BackupInfo struct {
	WorkDir string `json:"work_dir" validate:"required" example:"/data1/dbbak"` // 备份恢复目录，工作目录
	// 备份文件所在本地目录，理论上doDr不会对该目录写入，而是写入 targetDir
	BackupDir string `json:"backup_dir"  validate:"required" example:"/data/dbbak"`
	// 备份文件名列表，key 是 info|full|priv|index, value 是是相对于 backup_dir 的文件名列表
	BackupFiles map[string][]string `json:"backup_files" validate:"required"`

	backupType string
	backupHost string
	backupPort int

	infoFilePath string // InfoFileDetail full path filename
	infoObj      *dbbackup.InfoFileDetail

	indexFilePath string
	indexObj      *dbbackup.BackupIndexFile
}

func (b *BackupInfo) initWorkDirs() error {
	/*
		backup_file_list       a, b, c
		backup_dir /data/dbbak/
		backup_untar_dir ${task_dir}/base_name/

		work_base_dir /data1/dbbak
		task_dir   ${work_base_dir}/doDr_10941094/20000/
	*/
	if !cmutil.IsDirectory(b.WorkDir) {
		return errors.Errorf("error work_dir %s", b.WorkDir)
	}
	return nil
}

// CheckIntegrity TODO
func (b *BackupInfo) CheckIntegrity() error {
	return nil
}

// GetBackupMetaFile godoc
// 获取 .info / .index 文件名，解析文件内容
func (b *BackupInfo) GetBackupMetaFile(fileType string) error {
	fileList, ok := b.BackupFiles[fileType]
	if !ok {
		return errors.Errorf("backup_files has no file_type: %s", fileType)
	}
	if len(fileList) != 1 {
		return fmt.Errorf("expect one meta file but got %v", fileList)
	}
	metaFilename := strings.TrimSpace(fileList[0])
	metaFilePath := filepath.Join(b.BackupDir, metaFilename)
	logger.Info("GetBackupMetaFile metaFilePath:%+v", metaFilePath)
	if err := cmutil.FileExistsErr(metaFilePath); err != nil {
		return err
	}
	if strings.HasSuffix(metaFilename, ".index") {
		b.indexFilePath = metaFilePath
		//b.backupBaseName = strings.TrimSuffix(metaFilename, ".index")
		var indexObj = &dbbackup.BackupIndexFile{}
		if err := dbbackup.ParseBackupIndexFile(b.indexFilePath, indexObj); err != nil {
			return err
		} else {
			b.indexObj = indexObj
			b.backupType = b.indexObj.BackupType
			b.backupHost = b.indexObj.BackupHost
			b.backupPort = b.indexObj.BackupPort
		}
	}

	if b.indexObj != nil && cmutil.StringsHas([]string{"gztab", "xtra"}, b.indexObj.BackupTool) {
		indexInfoFile := ""
		for _, fileItem := range b.indexObj.FileList {
			if fileItem.FileType == "index" {
				indexInfoFile = fileItem.FileName
				break
			}
		}
		metaFilename = indexInfoFile
		metaFilePath = filepath.Join(b.BackupDir, metaFilename)
	}
	logger.Info("GetBackupMetaFile backupType:%s metaFilename %s", b.backupType, metaFilename)
	if strings.HasSuffix(metaFilename, ".info") {
		b.infoFilePath = metaFilePath
		//b.backupBaseName = strings.TrimSuffix(metaFilename, ".info")
		var infoObj = &dbbackup.InfoFileDetail{}
		if err := dbbackup.ParseBackupInfoFile(b.infoFilePath, infoObj); err != nil {
			return err
		} else {
			b.infoObj = infoObj
			b.backupType = b.infoObj.BackupType
			b.backupHost = b.infoObj.BackupHost
			b.backupPort = b.infoObj.BackupPort
			logger.Info("GetBackupMetaFile infoObj:%+v", b.infoObj)
			b.indexObj = &dbbackup.BackupIndexFile{}
			b.indexObj.BackupTool = b.infoObj.BackupType
			if infoObj.BackupType == "gztab" {
				b.indexObj.BackupType = "logical"
			} else if infoObj.BackupType == "xtra" {
				b.indexObj.BackupType = "physical"
			}
		}
	}
	logger.Info("backupType=%s, backupHost=%s, backupPort=%d", b.backupType, b.backupHost, b.backupPort)
	return nil
}

func newTimestampString() string {
	return time.Now().Format("20060102150405")
}
