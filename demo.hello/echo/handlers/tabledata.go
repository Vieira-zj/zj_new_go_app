package handlers

import (
	"encoding/json"
	"net/http"
	"sort"

	"demo.hello/utils"
	"github.com/labstack/echo"
)

type dataResp struct {
	Code uint32      `json:"code"`
	Data interface{} `json:"data"`
}

type tableUser struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Role   string   `json:"role"`
	Skills []string `json:"skills"`
}

type tableRow struct {
	tableUser
	RowSpan uint32 `json:"rowspan"`
	ColSpan uint32 `json:"jsonspan"`
}

// GetTableRowSpanData returns table span rows for element ui table. Frontend: {js_project}/vue_pages/vue_apps/vue_spantable.html
func GetTableRowSpanData(c echo.Context) error {
	users, err := buildMockTableUsers()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	retData, err := addDefaultSpanValues(users)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	usersSpanByRole(retData)

	c.Response().Header().Add("Access-Control-Allow-Origin", "*")
	c.Response().Header().Add("Access-Control-Allow-Credentials", "true")
	return c.JSON(http.StatusOK, &dataResp{
		Code: 0,
		Data: retData,
	})
}

func buildMockTableUsers() ([]tableUser, error) {
	data := `[
{"id": "id01", "name": "name-001", "role": "tester", "skills": ["api-test", "function-test"]},
{"id": "id02", "name": "name-002", "role": "tester", "skills": ["api-test", "integration-test"]},
{"id": "id03", "name": "name-003", "role": "sre", "skills": ["linux", "docker"]},
{"id": "id04", "name": "name-004", "role": "tester", "skills": ["uat"]},
{"id": "id05", "name": "name-005", "role": "dev", "skills": ["javascript", "java"]},
{"id": "id06", "name": "name-006", "role": "dev", "skills": ["c++", "java"]},
{"id": "id07", "name": "name-007", "role": "tester", "skills": ["uat", "automation", "api-test"]},
{"id": "id08", "name": "name-008", "role": "sre", "skills": ["linux", "k8s"]},
{"id": "id09", "name": "name-009", "role": "dev", "skills": ["javascript", "java", "golang"]},
{"id": "id10", "name": "name-010", "role": "dev", "skills": ["java", "python"]},
{"id": "id11", "name": "name-011", "role": "manager", "skills": ["project"]}
]`
	var users []tableUser
	if err := json.Unmarshal([]byte(data), &users); err != nil {
		return nil, err
	}
	return users, nil
}

func addDefaultSpanValues(users []tableUser) ([]map[string]interface{}, error) {
	b, err := json.Marshal(users)
	if err != nil {
		return nil, err
	}

	rows := make([]map[string]interface{}, len(users))
	if err = json.Unmarshal(b, &rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		row["rowspan"] = 1
		row["colspan"] = 1
	}

	// TODO: convert to []tableRow instead of []map[string]interface{}
	return rows, nil
}

func usersSpanByRole(users []map[string]interface{}) {
	sort.SliceStable(users, func(i, j int) bool {
		srcRole := users[i]["role"].(string)
		dstRole := users[j]["role"].(string)
		return srcRole < dstRole
	})

	appendUserSkills := func(curUser, nextUser map[string]interface{}) {
		curUserSkills := curUser["skills"].([]interface{})
		nextUserSkills := nextUser["skills"].([]interface{})
		mergedSkills := append(curUserSkills, nextUserSkills...)
		skillsSet := utils.NewSet(10, mergedSkills...)
		curUser["skills"] = skillsSet.ToSlice()
	}

	// span默认为1, 被合并单元格的span设置为0
	for i := 0; i < (len(users) - 1); i++ {
		curUser := users[i]
		for i < (len(users) - 1) {
			nextUser := users[i+1]
			if curUser["role"] == nextUser["role"] {
				curUser["rowspan"] = curUser["rowspan"].(int) + 1
				nextUser["rowspan"] = 0
				nextUser["colspan"] = 0
				appendUserSkills(curUser, nextUser)
				i++
			} else {
				break
			}
		}
	}
}
