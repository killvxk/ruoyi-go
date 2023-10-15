package config

import (
	"errors"
	"github.com/gin-gonic/gin"
	"lostvip.com/conf"
	"lostvip.com/utils/convert"
	"lostvip.com/utils/page"
	"robvi/app/common/cache"

	config2 "robvi/app/modules/sys/model/system/config"
	"robvi/app/modules/sys/service"
	"time"
)

func GetCopyright() string {
	ctx := conf.Config().GetConf("copyright")
	return ctx
}

func GetCtx() string {
	ctx := conf.Config().GetContextPath()
	return ctx
}
func GetOssUrl() string {
	ossUrl := GetValueByKey("sys.resource.url")
	if ossUrl == "" {
		ossUrl = "/static"
	}
	return ossUrl
}

// 根据用户id和权限字符串判断是否有此权限
func AddInt(a, b int) int {
	return a + b
}

// 根据键获取值
func GetValueByKey(key string) string {
	resultStr := ""
	//从缓存读取
	c := cache.Instance()
	result, ok := c.Get(key)

	if ok {
		return result.(string)
	}

	if result == nil {
		entity := &config2.Entity{ConfigKey: key}
		ok, _ := entity.FindOne()
		if !ok {
			return ""
		}

		resultStr = entity.ConfigValue
		c.Set(key, resultStr, 0)
	} else {
		resultStr = result.(string)
	}

	return resultStr
}

// 根据主键查询数据
func SelectRecordById(id int64) (*config2.Entity, error) {
	entity := &config2.Entity{ConfigId: id}
	_, err := entity.FindOne()
	return entity, err
}

// 根据主键删除数据
func DeleteRecordById(id int64) bool {
	entity := &config2.Entity{ConfigId: id}
	ok, _ := entity.FindOne()
	if ok {
		result, err := entity.Delete()
		if err == nil {
			if result > 0 {
				//从缓存删除
				c := cache.Instance()
				c.Delete(entity.ConfigKey)
				return true
			}
		}
	}
	return false
}

// 批量删除数据记录
func DeleteRecordByIds(ids string) int64 {
	idarr := convert.ToInt64Array(ids, ",")
	list, _ := config2.FindIn("config_id", idarr)
	rs, err := config2.DeleteBatch(idarr...)
	if err != nil {
		return 0
	}

	if len(list) > 0 {
		for _, item := range list {
			//从缓存删除
			c := cache.Instance()
			c.Delete(item.ConfigKey)
		}
	}

	return rs
}

// 添加数据
func AddSave(req *config2.AddReq, c *gin.Context) (int64, error) {
	var entity config2.Entity
	entity.ConfigName = req.ConfigName
	entity.ConfigKey = req.ConfigKey
	entity.ConfigType = req.ConfigType
	entity.ConfigValue = req.ConfigValue
	entity.Remark = req.Remark
	entity.CreateTime = time.Now()
	entity.CreateBy = ""
	var userService service.UserService
	user := userService.GetProfile(c)

	if user != nil {
		entity.CreateBy = user.LoginName
	}

	_, err := entity.Insert()
	return entity.ConfigId, err
}

// 修改数据
func EditSave(req *config2.EditReq, c *gin.Context) (int64, error) {
	entity := &config2.Entity{ConfigId: req.ConfigId}
	ok, err := entity.FindOne()

	if err != nil {
		return 0, err
	}

	if !ok {
		return 0, errors.New("数据不存在")
	}

	entity.ConfigName = req.ConfigName
	entity.ConfigKey = req.ConfigKey
	entity.ConfigValue = req.ConfigValue
	entity.Remark = req.Remark
	entity.ConfigType = req.ConfigType
	entity.UpdateTime = time.Now()
	entity.UpdateBy = ""
	var userService service.UserService
	user := userService.GetProfile(c)

	if user == nil {
		entity.UpdateBy = user.LoginName
	}

	rs, err := entity.Update()

	if err != nil {
		return 0, err
	}

	//保存到缓存
	cache := cache.Instance()
	cache.Set(entity.ConfigKey, entity.ConfigValue, 0)

	return rs, nil
}

// 根据条件分页查询角色数据
func SelectListAll(params *config2.SelectPageReq) ([]config2.Entity, error) {
	return config2.SelectListAll(params)
}

// 根据条件分页查询角色数据
func SelectListByPage(params *config2.SelectPageReq) ([]config2.Entity, *page.Paging, error) {
	return config2.SelectListByPage(params)
}

// 导出excel
func Export(param *config2.SelectPageReq) (string, error) {
	head := []string{"参数主键", "参数名称", "参数键名", "参数键值", "系统内置（Y是 N否）", "状态"}
	col := []string{"config_id", "config_name", "config_key", "config_value", "config_type"}
	return config2.SelectListExport(param, head, col)
}

// 检查角色名是否唯一
func CheckConfigKeyUniqueAll(configKey string) string {
	entity, err := config2.CheckPostCodeUniqueAll(configKey)
	if err != nil {
		return "1"
	}
	if entity != nil && entity.ConfigId > 0 {
		return "1"
	}
	return "0"
}

// 检查岗位名称是否唯一
func CheckConfigKeyUnique(configKey string, configId int64) string {
	entity, err := config2.CheckPostCodeUniqueAll(configKey)
	if err != nil {
		return "1"
	}
	if entity != nil && entity.ConfigId > 0 && entity.ConfigId != configId {
		return "1"
	}
	return "0"
}