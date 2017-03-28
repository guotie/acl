# ACL access control list

参考借鉴 shiro和spring security的权限管理

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

