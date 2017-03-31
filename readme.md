# ACL access control list

参考借鉴 shiro和spring security的权限管理, 100% 测试覆盖

## ACL权限判断

```
func (mgr *AclManager) IsGrant(who AclObject, what AclObject, perm Permission) bool
```
流程：

0. 首先，根据who查找who的principls;
1. 查找who的每一个principls，是否有权限do waht;
2. 

## rule权限判断

在有些时候，ACL判断规则会比较繁琐，例如：

在分级权限管理中，假如分为四级：总部级，省级，市级，县级。
省级管理人员可以管理下面所有市级、县级的营业厅，市级管理人员可以管理所有的本市的营业厅和下面县的营业厅，A市共有100家营业厅。

如果按照ACL的规则，则需要给省级管理人员配置所有营业厅的管理权限，就是需要N条记录。而且，每增加、删除营业厅，都需要重新修改权限
的ACL规则。

按照rule的配置规则，则仅仅需要编写一个函数，判断省级的管理人员是否有该营业厅的权限即可。

当然，rule函数规则和业务关联紧密，需要在代码中写死，灵活性欠缺。


## 缓存

查询权限时，优先从缓存中查找。

# 用法说明

## 权限主体 Account

你需要自己定义Account Model，并为Account实现 Principal 接口的两个方法：

```
type AclObject interface {
	GetSid() string // 通常是 uuid
	GetTyp() string // object类型, user, role, post, etc...
}

type Principal interface {
	AclObject
}

```

Account 的 GetSid方法返回Account的Sid，通常为UUID
Account 的 GetTyp方法返回 _*acl.AccountTyp*_, 为"AccountTyp"

## 权限标的

Account 要操作的权限标的必须实现 AclObject 接口.

权限标的 GetTyp 方法返回权限标的 类型。

## 角色Role

角色最重要的两个接口：

1. 创建角色
```
func (mgr *AclManager) CreateRole(name, sid, depart, corp, city, province string, level int) (*Role, error) {
```
- name:     角色名称
- sid:      角色全局id, 通常为空, 由uuid自动生成
- depart:   角色所在部门
- corp:     角色所在公司
- city:     角色所在城市
- province: 角色所在省份
- level:    角色级别

2. 将Account加入到角色中：
```
func (mgr *AclManager) AddUserRoleRelation(uid, rid string) error
```
- uid: Account的uuid
- rid: role的uuid

## Permission

Permission 比较简单, 核心是一个类型为uint32, 名为Mask的字段。 通过 请求的操作权限 与 AclEntry 中的权限做 & 操作, 来决定是否具备操作权限

Mask每一位代表一种权限，目前已定义的权限如下：
```
	PermissionRead   = 1 << permIndexRead   // 读权限
	PermissionWrite  = 1 << permIndexWrite  // 写权限
	PermissionCreate = 1 << permIndexCreate // 创建权限
	PermissionDelete = 1 << permIndexDelete // 删除权限
	PermissionManage = 1 << permIndexManage // 管理权限
```

## 权限访问控制列表 AclEntry

通过AclEntry来控制一个账号或一个角色是否对权限标的有特定的权限

创建ACL entry：
```
func (mgr *AclManager) CreateACE(whoSid, whoTyp, whatSid, whatTyp    string, mask uint32, order int,
	grant, auditSuccess, auditFailure bool) (*AclEntry, error)
```

- whoSid:      权限主体sid
- whoTyp:      权限主体类型, 通常为AccountTyp或RoleTyp
- whatSid:     权限标的sid
- whatTyp:     权限标的类型
- mask:        权限
- order:       创建的Acl在Acl list的顺序, 判断权限时, AclEntry安装order从小到大逐条判定
- grant:       授权或拒绝
- auditSuccess: 当授权时, 是否审计
- auditFailure: 当拒绝时, 是否审计

## AclManager
```
func CreateAclManager(db *gorm.DB, pool *redis.Pool) *AclManager 
```
CreateAclManager 用来创建 AclManager 对象

在上面的工作都准备好以后，就可以通过IsGrant接口来查询权限：
```
func (mgr *AclManager) IsGrant(who AclObject, what AclObject, perm Permission) bool {
```
- who:  权限主体, 可以是Account，也可以是Role
- what: 权限标的
- perm: 请求的权限
- 返回： true: 授权 false：拒绝

