// Go 1.22+ | 验证日期：2025-01-01
// Mock 测试示例
// 演示手动 Mock 实现接口进行单元测试
// 无需第三方 Mock 框架，通过手动实现接口即可完成 Mock
package mock

import (
	"errors"
	"testing"
)

// ============================================================
// 手动 Mock 实现（推荐简单场景使用）
// ============================================================

// fakeUserRepo 手动实现的 Mock UserRepository
// 对于简单接口，手动实现比使用 Mock 框架更清晰
type fakeUserRepo struct {
	users   map[int]*User // 模拟数据存储
	nextID  int           // 自增 ID
	createErr error       // 模拟创建错误
	updateErr error       // 模拟更新错误
}

// newFakeUserRepo 创建 fake 仓库
func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

// GetByID 根据 ID 获取用户
func (f *fakeUserRepo) GetByID(id int) (*User, error) {
	user, ok := f.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	// 返回副本，避免外部修改影响内部数据
	copy := *user
	return &copy, nil
}

// Create 创建用户
func (f *fakeUserRepo) Create(user *User) error {
	if f.createErr != nil {
		return f.createErr
	}
	user.ID = f.nextID
	f.nextID++
	// 存储副本
	copy := *user
	f.users[user.ID] = &copy
	return nil
}

// Update 更新用户
func (f *fakeUserRepo) Update(user *User) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	if _, ok := f.users[user.ID]; !ok {
		return ErrNotFound
	}
	copy := *user
	f.users[user.ID] = &copy
	return nil
}

// addUser 辅助方法：预置用户数据
func (f *fakeUserRepo) addUser(user *User) {
	f.users[user.ID] = user
}

// ============================================================
// 测试用例
// ============================================================

// TestGetUser_Success 测试正常获取用户
func TestGetUser_Success(t *testing.T) {
	repo := newFakeUserRepo()
	repo.addUser(&User{ID: 1, Name: "Alice", Email: "alice@example.com"})

	svc := NewUserService(repo)
	user, err := svc.GetUser(1)

	if err != nil {
		t.Fatalf("期望无错误，得到: %v", err)
	}
	if user.Name != "Alice" {
		t.Errorf("用户名 = %s, 期望 Alice", user.Name)
	}
	if user.Email != "alice@example.com" {
		t.Errorf("邮箱 = %s, 期望 alice@example.com", user.Email)
	}
}

// TestGetUser_NotFound 测试用户不存在
func TestGetUser_NotFound(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewUserService(repo)

	_, err := svc.GetUser(999)

	if err == nil {
		t.Fatal("期望返回错误，得到 nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("期望 ErrNotFound 错误，得到: %v", err)
	}
}

// TestGetUser_InvalidID 测试无效 ID
func TestGetUser_InvalidID(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewUserService(repo)

	tests := []struct {
		name string
		id   int
	}{
		{"零值 ID", 0},
		{"负数 ID", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetUser(tt.id)
			if err == nil {
				t.Error("期望返回错误，得到 nil")
			}
		})
	}
}

// TestCreateUser 表驱动测试创建用户
func TestCreateUser(t *testing.T) {
	tests := []struct {
		name      string
		userName  string
		email     string
		createErr error // 模拟仓库层错误
		wantErr   bool
	}{
		{"正常创建", "Bob", "bob@example.com", nil, false},
		{"空用户名", "", "bob@example.com", nil, true},
		{"空邮箱", "Bob", "", nil, true},
		{"仓库层错误", "Bob", "bob@example.com", ErrDuplicateEmail, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeUserRepo()
			repo.createErr = tt.createErr

			svc := NewUserService(repo)
			user, err := svc.CreateUser(tt.userName, tt.email)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if user == nil {
					t.Fatal("期望返回用户，得到 nil")
				}
				if user.Name != tt.userName {
					t.Errorf("用户名 = %s, 期望 %s", user.Name, tt.userName)
				}
				if user.Email != tt.email {
					t.Errorf("邮箱 = %s, 期望 %s", user.Email, tt.email)
				}
				if user.ID <= 0 {
					t.Errorf("用户 ID = %d, 期望 > 0", user.ID)
				}
			}
		})
	}
}

// TestUpdateUserName 测试更新用户名称
func TestUpdateUserName(t *testing.T) {
	tests := []struct {
		name    string
		userID  int
		newName string
		setup   func(*fakeUserRepo) // 预置数据
		wantErr bool
	}{
		{
			name:    "正常更新",
			userID:  1,
			newName: "Alice Updated",
			setup: func(repo *fakeUserRepo) {
				repo.addUser(&User{ID: 1, Name: "Alice", Email: "alice@example.com"})
			},
			wantErr: false,
		},
		{
			name:    "用户不存在",
			userID:  999,
			newName: "Nobody",
			setup:   func(repo *fakeUserRepo) {},
			wantErr: true,
		},
		{
			name:    "空用户名",
			userID:  1,
			newName: "",
			setup: func(repo *fakeUserRepo) {
				repo.addUser(&User{ID: 1, Name: "Alice", Email: "alice@example.com"})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeUserRepo()
			tt.setup(repo)

			svc := NewUserService(repo)
			err := svc.UpdateUserName(tt.userID, tt.newName)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			// 验证更新后的数据
			if !tt.wantErr {
				user, _ := repo.GetByID(tt.userID)
				if user.Name != tt.newName {
					t.Errorf("更新后用户名 = %s, 期望 %s", user.Name, tt.newName)
				}
			}
		})
	}
}
