package plan

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cectc/dbpack/pkg/visitor"
	"github.com/cectc/dbpack/third_party/parser"
	"github.com/cectc/dbpack/third_party/parser/ast"
)

func TestQueryOnSingleDBPlan(t *testing.T) {
	testCases := []struct {
		selectSql           string
		tables              []string
		args                []interface{}
		expectedGenerateSql string
	}{
		{
			selectSql:           "select * from student where id in (?,?)",
			tables:              []string{"student_1", "student_5"},
			args:                []interface{}{1, 5},
			expectedGenerateSql: "(SELECT * FROM `student_1` WHERE `id` IN (?,?)) UNION ALL (SELECT * FROM `student_5` WHERE `id` IN (?,?))",
		},
		{
			selectSql:           "select * from student where id in (?,?) order by id desc",
			tables:              []string{"student_1", "student_5"},
			args:                []interface{}{1, 5},
			expectedGenerateSql: "SELECT * FROM ((SELECT * FROM `student_1` WHERE `id` IN (?,?) ORDER BY `id` DESC) UNION ALL (SELECT * FROM `student_5` WHERE `id` IN (?,?) ORDER BY `id` DESC)) t ORDER BY `id` DESC",
		},
		{
			selectSql:           "select * from student where id in (?,?) order by id desc limit ?, ?",
			tables:              []string{"student_1", "student_5"},
			args:                []interface{}{1, 5, 1000, 20},
			expectedGenerateSql: "SELECT * FROM ((SELECT * FROM `student_1` WHERE `id` IN (?,?) ORDER BY `id` DESC limit 1020) UNION ALL (SELECT * FROM `student_5` WHERE `id` IN (?,?) ORDER BY `id` DESC limit 1020)) t ORDER BY `id` DESC",
		},
	}

	for _, c := range testCases {
		t.Run(c.selectSql, func(t *testing.T) {
			p := parser.New()
			stmt, err := p.ParseOneStmt(c.selectSql, "", "")
			if err != nil {
				t.Error(err)
				return
			}
			stmt.Accept(&visitor.ParamVisitor{})
			selectStmt := stmt.(*ast.SelectStmt)
			plan := &QueryOnSingleDBPlan{
				Database: "school_0",
				Tables:   c.tables,
				Stmt:     selectStmt,
				Args:     c.args,
				Executor: nil,
			}
			var (
				sb   strings.Builder
				args []interface{}
			)
			plan.castLimit()
			err = plan.generate(&sb, &args)
			assert.Nil(t, err)
			assert.Equal(t, c.expectedGenerateSql, sb.String())
		})
	}
}