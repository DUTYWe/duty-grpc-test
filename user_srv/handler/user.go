package handler

import (
	"context"
	"crypto/sha512"
	"fmt"
	"mytest/user_srv/global"
	"mytest/user_srv/model"
	"mytest/user_srv/proto"
	"strings"
	"time"

	"github.com/anaskhan96/go-password-encoder"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type UserServer struct{}

func ModelToResponse(user model.User) proto.UserInfoResponse {
	//在grpc的message中字段有默认值，不能随便将nil赋值进去
	useInfoRsp := proto.UserInfoResponse{
		Id:       user.ID,
		Password: user.Password,
		NickName: user.NickName,
		Mobile:   user.Mobile,
		Gender:   user.Gender,
		Role:     int32(user.Role),
	}
	if user.Birthday != nil {
		useInfoRsp.Birthday = uint32(user.Birthday.Unix())
	}
	return useInfoRsp
}

func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page == 0 {
			page = 1
		}
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

//   db.Scopes(Paginate(r)).Find(&users)
//   db.Scopes(Paginate(r)).Find(&articles)

func (s *UserServer) GetUserList(ctx context.Context, req *proto.Pageinfo) (*proto.UserListResponse, error) {
	//获取用户列表
	var users []model.User
	result := global.DB.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	rsp := &proto.UserListResponse{}
	rsp.Total = int32(result.RowsAffected)
	global.DB.Scopes(Paginate(int(req.Pn), int(req.Psize))).Find(&users)

	for _, user := range users {
		userInfoRsp := ModelToResponse(user)
		rsp.Data = append(rsp.Data, &userInfoRsp)
	}
	return rsp, nil
}

func (s *UserServer) GetUserByMobile(ctx context.Context, req *proto.MobileRequest) (*proto.UserInfoResponse, error) {
	//通过手机号码查询用户
	var user model.User
	result := global.DB.Where(&model.User{Mobile: req.Mobile}).First(&user)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	if result.Error != nil {
		return nil, result.Error
	}

	userInfoRsp := ModelToResponse(user)

	return &userInfoRsp, nil
}
func (s *UserServer) GetUserById(ctx context.Context, req *proto.IdRequest) (*proto.UserInfoResponse, error) {
	//通过id查询用户
	var user model.User
	result := global.DB.First(&user, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	if result.Error != nil {
		return nil, result.Error
	}

	userInfoRsp := ModelToResponse(user)

	return &userInfoRsp, nil
}

func (s *UserServer) CreateUser(ctx context.Context, req *proto.CreateUserInfoRequest) (*proto.UserInfoResponse, error) {
	//新建用户
	//1、查询用户是否存在
	var user model.User
	result := global.DB.Where(&model.User{Mobile: req.Mobile}).First(&user)
	if result.RowsAffected == 1 {
		return nil, status.Errorf(codes.AlreadyExists, "用户已存在")
	}
	//2、用户不存在
	user.Mobile = req.Mobile
	user.NickName = req.NickName
	// user.CreateAt = time.Now()
	// user.UpdateAt = time.Now()
	//3、密码加密
	options := &password.Options{10, 100, 32, sha512.New}
	salt, encodedPwd := password.Encode(req.Password, options)
	user.Password = fmt.Sprintf("$pkbdf2-sha512$%s$%s", salt, encodedPwd)

	result = global.DB.Create(&user)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, result.Error.Error())
	}
	userInfoRsp := ModelToResponse(user)
	return &userInfoRsp, nil
}

func (s *UserServer) UpdateUser(ctx context.Context, req *proto.UpdateUserInfoRequest) (*empty.Empty, error) {
	//更新用户
	var user model.User
	result := global.DB.First(&user, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	birthday := time.Unix(int64(req.Birthday), 0)
	user.Birthday = &birthday
	user.NickName = req.NickName
	user.Gender = req.Gender
	result = global.DB.Save(user)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, result.Error.Error())
	}

	return &empty.Empty{}, nil
}

func (s *UserServer) CheckPassword(ctx context.Context, req *proto.CheckPasswordInfoRequest) (*proto.CheckResponse, error) {
	//校验密码
	passwordinfo := strings.Split(req.EncryptedPassword, "$")
	options := &password.Options{10, 100, 32, sha512.New}
	check := password.Verify(req.Password, passwordinfo[2], passwordinfo[3], options)
	return &proto.CheckResponse{Success: check}, nil
}
