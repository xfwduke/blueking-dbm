[[toc]]

`dbm-services/mysql/db-priv/service/add_priv.go:62`

`func (m *PrivTaskPara) AddPriv(jsonPara string, ticket string)`

# `AddPrivDryRun`

调用了 `AddPrivDryRun`, 这其实有点类似一个 _validator_ , 但又还会干点其他的事情
  * 删除重复的源 _IP_ 和目标实例
  * 用参数中的 _BkBizId, ClusterType, User, Dbname_ 到元数据 _DB_ 中再查询一次规则详情. 但是这个详情又没有返回, 看起来还是在做输入验证

# 迭代 `AccountRules`

这实际是在迭代输入的 _dbname list_

* 代码中的 _Account_ 拼错了
* `AccountRules` 是 `[]TbAccountsRules` 类型, 这个对应权限申请单据的规则明细
  * 参数传入的规则明细不完整, 用 `GetAccountRuleInfo` 重新取一遍

## 迭代 `TargetInstances`

`TargetInstances` 是 `[]string` 类型, 应该是权限申请单中的目标实例或者域名

* 调用 `GetCluster` 获取对应集群的详情, 这个函数对应了 _db_meta/priv_api_

然后按不同集群类型有不同的分支

### _TenDBHA, TenDBSingle_
#### _Proxy_ 白名单
1. 调用 `GenerateProxyPrivilege`
2. 调用 `ImportProxyPrivilege`

#### _Backend_ 权限
调用 `ImportBackendPrivilege`
* 客户端 `sourceIps` 类型是 `[]string`
* `GetMySQLVersion` 获取目标实例版本, 因为授权语句有版本差异
* 调用 `GenerateBackendSQL` 生成在 _backend_ 执行的语句, 这里的客户端 _ip_ 按需替换成了目标实例的 _proxy ip_
  * _ips_ 参数是客户端 _ip_ 
  * 迭代 _ips_
    * 调用 `GetPassword` 
      1. 通过 _DRS_ 获取 _user@ip_ 在目标实例的真实密码
      2. 判断密码算法版本
      3. 还判断了实际密码和账号规则密码是否一致
    * 调用 `CheckDbCross`
      1. 通过 _DRS_ 获取 _user@ip_ 在库模式详情
      2. 迭代返回的每一个库模式检查是否存在库模式冲突
      ```go
        if dbname == existDb {
          continue
        }
        if CrossCheck(dbname, existDb) { // 这个函数返回 ture 表示存在冲突
          ... // 覆盖报告
        }
      ```
      这里可以看到, 如果申请的库和存在的库全等, 直接返回
    
      否则需要调用 `CrossCheck` 再生成冲突报告, 很耗时
    * 如果申请的是 _slave_ 权限, 裁剪掉写权限
    * 拼接其他的 _sql_ 语句部分
* 在目标实例执行生成的 _sql_
      
### _TenDBCluster_
