package user

import (
	"github.com/gin-gonic/gin"
	response2 "lostvip.com/utils/response"
	"net/http"
	"os"
	"robvi/app/modules/sys/model"
	userModel "robvi/app/modules/sys/model/system/user"
	"robvi/app/modules/sys/service"
	"strconv"
	"time"
)

// 用户资料页面
func Profile(c *gin.Context) {
	var userService service.UserService
	user := userService.GetProfile(c)
	response2.BuildTpl(c, "system/user/profile/profile").WriteTpl(gin.H{
		"user": user,
	})
}

// 修改用户信息
func Update(c *gin.Context) {
	var req *userModel.ProfileReq
	//获取参数
	if err := c.ShouldBind(&req); err != nil {
		response2.ErrorResp(c).SetBtype(model.Buniss_Edit).SetMsg(err.Error()).Log("用户管理", req).WriteJsonExit()
		return
	}
	var userService service.UserService
	err := userService.UpdateProfile(req, c)

	if err != nil {
		response2.ErrorResp(c).SetBtype(model.Buniss_Edit).SetMsg(err.Error()).Log("用户管理", req).WriteJsonExit()
	} else {
		response2.SucessResp(c).SetBtype(model.Buniss_Edit).Log("用户管理", req).WriteJsonExit()
	}
}

// 修改用户密码
func UpdatePassword(c *gin.Context) {
	var req *userModel.PasswordReq
	if err := c.ShouldBind(&req); err != nil {
		response2.ErrorResp(c).SetBtype(model.Buniss_Edit).SetMsg(err.Error()).Log("用户管理", req).WriteJsonExit()
	}
	var userService service.UserService
	err := userService.UpdatePassword(req, c)

	if err != nil {
		response2.ErrorResp(c).SetBtype(model.Buniss_Edit).SetMsg(err.Error()).Log("用户管理", req).WriteJsonExit()
	} else {
		response2.SucessResp(c).SetBtype(model.Buniss_Edit).Log("修改用户密码", req).WriteJsonExit()
	}
}

// 修改头像页面
func Avatar(c *gin.Context) {
	var userService service.UserService
	user := userService.GetProfile(c)
	response2.BuildTpl(c, "system/user/profile/avatar").WriteTpl(gin.H{
		"user": user,
	})
}

// 修改密码页面
func EditPwd(c *gin.Context) {
	var userService service.UserService
	user := userService.GetProfile(c)
	response2.BuildTpl(c, "system/user/profile/resetPwd").WriteTpl(gin.H{
		"user": user,
	})
}

// 检查登陆名是否存在
func CheckLoginNameUnique(c *gin.Context) {
	var req *userModel.CheckLoginNameReq
	if err := c.ShouldBind(&req); err != nil {
		c.Writer.WriteString("1")
		return
	}
	var userService service.UserService
	result := userService.CheckLoginName(req.LoginName)

	if result {
		c.Writer.WriteString("1")
	} else {
		c.Writer.WriteString("0")
	}
}

// 检查邮箱是否存在
func CheckEmailUnique(c *gin.Context) {
	var req *userModel.CheckEmailReq
	if err := c.ShouldBind(&req); err != nil {
		c.Writer.WriteString("1")
		return
	}
	var userService service.UserService
	result := userService.CheckEmailUnique(req.UserId, req.Email)

	if result {
		c.Writer.WriteString("1")
	} else {
		c.Writer.WriteString("0")
	}
}

// 检查邮箱是否存在
func CheckEmailUniqueAll(c *gin.Context) {
	var req *userModel.CheckEmailAllReq
	if err := c.ShouldBind(&req); err != nil {
		c.Writer.WriteString("1")
		return
	}
	var userService service.UserService
	result := userService.CheckEmailUniqueAll(req.Email)

	if result {
		c.Writer.WriteString("1")
	} else {
		c.Writer.WriteString("0")
	}
}

// 检查手机号是否存在
func CheckPhoneUnique(c *gin.Context) {
	var req *userModel.CheckPhoneReq
	if err := c.ShouldBind(&req); err != nil {
		c.Writer.WriteString("1")
		return
	}
	var userService service.UserService
	result := userService.CheckPhoneUnique(req.UserId, req.Phonenumber)

	if result {
		c.Writer.WriteString("1")
	} else {
		c.Writer.WriteString("0")
	}

}

// 检查手机号是否存在
func CheckPhoneUniqueAll(c *gin.Context) {
	var req *userModel.CheckPhoneAllReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, model.CommonRes{
			Code: 500,
			Msg:  err.Error(),
		})
	}
	var userService service.UserService
	result := userService.CheckPhoneUniqueAll(req.Phonenumber)

	if result {
		c.Writer.WriteString("1")
	} else {
		c.Writer.WriteString("0")
	}

}

// 校验密码是否正确
func CheckPassword(c *gin.Context) {
	var req *userModel.CheckPasswordReq
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusOK, model.CommonRes{
			Code: 500,
			Msg:  err.Error(),
		})
	}
	var userService service.UserService
	user := userService.GetProfile(c)

	result := userService.CheckPassword(user, req.Password)

	if result {
		c.Writer.WriteString("true")
	} else {
		c.Writer.WriteString("false")
	}
}

// 保存头像
func UpdateAvatar(c *gin.Context) {
	var userService service.UserService
	user := userService.GetProfile(c)

	curDir, err := os.Getwd()

	if err != nil {
		response2.ErrorResp(c).SetBtype(model.Buniss_Edit).SetMsg(err.Error()).Log("保存头像", gin.H{"userid": user.UserId}).WriteJsonExit()
	}

	saveDir := curDir + "/public/upload/"

	fileHead, err := c.FormFile("avatarfile")

	if err != nil {
		response2.ErrorResp(c).SetBtype(model.Buniss_Edit).SetMsg("没有获取到上传文件").Log("保存头像", gin.H{"userid": user.UserId}).WriteJsonExit()
	}

	curdate := time.Now().UnixNano()
	filename := user.LoginName + strconv.FormatInt(curdate, 10) + ".png"
	dts := saveDir + filename

	if err := c.SaveUploadedFile(fileHead, dts); err != nil {
		response2.ErrorResp(c).SetBtype(model.Buniss_Edit).SetMsg(err.Error()).Log("保存头像", gin.H{"userid": user.UserId}).WriteJsonExit()
	}

	avatar := "/upload/" + filename

	err = userService.UpdateAvatar(avatar, c)

	if err != nil {
		response2.ErrorResp(c).SetBtype(model.Buniss_Edit).SetMsg(err.Error()).Log("保存头像", gin.H{"userid": user.UserId}).WriteJsonExit()
	} else {
		response2.SucessResp(c).SetBtype(model.Buniss_Edit).Log("保存头像", gin.H{"userid": user.UserId}).WriteJsonExit()
	}
}
