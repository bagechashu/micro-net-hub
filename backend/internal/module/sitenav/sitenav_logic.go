package sitenav

import (
	"micro-net-hub/internal/module/sitenav/model"

	"github.com/gin-gonic/gin"
)

// GetGroups returns a list of groups.
func GetNavSites(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	// 构建 JSON 数据
	result := make(map[string]interface{})

	// 网址分组及其网址项查询
	var navGroups model.NavGroups
	if err := navGroups.FindWithItems(); err != nil {
		return nil, err
	}

	// 侧边分租解析
	for _, group := range navGroups {
		groupData := map[string]interface{}{
			"title": group.Title,
			"name":  group.Name,
			"nav":   group.NavItems,
		}
		result[group.Name] = groupData
	}

	// global.Log.Debugf("sitenav result: %+v", result)
	return result, nil
}

func AddSideNavGroup(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	return nil, nil
}

func UpdateSideNavGroup(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	return nil, nil
}

func DeleteSideNavGroup(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	return nil, nil
}

func AddNavGroup(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	return nil, nil
}

func UpdateNavGroup(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	return nil, nil
}

func DeleteNavGroup(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	return nil, nil
}

func AddNavItem(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	return nil, nil
}

func UpdateNavItem(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	return nil, nil
}

func DeleteNavItem(c *gin.Context, req interface{}) (data interface{}, rspError interface{}) {
	return nil, nil
}
