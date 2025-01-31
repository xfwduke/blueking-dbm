/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-DB管理系统(BlueKing-BK-DBM) available.
 * Copyright (C) 2017-2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package syntax

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"runtime/debug"
	"slices"
	"sync"

	"github.com/samber/lo"

	"dbm-services/common/go-pubpkg/logger"
)

// AnalyzeConcurrency 解析并发度
const AnalyzeConcurrency = 10

// DoParseRelationDbs parse relation db from sql file
func (tf *TmysqlParseFile) DoParseRelationDbs(version string) (createDbs, relationDbs []string, dumpAll bool,
	err error) {
	logger.Info("doing....")
	tf.result = make(map[string]*CheckInfo)
	tf.tmpWorkdir = tf.BaseWorkdir
	tf.mu = sync.Mutex{}

	if !tf.IsLocalFile {
		if err = tf.Init(); err != nil {
			logger.Error("Do init failed %s", err.Error())
			return nil, nil, false, err
		}
		if err = tf.Downloadfile(); err != nil {
			logger.Error("failed to download sql file from the product library %s", err.Error())
			return nil, nil, false, err
		}
	}
	// 最后删除临时目录,不会返回错误
	defer tf.delTempDir()
	logger.Info("all sqlfiles download ok ~")
	alreadExecutedSqlfileChan := make(chan string, len(tf.Param.FileNames))

	go func() {
		if err = tf.Execute(alreadExecutedSqlfileChan, version); err != nil {
			logger.Error("failed to execute tmysqlparse: %s", err.Error())
		}
		close(alreadExecutedSqlfileChan)
	}()

	logger.Info("start to analyze the parsing result")
	createDbs, relationDbs, dumpAll, err = tf.doParseInchan(alreadExecutedSqlfileChan, version)
	if err != nil {
		logger.Error("failed to analyze the parsing result:%s", err.Error())
		return nil, nil, false, err
	}
	logger.Info("createDbs:%v,relationDbs:%v,dumpAll:%v,err:%v", createDbs, relationDbs, dumpAll, err)
	dumpdbs := []string{}
	for _, d := range relationDbs {
		if slices.Contains(createDbs, d) {
			continue
		}
		dumpdbs = append(dumpdbs, d)
	}
	return lo.Uniq(createDbs), lo.Uniq(dumpdbs), dumpAll, nil
}

// doParseInchan RelationDbs do parse relation db
func (t *TmysqlParse) doParseInchan(alreadExecutedSqlfileCh chan string,
	mysqlVersion string) (createDbs []string, relationDbs []string, dumpAll bool, err error) {
	var errs []error
	c := make(chan struct{}, AnalyzeConcurrency)
	errChan := make(chan error)
	wg := &sync.WaitGroup{}
	//  stopchan len 必须要和 AnalyzeConcurrency 保持一致
	stopChan := make(chan struct{}, AnalyzeConcurrency)

	for sqlfile := range alreadExecutedSqlfileCh {
		wg.Add(1)
		c <- struct{}{}
		go func(fileName string) {
			defer wg.Done()
			cdbs, dbs, dumpAllDbs, err := t.analyzeRelationDbs(fileName, mysqlVersion)
			logger.Info("createDbs:%v,dbs:%v,dumpAllDbs:%v,err:%v", cdbs, dbs, dumpAllDbs, err)
			if err != nil {
				errChan <- err
			}
			// 如果有dumpall 则直接返回退出,不在继续分析
			if dumpAllDbs {
				dumpAll = true
				<-c
				wg.Done()
				stopChan <- struct{}{}
			}
			t.mu.Lock()
			relationDbs = append(relationDbs, dbs...)
			createDbs = append(createDbs, cdbs...)
			t.mu.Unlock()
			<-c
		}(sqlfile)
	}

	go func() {
		wg.Wait()
		close(errChan)
		stopChan <- struct{}{}
	}()

	for {
		select {
		case err := <-errChan:
			errs = append(errs, err)
		case <-stopChan:
			return createDbs, relationDbs, dumpAll, errors.Join(errs...)
		}
	}
}

// analyzeRelationDbs 分析变更sql相关的db
func (t *TmysqlParse) analyzeRelationDbs(inputfileName, mysqlVersion string) (
	createDbs []string,
	relationDbs []string,
	dumpAll bool,
	err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic error:%v,stack:%s", r, string(debug.Stack()))
		}
	}()
	f, err := os.Open(t.getAbsoutputfilePath(inputfileName, mysqlVersion))
	if err != nil {
		logger.Error("open file failed %s", err.Error())
		return nil, nil, false, err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	for {
		line, errx := reader.ReadBytes(byte('\n'))
		if errx != nil {
			if errx == io.EOF {
				break
			}
			logger.Error("read Line Error %s", errx.Error())
			return nil, nil, false, errx
		}
		if len(line) == 1 && line[0] == byte('\n') {
			continue
		}
		var res ParseLineQueryBase
		if err = json.Unmarshal(line, &res); err != nil {
			logger.Error("json unmasrshal line:%s failed %s", string(line), err.Error())
			return nil, nil, false, err
		}
		// 判断是否有语法错误
		if res.ErrorCode != 0 {
			return nil, nil, false, err
		}
		if slices.Contains([]string{SQLTypeCreateProcedure, SQLTypeCreateFunction, SQLTypeCreateView, SQLTypeCreateTrigger,
			SQLTypeInsertSelect, SQLTypeRelaceSelect},
			res.Command) {
			return nil, nil, true, nil
		}
		if lo.IsEmpty(res.DbName) {
			continue
		}
		// create db not need dump db
		if slices.Contains([]string{SQLTypeCreateDb}, res.Command) {
			createDbs = append(createDbs, res.DbName)
			continue
		}
		relationDbs = append(relationDbs, res.DbName)

	}
	return createDbs, relationDbs, false, nil
}
