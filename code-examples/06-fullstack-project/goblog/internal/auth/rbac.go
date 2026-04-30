package auth

// 角色常量定义
const (
	RoleAdmin  = "admin"  // 管理员：拥有所有权限
	RoleAuthor = "author" // 作者：可创建/编辑文章和标签
	RoleReader = "reader" // 读者：可浏览和评论
)

// 权限常量定义
const (
	PermManageUsers         = "manage_users"          // 用户管理
	PermReviewArticles      = "review_articles"       // 文章审核
	PermViewStats           = "view_stats"            // 系统统计
	PermCreateArticle       = "create_article"        // 创建文章
	PermEditOwnArticle      = "edit_own_article"      // 编辑自己的文章
	PermEditAllArticles     = "edit_all_articles"     // 编辑所有文章
	PermCreateTag           = "create_tag"            // 创建标签
	PermCreateComment       = "create_comment"        // 发表评论
	PermDeleteOwnComment    = "delete_own_comment"    // 删除自己的评论
	PermDeleteAllComments   = "delete_all_comments"   // 删除所有评论
	PermBrowse              = "browse"                // 浏览文章/标签/评论
)

// rolePermissions 角色权限矩阵
// 定义每个角色拥有的权限集合
var rolePermissions = map[string]map[string]bool{
	RoleAdmin: {
		PermManageUsers:       true,
		PermReviewArticles:    true,
		PermViewStats:         true,
		PermCreateArticle:     true,
		PermEditOwnArticle:    true,
		PermEditAllArticles:   true,
		PermCreateTag:         true,
		PermCreateComment:     true,
		PermDeleteOwnComment:  true,
		PermDeleteAllComments: true,
		PermBrowse:            true,
	},
	RoleAuthor: {
		PermCreateArticle:    true,
		PermEditOwnArticle:   true,
		PermCreateTag:        true,
		PermCreateComment:    true,
		PermDeleteOwnComment: true,
		PermBrowse:           true,
	},
	RoleReader: {
		PermCreateComment:    true,
		PermDeleteOwnComment: true,
		PermBrowse:           true,
	},
}

// CheckPermission 检查指定角色是否拥有指定权限
func CheckPermission(role, permission string) bool {
	perms, ok := rolePermissions[role]
	if !ok {
		return false
	}
	return perms[permission]
}

// HasRole 检查用户角色是否在允许的角色列表中
func HasRole(userRole string, allowedRoles ...string) bool {
	for _, role := range allowedRoles {
		if userRole == role {
			return true
		}
	}
	return false
}

// ValidRoles 返回所有有效角色列表
func ValidRoles() []string {
	return []string{RoleAdmin, RoleAuthor, RoleReader}
}

// IsValidRole 检查角色是否有效
func IsValidRole(role string) bool {
	for _, r := range ValidRoles() {
		if r == role {
			return true
		}
	}
	return false
}
