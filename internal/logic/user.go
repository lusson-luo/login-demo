package logic

import (
	"context"
	"errors"
	"fmt"
	"log"
	"login-demo/internal/dao"
	"login-demo/internal/model"
	"login-demo/internal/model/do"
	"login-demo/internal/model/entity"

	"crypto/sha256"

	"github.com/fatih/color"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type LogicUser struct {
}

var (
	User = LogicUser{}
)

// Login 登录
func (lu *LogicUser) Login(ctx context.Context, username string, password string) (role string, token string, err error) {
	// 计算密码哈希值
	passwordHash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	// 查询用户信息
	user, err := dao.User.Ctx(ctx).One("passport=? and password=?", username, passwordHash)
	switch {
	case err != nil:
		return "", "", err
	case user == nil:
		return "", "", errors.New("账户或密码错误")
	default:
		// 生成 jwt token
		token, err = JwtHandler.GenerateToken(ctx, username)
		if err != nil {
			return "", "", err
		}
		// todo: 暂时没有 role
		return "", token, nil
	}
}

type AdminInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// InitAdmin 初始化管理员账户，从配置文件中
func (*LogicUser) InitAdmin(ctx context.Context) {
	admin := &AdminInfo{}
	err := g.Cfg().MustGet(ctx, "admin").Scan(admin)
	if err != nil {
		log.Fatal("读取admin配置失败")
		return
	}
	var user *entity.User
	err = dao.User.Ctx(ctx).Where(do.User{
		Passport: admin.Username,
	}).Scan(&user)
	if err != nil {
		g.Log().Infof(ctx, color.RedString("err=%v, dao.User=%v"), err, user)
		panic("查询用户表失败，是否没有连接正确的数据库")
	}
	if user == nil {
		dao.User.Ctx(ctx).Insert(do.User{
			Passport: admin.Username,
			Password: fmt.Sprintf("%x", sha256.Sum256([]byte(admin.Password))),
			Nickname: admin.Username,
		})
	}
}

func (*LogicUser) UserList(ctx context.Context, username string, page model.PageReq) (users []entity.User, count int) {
	dao.User.Ctx(ctx).WhereLike("passport", fmt.Sprintf("%%%s%%", username)).Page(page.PageNo, page.PageSize).Scan(&users)
	count, _ = dao.User.Ctx(ctx).Count("passport like ?", fmt.Sprintf("%%%s%%", username))
	return
}

func (*LogicUser) Add(ctx context.Context, user do.User) (err error) {
	// 计算密码哈希值
	passwordHash := fmt.Sprintf("%x", sha256.Sum256([]byte(user.Password.(string))))
	user.CreateAt, user.UpdateAt, user.Password = gtime.New(), gtime.New(), passwordHash
	_, err = dao.User.Ctx(ctx).Insert(user)
	return
}

func (*LogicUser) Del(ctx context.Context, id int) (err error) {
	_, err = dao.User.Ctx(ctx).Delete("id = ?", id)
	return
}

func (*LogicUser) Update(ctx context.Context, user do.User) (err error) {
	// 计算密码哈希值
	passwordHash := fmt.Sprintf("%x", sha256.Sum256([]byte(user.Password.(string))))
	user.UpdateAt, user.Password = gtime.New(), passwordHash
	_, err = dao.User.Ctx(ctx).Update(user, "id = ?", user.Id)
	return
}
