package acl

// 类似于spring security的角色层次关系
// 例如：
//    A > B, 角色A包含角色B的所有权限
//    B > C, 角色B包含角色C的所有权限
//    A > D, 角色A包含角色D的所有权限
//    B > E, 角色B包含角色E的所有权限
//  通过上面的层次关系，可以降低重复的权限定义
//
// 暂未实现
//
