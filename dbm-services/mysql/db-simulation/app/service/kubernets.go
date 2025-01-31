/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-DB管理系统(BlueKing-BK-DBM) available.
 * Copyright (C) 2017-2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at https://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package service

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"

	"dbm-services/common/go-pubpkg/cmutil"
	"dbm-services/common/go-pubpkg/logger"
	"dbm-services/mysql/db-simulation/app"
	"dbm-services/mysql/db-simulation/app/config"
	"dbm-services/mysql/db-simulation/model"
)

// Kcs k8s client sets
var Kcs KubeClientSets

// DefaultUser default user
const DefaultUser = "root"

// KubeClientSets k8s client sets
type KubeClientSets struct {
	Cli        *kubernetes.Clientset
	RestConfig *rest.Config
	Namespace  string // namespace
}

// MySQLPodBaseInfo mysql pod base info
type MySQLPodBaseInfo struct {
	PodName string
	Lables  map[string]string
	Args    []string
	RootPwd string
	Charset string
}

// DbPodSets db pod sets
type DbPodSets struct {
	K8S         KubeClientSets
	BaseInfo    *MySQLPodBaseInfo
	DbWork      *cmutil.DbWorker
	DbImage     string
	TdbCtlImage string
	SpiderImage string
}

// ClusterPodSets cluster pod sets
type ClusterPodSets struct {
	DbPodSets
}

var startArgsSplitRe *regexp.Regexp

func init() {
	logger.Info("start init bcs client ")
	Kcs.RestConfig = &rest.Config{
		Host:        config.GAppConfig.Bcs.EndpointUrl + "/clusters/" + config.GAppConfig.Bcs.ClusterId + "/",
		BearerToken: config.GAppConfig.Bcs.Token,
		ContentConfig: rest.ContentConfig{
			ContentType:  "application/json",
			GroupVersion: &v1.SchemeGroupVersion,
		},
		Timeout: 10 * time.Second,
	}
	clientSet, err := kubernetes.NewForConfig(Kcs.RestConfig)
	if err != nil {
		logger.Fatal("init kubernets client failed %s", err.Error())
		return
	}
	Kcs.Cli = clientSet
	Kcs.Namespace = config.GAppConfig.Bcs.NameSpace
	startArgsSplitRe = regexp.MustCompile(` |,`)
}

// NewDbPodSets new db pod sets
func NewDbPodSets() *DbPodSets {
	return &DbPodSets{
		K8S: Kcs,
	}
}

func (k *DbPodSets) getCreateClusterSqls() []string {
	var ss []string
	ss = append(ss, fmt.Sprintf(
		"tdbctl create node wrapper 'SPIDER' options(user 'root', password '%s', host '127.0.0.1', port 25000);",
		k.BaseInfo.RootPwd))
	ss = append(ss, fmt.Sprintf(
		"tdbctl create node wrapper 'mysql' options(user 'root', password '%s', host '127.0.0.1', port 20000);",
		k.BaseInfo.RootPwd))
	ss = append(ss, fmt.Sprintf(
		"tdbctl create node wrapper 'TDBCTL' options(user 'root', password '%s', host '127.0.0.1', port 26000);",
		k.BaseInfo.RootPwd))
	ss = append(ss, "tdbctl enable primary;")
	ss = append(ss, "tdbctl flush routing;")
	return ss
}

// getClusterPodContanierSpec create cluster pod container spec
// nolint
func (k *DbPodSets) getClusterPodContanierSpec(mysqlVersion string) []v1.Container {
	return []v1.Container{
		{
			Name: "backend",
			Env: []v1.EnvVar{{
				Name:  "MYSQL_ROOT_PASSWORD",
				Value: k.BaseInfo.RootPwd,
			}},
			Resources:       k.getResourceLimit(),
			ImagePullPolicy: v1.PullIfNotPresent,
			Image:           k.DbImage,
			Args:            k.getbackendStartArgs(mysqlVersion),
			ReadinessProbe: &v1.Probe{
				ProbeHandler: v1.ProbeHandler{
					Exec: &v1.ExecAction{
						Command: []string{"/bin/bash", "-c", fmt.Sprintf("mysql -uroot -p%s -e 'select 1'", k.BaseInfo.RootPwd)},
					},
				},
				InitialDelaySeconds: 3,
				PeriodSeconds:       5,
			},
		}, {
			Name: "spider",
			Env: []v1.EnvVar{{
				Name:  "MYSQL_ROOT_PASSWORD",
				Value: k.BaseInfo.RootPwd,
			}},
			Resources:       k.getResourceLimit(),
			ImagePullPolicy: v1.PullIfNotPresent,
			Image:           k.SpiderImage,
			Args:            k.getSpiderStartArgs(),
			ReadinessProbe: &v1.Probe{
				ProbeHandler: v1.ProbeHandler{
					Exec: &v1.ExecAction{
						Command: []string{"/bin/bash", "-c", fmt.Sprintf("mysql -uroot -p%s -e 'select 1'", k.BaseInfo.RootPwd)},
					},
				},
				InitialDelaySeconds: 3,
				PeriodSeconds:       5,
			},
		},
		{
			Name: "tdbctl",
			Env: []v1.EnvVar{{
				Name:  "MYSQL_ROOT_PASSWORD",
				Value: k.BaseInfo.RootPwd,
			}},
			Resources:       k.gettdbctlResourceLimit(),
			ImagePullPolicy: v1.PullIfNotPresent,
			Image:           k.TdbCtlImage,
			Args:            k.getTdbctlStartArgs(),
			ReadinessProbe: &v1.Probe{
				ProbeHandler: v1.ProbeHandler{
					Exec: &v1.ExecAction{
						Command: []string{"/bin/bash", "-c", fmt.Sprintf("mysql -uroot -p%s -e 'select 1'", k.BaseInfo.RootPwd)},
					},
				},
				InitialDelaySeconds: 3,
				PeriodSeconds:       5,
			},
		},
	}
}

func (k *DbPodSets) getTdbctlStartArgs() (args []string) {
	args = []string{"mysqld",
		"--defaults-file=/etc/my.cnf",
		"--port=26000",
		"--tc-admin=1",
		"--dbm-allow-standalone-primary",
		"--max_allowed_packet=1073741824",
		fmt.Sprintf("--character-set-server=%s",
			k.BaseInfo.Charset),
		"--user=mysql"}
	dbArgs, err := model.GetStartArsg("tdbctl", LatestVersion)
	if err != nil {
		logger.Warn("get tdbctl start args failed %s", err.Error())
		return
	}
	for _, arg := range startArgsSplitRe.Split(dbArgs, -1) {
		if lo.IsNotEmpty(arg) {
			args = append(args, strings.TrimSpace(arg))
		}
	}
	return
}

func (k *DbPodSets) getSpiderStartArgs() (args []string) {
	args = []string{"mysqld",
		"--defaults-file=/etc/my.cnf",
		"--log_bin_trust_function_creators",
		"--port=25000",
		"--max_allowed_packet=1073741824",
		fmt.Sprintf("--character-set-server=%s",
			k.BaseInfo.Charset),
		"--user=mysql"}
	dbArgs, err := model.GetStartArsg("spider", LatestVersion)
	if err != nil {
		logger.Warn("get spider start args failed %s", err.Error())
		return
	}
	for _, arg := range startArgsSplitRe.Split(dbArgs, -1) {
		if lo.IsNotEmpty(arg) {
			args = append(args, strings.TrimSpace(arg))
		}
	}
	return
}

func (k *DbPodSets) getbackendStartArgs(mysqlVersion string) (args []string) {
	args = []string{"mysqld",
		"--defaults-file=/etc/my.cnf",
		"--log_bin_trust_function_creators",
		"--port=20000",
		"--max_allowed_packet=1073741824",
		"--sql-mode=",
		fmt.Sprintf("--character-set-server=%s",
			k.BaseInfo.Charset),
		"--user=mysql"}
	if cmutil.MySQLVersionParse(mysqlVersion) >= cmutil.MySQLVersionParse("8.0.0") {
		args = append(args, "--default-authentication-plugin=mysql_native_password")
	}
	dbArgs, err := model.GetStartArsg("mysql", mysqlVersion)
	if err != nil {
		logger.Warn("get mysql start args failed %s", err.Error())
		return
	}
	for _, arg := range startArgsSplitRe.Split(dbArgs, -1) {
		if lo.IsNotEmpty(arg) {
			args = append(args, strings.TrimSpace(arg))
		}
	}
	return
}

// CreateClusterPod create tendbcluster simulation pod
func (k *DbPodSets) CreateClusterPod(mySQLVersion string) (err error) {
	c := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.BaseInfo.PodName,
			Namespace: k.K8S.Namespace,
			Labels:    k.BaseInfo.Lables,
		},
		Spec: v1.PodSpec{
			NodeSelector: lo.SliceToMap(config.GAppConfig.SimulationNodeLables, func(item config.LabelItem) (k, v string) {
				return item.Key,
					item.Value
			}),
			Tolerations: k.getToleration(),
			Containers:  k.getClusterPodContanierSpec(mySQLVersion),
		},
	}
	if err = k.createpod(c, 26000); err != nil {
		logger.Error("create spider cluster failed %s", err.Error())
		return err
	}
	logger.Info("connect tdbctl success ~")
	// create cluster relation
	for _, ql := range k.getCreateClusterSqls() {
		logger.Info("exec init cluster sql %s", ql)
		if _, err = k.DbWork.Db.Exec(ql); err != nil {
			return err
		}
	}
	return nil
}

// createpod create pod
func (k *DbPodSets) createpod(pod *v1.Pod, probePort int) (err error) {
	podc, err := k.K8S.Cli.CoreV1().Pods(k.K8S.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		logger.Error("create pod failed %s", err.Error())
		return err
	}
	uid := string(podc.GetUID())
	model.DB.Create(&model.TbContainerRecord{
		Container:     k.BaseInfo.PodName,
		Uid:           uid,
		CreatePodTime: time.Now(),
		CreateTime:    time.Now()})
	podIp := podc.Status.PodIP
	// 连续多次探测pod的状态
	fn := func() (err error) {
		var podI *v1.Pod
		podI, err = k.K8S.Cli.CoreV1().Pods(k.K8S.Namespace).Get(context.TODO(), k.BaseInfo.PodName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if len(podI.Status.ContainerStatuses) == 0 {
			return fmt.Errorf("get pod status is empty,wait some seconds")
		}
		for _, cStatus := range podI.Status.ContainerStatuses {
			logger.Info("%s: %v", cStatus.Name, cStatus.Ready)
			if !cStatus.Ready {
				return fmt.Errorf("container %s is not ready", cStatus.Name)
			}
			for _, podCondition := range podI.Status.Conditions {
				if podI.Status.Phase != v1.PodRunning {
					logger.Warn("%s: %v", podCondition.Status, podCondition.Message, podCondition.Reason)
				}
			}
		}
		podIp = podI.Status.PodIP
		logger.Info("the pod is ready,ip is %s", podIp)
		return nil
	}
	if err = cmutil.Retry(cmutil.RetryConfig{Times: 120, DelayTime: 2 * time.Second}, fn); err != nil {
		return err
	}
	logger.Info("the podIp is %s", podIp)
	fnc := func() error {
		k.DbWork, err = cmutil.NewDbWorker(fmt.Sprintf("%s:%s@tcp(%s:%d)/?timeout=5s&multiStatements=true",
			DefaultUser,
			k.BaseInfo.RootPwd,
			podIp, probePort))
		if err != nil {
			logger.Error("connect to pod %s failed %s", podIp, err.Error())
			return errors.Wrap(err, "create pod success,connect to mysql pod failed")
		}
		return nil
	}
	if err = cmutil.Retry(cmutil.RetryConfig{Times: 60, DelayTime: 1 * time.Second}, fnc); err == nil {
		model.UpdateTbContainerRecord(k.BaseInfo.PodName)
	}
	_, errx := k.DbWork.Db.Exec("create user ADMIN@localhost;")
	if errx != nil {
		logger.Error("create user ADMIN@localhost failed %s", errx.Error())
	}
	_, errx = k.DbWork.Db.Exec("grant all on *.* to ADMIN@localhost;")
	if errx != nil {
		logger.Error("grants user failed %s", errx.Error())
	}
	return err
}

// getToleration special  node
func (k *DbPodSets) getToleration() []v1.Toleration {
	ts := []v1.Toleration{}
	for _, item := range config.GAppConfig.SimulationNodeLables {
		ts = append(ts, v1.Toleration{
			Key:      item.Key,
			Operator: v1.TolerationOpExists,
		})
	}
	return ts
}

func (k *DbPodSets) getResourceLimit() v1.ResourceRequirements {
	if !config.IsEmptyMySQLPodResourceConfig() {
		return v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(config.GAppConfig.MySQLPodResource.Limits.Cpu),
				v1.ResourceMemory: resource.MustParse(config.GAppConfig.MySQLPodResource.Limits.Mem),
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(config.GAppConfig.MySQLPodResource.Requests.Cpu),
				v1.ResourceMemory: resource.MustParse(config.GAppConfig.MySQLPodResource.Requests.Mem),
			},
		}
	}
	return v1.ResourceRequirements{}
}

func (k *DbPodSets) gettdbctlResourceLimit() v1.ResourceRequirements {
	if !config.IsEmptyTdbctlPodResourceConfig() {
		return v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(config.GAppConfig.TdbctlPodResource.Limits.Cpu),
				v1.ResourceMemory: resource.MustParse(config.GAppConfig.TdbctlPodResource.Limits.Mem),
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse(config.GAppConfig.TdbctlPodResource.Requests.Cpu),
				v1.ResourceMemory: resource.MustParse(config.GAppConfig.TdbctlPodResource.Requests.Mem),
			},
		}
	}
	return v1.ResourceRequirements{}
}

func (k *DbPodSets) getTendbhaPodStartArgs(mysqlVersion string) (args []string) {
	args = []string{
		"mysqld",
		"--defaults-file=/etc/my.cnf",
		"--skip-log-bin",
		"--max_allowed_packet=1073741824",
		fmt.Sprintf("--character-set-server=%s", k.BaseInfo.Charset)}
	if cmutil.MySQLVersionParse(mysqlVersion) >= cmutil.MySQLVersionParse("8.0.0") {
		args = append(args, "--default-authentication-plugin=mysql_native_password")
	}
	dbArgs, err := model.GetStartArsg("mysql", mysqlVersion)
	if err != nil {
		logger.Warn("get mysql start args failed %s", err.Error())
		return
	}
	for _, arg := range startArgsSplitRe.Split(dbArgs, -1) {
		if lo.IsNotEmpty(arg) {
			args = append(args, strings.TrimSpace(arg))
		}
	}
	return
}

// CreateMySQLPod create mysql pod
func (k *DbPodSets) CreateMySQLPod(mysqlVersion string) (err error) {
	startArgs := k.getTendbhaPodStartArgs(mysqlVersion)
	startArgs = append(startArgs, k.BaseInfo.Args...)
	startArgs = append(startArgs, "--user=mysql")
	logger.Info("start pod args %v", startArgs)
	c := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.BaseInfo.PodName,
			Namespace: k.K8S.Namespace,
			Labels:    k.BaseInfo.Lables,
		},
		Spec: v1.PodSpec{
			NodeSelector: lo.SliceToMap(config.GAppConfig.SimulationNodeLables, func(item config.LabelItem) (k, v string) {
				return item.Key,
					item.Value
			}),
			Tolerations: k.getToleration(),
			Containers: []v1.Container{{
				Resources: k.getResourceLimit(),
				Name:      app.MySQL,
				Env: []v1.EnvVar{{
					Name:  "MYSQL_ROOT_PASSWORD",
					Value: k.BaseInfo.RootPwd,
				}},
				Ports: []v1.ContainerPort{
					{ContainerPort: 3306},
				},
				ImagePullPolicy: v1.PullIfNotPresent,
				Image:           k.DbImage,
				Args:            startArgs,
				ReadinessProbe: &v1.Probe{
					ProbeHandler: v1.ProbeHandler{
						Exec: &v1.ExecAction{
							Command: []string{"/bin/bash", "-c", fmt.Sprintf("mysql -uroot -p%s -e 'select 1'", k.BaseInfo.RootPwd)},
						},
					},
					InitialDelaySeconds: 2,
					PeriodSeconds:       5,
				},
			}},
		},
	}

	return k.createpod(c, 3306)
}

// DeletePod delete pod
func (k *DbPodSets) DeletePod() (err error) {
	return k.K8S.Cli.CoreV1().Pods(k.K8S.Namespace).Delete(context.TODO(), k.BaseInfo.PodName, metav1.DeleteOptions{})
}

// getLoadSchemaSQLCmd create load schema sql cmd
func (k *DbPodSets) getLoadSchemaSQLCmd(bkpath, file string) (cmd string) {
	commands := []string{}
	commands = append(commands, k.getDownloadSqlCmd(bkpath, file))
	// sed -i '/50720 SET tc_admin=0/d'
	// 从中控dump的schema文件,默认是添加了tc_admin=0,需要删除
	// 因为模拟执行是需要将中控进行sql转发
	commands = append(commands, fmt.Sprintf("sed -i '/50720 SET tc_admin=0/d' %s", file))
	commands = append(commands, fmt.Sprintf("mysql -uroot -p%s --default-character-set=%s -vvv < %s", k.BaseInfo.RootPwd,
		k.BaseInfo.Charset, file))
	return strings.Join(commands, " && ")
}

// getLoadSQLCmd get load sql cmd
func (k *DbPodSets) getLoadSQLCmd(bkpath, file string, dbs []string) (cmd []string) {
	cmd = append(cmd, k.getDownloadSqlCmd(bkpath, file))
	for _, db := range dbs {
		cmd = append(cmd, fmt.Sprintf("mysql --defaults-file=/etc/my.cnf -uroot -p%s --default-character-set=%s -vvv %s < %s",
			k.BaseInfo.RootPwd, k.BaseInfo.Charset, db, file))
	}
	return cmd
}

func (k *DbPodSets) getDownloadSqlCmd(bkpath, file string) string {
	downloadcmd := fmt.Sprintf("curl -s -S -o %s %s", file, getdownloadUrl(bkpath, file))
	if cmutil.IsNotEmpty(config.GAppConfig.BkRepo.User) && cmutil.IsNotEmpty(config.GAppConfig.BkRepo.Pwd) {
		downloadcmd = fmt.Sprintf("curl -u %s:%s  -s -S -o %s %s", config.GAppConfig.BkRepo.User,
			config.GAppConfig.BkRepo.Pwd, file, getdownloadUrl(bkpath, file))
	}
	return downloadcmd
}

func getdownloadUrl(bkpath, file string) string {
	endpoint := config.GAppConfig.BkRepo.EndPointUrl
	project := config.GAppConfig.BkRepo.Project
	publicbucket := config.GAppConfig.BkRepo.PublicBucket
	u, err := url.Parse(endpoint)
	if err != nil {
		return ""
	}
	r, err := url.Parse(path.Join("/generic", project, publicbucket, bkpath, file))
	if err != nil {
		logger.Error(err.Error())
		return ""
	}
	ll := u.ResolveReference(r).String()
	logger.Info("download url: %s", ll)
	return ll
}

// executeInPod TODO
func (k *DbPodSets) executeInPod(cmd, container string, extMap map[string]string, noLogger bool) (stdout,
	stderr bytes.Buffer,
	err error) {
	xlogger := logger.New(os.Stdout, true, logger.InfoLevel, extMap)
	logger.Info("start exec...")
	req := k.K8S.Cli.CoreV1().RESTClient().Post().Resource("pods").Name(k.BaseInfo.PodName).Namespace(k.K8S.Namespace).
		SubResource("exec").
		Param("container", container)
	logger.Info(cmd)
	req.VersionedParams(
		&v1.PodExecOptions{
			Command: []string{"/bin/bash", "-c", cmd},
			Stdin:   false,
			Stdout:  true,
			Stderr:  true,
		},
		scheme.ParameterCodec,
	)
	reader, writer := io.Pipe()
	exec, err := remotecommand.NewSPDYExecutor(k.K8S.RestConfig, "POST", req.URL())
	if err != nil {
		logger.Error("at remotecommand.NewSPDYExecutor %s", err.Error())
		return bytes.Buffer{}, bytes.Buffer{}, err
	}
	// 导入表结构的时候不打印普通非关键日志

	go func() {
		buf := []byte{}
		sc := bufio.NewScanner(reader)
		sc.Buffer(buf, 2048*1024)
		lineNumber := 1
		for sc.Scan() {
			if !noLogger {
				// 此方案打印的日志会在前端展示
				xlogger.Info("%s", sc.Text())
			} else {
				logger.Info(sc.Text())
			}
			lineNumber++
		}
		if err = sc.Err(); err != nil {
			logger.Error("something bad happened in the line %v: %v", lineNumber, err)
			return
		}
	}()
	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: writer,
		Stderr: &stderr,
		Tty:    false,
	})
	if err != nil {
		xlogger.Error("exec.Stream failed %s:\n stdout:%s\n stderr: %s", err.Error(), strings.TrimSpace(stdout.String()),
			strings.TrimSpace(stderr.String()))
		return stdout, stderr, err
	}
	xlogger.Info("exec successfully...")
	logger.Info("info stdout:%s\nstderr:%s ", strings.TrimSpace(stdout.String()),
		strings.TrimSpace(stderr.String()))
	return stdout, stderr, nil
}
