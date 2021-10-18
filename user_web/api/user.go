package api

import (
	"context"
	"fmt"
	"mytest/user_web/forms"
	"mytest/user_web/global"
	"mytest/user_web/global/response"
	"mytest/user_web/middlewares"
	"mytest/user_web/models"
	"mytest/user_web/proto"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func removeTopStruct(fileds map[string]string) map[string]string {
	rsp := map[string]string{}
	for field, err := range fileds {
		rsp[field[strings.Index(field, ".")+1:]] = err
	}
	return rsp
}

func HandleGrpcErrorToHttp(err error, c *gin.Context) {
	//将grpc的code转换成http的状态码
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{
					"msg:": e.Message(),
				})
			case codes.Internal:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg:": "内部错误",
				})
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{
					"msg:": "参数错误",
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg:": "其他错误",
				})
			}
			return
		}
	}
}

func HandleValidatorError(c *gin.Context, err error) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"msg": err.Error(),
		})
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"error": removeTopStruct(errs.Translate(global.Trans)),
	})
}

func GetUserList(ctx *gin.Context) {
	// ip := "127.0.0.1"
	// port := 50051
	//拨号连接用户grpc服务
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", global.ServerConfig.Usersrvconfig.Host, global.ServerConfig.Usersrvconfig.Port), grpc.WithInsecure())
	if err != nil {
		zap.S().Errorw("[GetUserList] 连接 【用户服务失败】",
			"msg", err.Error())
	}
	claims, _ := ctx.Get("claims")
	zap.S().Infof("访问用户：%d", claims.(*models.CustomClaims).ID)
	//调用服务
	usersrvClient := proto.NewUserClient(conn)

	pn := ctx.DefaultQuery("pn", "0")
	pnInt, _ := strconv.Atoi(pn)
	psize := ctx.DefaultQuery("psize", "10")
	psizeInt, _ := strconv.Atoi(psize)
	rsp, err := usersrvClient.GetUserList(context.Background(), &proto.Pageinfo{
		Pn:    uint32(pnInt),
		Psize: uint32(psizeInt),
	})
	if err != nil {
		zap.S().Errorw("[GetUserList] 查询 【用户列表】失败")
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	result := make([]interface{}, 0)
	for _, v := range rsp.Data {
		// data := make(map[string]interface{})
		resp := response.UserResponse{
			Id:       v.Id,
			NickName: v.NickName,
			Birthday: response.JsonTime(time.Unix(int64(v.Birthday), 0)),
			Gender:   v.Gender,
			Mobile:   v.Mobile,
		}
		// data["id"] =
		// data["name"] =
		// data["birthday"] =
		// data["gender"] =
		// data["mobile"] =

		result = append(result, resp)
	}
	ctx.JSON(http.StatusOK, result)
}

func PasswordLogin(ctx *gin.Context) {
	//表单验证
	passwordloginform := forms.PasswordLoginForm{}
	if err := ctx.ShouldBind(&passwordloginform); err != nil {
		HandleValidatorError(ctx, err)
		return
	}
	//图形验证码
	if !store.Verify(passwordloginform.CaptchaId, passwordloginform.Captcha, true) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"captcha": "验证码错误",
		})
		return
	}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", global.ServerConfig.Usersrvconfig.Host, global.ServerConfig.Usersrvconfig.Port), grpc.WithInsecure())
	if err != nil {
		zap.S().Errorw("[GetUserList] 连接 【用户服务失败】",
			"msg", err.Error())
	}
	//调用服务
	usersrvClient := proto.NewUserClient(conn)

	//登录的逻辑
	rsp, err := usersrvClient.GetUserByMobile(context.Background(), &proto.MobileRequest{
		Mobile: passwordloginform.Mobile,
	})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				ctx.JSON(http.StatusBadRequest, map[string]string{
					"mobile": "用户不存在",
				})
			default:
				ctx.JSON(http.StatusInternalServerError, map[string]string{
					"mobile": "登陆失败",
				})
			}
			return
		}
	} else {
		//只是查询到了用户，并没有检查密码
		if resp, rsperr := usersrvClient.CheckPassword(context.Background(), &proto.CheckPasswordInfoRequest{
			Password:          passwordloginform.Password,
			EncryptedPassword: rsp.Password,
		}); rsperr != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": "登录失败",
			})
		} else {
			if resp.Success {
				//生成Tocken
				j := *middlewares.NewJWT()
				claims := models.CustomClaims{
					ID:          uint(rsp.Id),
					NickName:    rsp.NickName,
					AuthorityId: uint(rsp.Role),
					StandardClaims: jwt.StandardClaims{
						NotBefore: time.Now().Unix(),               //签名生效时间
						ExpiresAt: time.Now().Unix() + 60*60*24*20, //20天过期
						Issuer:    "duty",
					},
				}
				token, err := j.CreateToken(claims)
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{
						"msg": "生成token失败",
					})
					return
				}

				ctx.JSON(http.StatusOK, gin.H{
					"id":         rsp.Id,
					"nick_name":  rsp.NickName,
					"token":      token,
					"expired_at": (time.Now().Unix() + 60*60*24*20) * 1000,
				})
			} else {
				ctx.JSON(http.StatusBadRequest, map[string]string{
					"password": "密码错误",
				})
			}

		}
	}
}

//用户注册
func Regisetr(ctx *gin.Context) {
	//表单验证
	RegisterForm := forms.RegisterForm{}
	if err := ctx.ShouldBind(&RegisterForm); err != nil {
		HandleValidatorError(ctx, err)
		return
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", global.ServerConfig.Usersrvconfig.Host, global.ServerConfig.Usersrvconfig.Port), grpc.WithInsecure())
	if err != nil {
		zap.S().Errorw("[Regisetr] 连接 【用户服务失败】",
			"msg", err.Error())
	}
	//调用服务
	usersrvClient := proto.NewUserClient(conn)
	user, err := usersrvClient.CreateUser(context.Background(), &proto.CreateUserInfoRequest{
		NickName: RegisterForm.Mobile,
		Password: RegisterForm.Password,
		Mobile:   RegisterForm.Mobile,
	})

	if err != nil {
		zap.S().Errorf("[Regisetr] 查询 【用户注册】失败:%s", err.Error())
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	j := *middlewares.NewJWT()
	claims := models.CustomClaims{
		ID:          uint(user.Id),
		NickName:    user.NickName,
		AuthorityId: uint(user.Role),
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix(),               //签名生效时间
			ExpiresAt: time.Now().Unix() + 60*60*24*20, //20天过期
			Issuer:    "duty",
		},
	}
	token, err := j.CreateToken(claims)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "生成token失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id":         user.Id,
		"nick_name":  user.NickName,
		"token":      token,
		"expired_at": (time.Now().Unix() + 60*60*24*20) * 1000,
	})
}
