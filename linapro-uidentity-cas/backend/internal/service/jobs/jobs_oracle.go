// This file maps the old Oracle-backed uidentity/admin jobs into LinaPro
// managed scheduled-job handlers. Oracle records are read in pages and converted
// into plugin-owned account, detail, group, and unit writes.

package jobs

import (
	"context"
	"strconv"
	"time"

	_ "github.com/gogf/gf/contrib/drivers/oracle/v2"
	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/logger"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

const (
	oracleTableBZKS = "T_BZKXS"
	oracleTableYJS  = "V_YJSJCXX_FULL"
	oracleTableWJ   = "WJ_STU_INFO"
	oracleTableJZG  = "V_RS_V_JZGJCSJXX_TEMP"
	oracleTableDept = "T_ORG_DEPT"
)

type oracleStudentInfo struct {
	XH     string `orm:"XH"`
	XM     string `orm:"XM"`
	XB     string `orm:"XB"`
	SFZHM  string `orm:"SFZHM"`
	CSRQ   string `orm:"CSRQ"`
	NJ     int    `orm:"NJ"`
	ZYMC   string `orm:"ZYMC"`
	BJMC   string `orm:"BJMC"`
	XYMC   string `orm:"XYMC"`
	XYDM   string `orm:"XYDM"`
	ZDQK   string `orm:"ZDQK"`
	XQ     string `orm:"XQ"`
	XZ     int    `orm:"XZ"`
	YJBYSJ int    `orm:"YJBYSJ"`
	ZP     string `orm:"ZP"`
	SJH    string `orm:"SJH"`
}

type oracleStudentYJS struct {
	Xh     string `orm:"XH"`
	XsXm   string `orm:"XS_XM"`
	XbMc   string `orm:"XB_MC"`
	Zjhm   string `orm:"ZJHM"`
	XsYxDm string `orm:"XS_YX_DM"`
	XsYxMc string `orm:"XS_YX_MC"`
	XsZyMc string `orm:"XS_ZY_MC"`
	Sfzj   string `orm:"SFZJ"`
	Xq     string `orm:"XQ"`
	Nj     int    `orm:"NJ"`
	Sjhm   string `orm:"SJHM"`
}

type oracleStudentWJ struct {
	XH    string `orm:"XH"`
	XM    string `orm:"XM"`
	SFZH  string `orm:"SFZH"`
	YDDH  string `orm:"YDDH"`
	QQ    string `orm:"QQ"`
	Email string `orm:"EMAIL"`
	XJZT  string `orm:"XJZT"`
}

type oracleStaffJZG struct {
	GH          string    `orm:"GH"`
	DWH         string    `orm:"DWH"`
	XM          string    `orm:"XM"`
	XBM         string    `orm:"XBM"`
	CSRQ        time.Time `orm:"CSRQ"`
	SFZJH       string    `orm:"SFZJH"`
	Email       string    `orm:"EMAIL"`
	RYZTM       string    `orm:"RYZTM"`
	XQ          string    `orm:"XQ"`
	MobilePhone string    `orm:"MOBILE_PHONE"`
}

type oracleDept struct {
	DeptCode string `orm:"DEPT_CODE"`
	DeptName string `orm:"DEPT_NAME"`
}

func (s *serviceImpl) syncStudent(ctx context.Context) error {
	return s.syncOracleStudentInfo(ctx, oracleTableBZKS, s.studentInput)
}

func (s *serviceImpl) syncStudentYJS(ctx context.Context) error {
	return s.syncOracleStudentYJS(ctx)
}

func (s *serviceImpl) syncStudentWJ(ctx context.Context) error {
	return s.syncOracleStudentWJ(ctx)
}

func (s *serviceImpl) syncJzg(ctx context.Context) error {
	return s.syncOracleStaff(ctx)
}

func (s *serviceImpl) syncDept(ctx context.Context) error {
	db, err := s.oracleDB(ctx)
	if err != nil {
		return err
	}
	pageSize := defaultPageSize
	tenantID := s.tenantID(ctx)
	stats := jobRunStats{}
	for page := 0; ; page++ {
		var rows []*oracleDept
		if err := oraclePageModel(db, oracleTableDept, page, pageSize).Scan(&rows); err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		if err := s.upsertOracleUnits(ctx, tenantID, rows, &stats); err != nil {
			return err
		}
		if len(rows) < pageSize {
			break
		}
	}
	logger.Infof(ctx, "uidentity oracle dept sync finished tenant=%d stats=%v", tenantID, sqlLogFields(stats))
	return nil
}

func (s *serviceImpl) syncOracleStudentInfo(ctx context.Context, table string, convert func(*oracleStudentInfo) *accountSyncInput) error {
	db, err := s.oracleDB(ctx)
	if err != nil {
		return err
	}
	syncCtx, err := s.accountSyncContext(ctx, s.tenantID(ctx), legacyContainerStudent, legacyContainerStudent)
	if err != nil {
		return err
	}
	stats := jobRunStats{}
	for page := 0; ; page++ {
		var rows []*oracleStudentInfo
		if err := oraclePageModel(db, table, page, defaultPageSize).Scan(&rows); err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		inputs := make([]*accountSyncInput, 0, len(rows))
		for _, row := range rows {
			inputs = append(inputs, convert(row))
		}
		pageStats, err := s.syncOracleAccounts(ctx, syncCtx, inputs)
		if err != nil {
			return err
		}
		stats.add(pageStats)
		if len(rows) < defaultPageSize {
			break
		}
	}
	logger.Infof(ctx, "uidentity oracle student sync finished stats=%v", sqlLogFields(stats))
	return nil
}

func (s *serviceImpl) syncOracleStudentYJS(ctx context.Context) error {
	db, err := s.oracleDB(ctx)
	if err != nil {
		return err
	}
	syncCtx, err := s.accountSyncContext(ctx, s.tenantID(ctx), legacyContainerStudentYJ, legacyContainerStudentYJ)
	if err != nil {
		return err
	}
	stats := jobRunStats{}
	for page := 0; ; page++ {
		var rows []*oracleStudentYJS
		if err := oraclePageModel(db, oracleTableYJS, page, defaultPageSize).Scan(&rows); err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		inputs := make([]*accountSyncInput, 0, len(rows))
		for _, row := range rows {
			inputs = append(inputs, studentYJSInput(row))
		}
		pageStats, err := s.syncOracleAccounts(ctx, syncCtx, inputs)
		if err != nil {
			return err
		}
		stats.add(pageStats)
		if len(rows) < defaultPageSize {
			break
		}
	}
	logger.Infof(ctx, "uidentity oracle graduate student sync finished stats=%v", sqlLogFields(stats))
	return nil
}

func (s *serviceImpl) syncOracleStudentWJ(ctx context.Context) error {
	db, err := s.oracleDB(ctx)
	if err != nil {
		return err
	}
	syncCtx, err := s.accountSyncContext(ctx, s.tenantID(ctx), legacyContainerStudentWJ, legacyContainerStudentWJ)
	if err != nil {
		return err
	}
	stats := jobRunStats{}
	for page := 0; ; page++ {
		var rows []*oracleStudentWJ
		if err := oraclePageModel(db, oracleTableWJ, page, defaultPageSize).Scan(&rows); err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		inputs := make([]*accountSyncInput, 0, len(rows))
		for _, row := range rows {
			inputs = append(inputs, studentWJInput(row))
		}
		pageStats, err := s.syncOracleAccounts(ctx, syncCtx, inputs)
		if err != nil {
			return err
		}
		stats.add(pageStats)
		if len(rows) < defaultPageSize {
			break
		}
	}
	logger.Infof(ctx, "uidentity oracle online student sync finished stats=%v", sqlLogFields(stats))
	return nil
}

func (s *serviceImpl) syncOracleStaff(ctx context.Context) error {
	db, err := s.oracleDB(ctx)
	if err != nil {
		return err
	}
	syncCtx, err := s.accountSyncContext(ctx, s.tenantID(ctx), legacyContainerStaff, legacyContainerStaff)
	if err != nil {
		return err
	}
	stats := jobRunStats{}
	for page := 0; ; page++ {
		var rows []*oracleStaffJZG
		if err := oraclePageModel(db, oracleTableJZG, page, defaultPageSize).Scan(&rows); err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		inputs := make([]*accountSyncInput, 0, len(rows))
		for _, row := range rows {
			inputs = append(inputs, staffJZGInput(row))
		}
		pageStats, err := s.syncOracleAccounts(ctx, syncCtx, inputs)
		if err != nil {
			return err
		}
		stats.add(pageStats)
		if len(rows) < defaultPageSize {
			break
		}
	}
	logger.Infof(ctx, "uidentity oracle staff sync finished stats=%v", sqlLogFields(stats))
	return nil
}

func (s *serviceImpl) oracleDB(ctx context.Context) (gdb.DB, error) {
	dsn, err := s.requireConfigString(ctx, configKeyOracleDSN)
	if err != nil {
		return nil, err
	}
	db, err := gdb.New(gdb.ConfigNode{Link: dsn})
	if err != nil {
		return nil, unsupportedOracleDriverError(err)
	}
	return db, nil
}

func (s *serviceImpl) studentInput(row *oracleStudentInfo) *accountSyncInput {
	if row == nil {
		return &accountSyncInput{}
	}
	return &accountSyncInput{
		number:                 row.XH,
		name:                   row.XM,
		phone:                  row.SJH,
		statusRaw:              row.ZDQK,
		unitCode:               row.XYDM,
		containerName:          legacyContainerStudent,
		groupName:              legacyContainerStudent,
		requireNumericUnitCode: true,
		detail: accountDetailSyncInput{
			idcard:       row.SFZHM,
			birthday:     row.CSRQ,
			avatar:       row.ZP,
			gender:       genderValue(row.XB),
			grade:        intString(row.NJ),
			schoolSystem: intString(row.XZ),
			graduatedAt:  intString(row.YJBYSJ),
			collegeCode:  row.XYDM,
			college:      row.XYMC,
			campus:       row.XQ,
			major:        row.ZYMC,
			className:    row.BJMC,
			source:       legacySourceSync,
		},
	}
}

func studentYJSInput(row *oracleStudentYJS) *accountSyncInput {
	if row == nil {
		return &accountSyncInput{}
	}
	return &accountSyncInput{
		number:                 row.Xh,
		name:                   row.XsXm,
		phone:                  row.Sjhm,
		statusRaw:              row.Sfzj,
		unitCode:               row.XsYxDm,
		containerName:          legacyContainerStudentYJ,
		groupName:              legacyContainerStudentYJ,
		requireNumericUnitCode: true,
		detail: accountDetailSyncInput{
			idcard:      row.Zjhm,
			birthday:    birthdayFromIDCard(row.Zjhm),
			gender:      genderValue(row.XbMc),
			grade:       intString(row.Nj),
			collegeCode: row.XsYxDm,
			college:     row.XsYxMc,
			major:       row.XsZyMc,
			campus:      row.Xq,
			source:      legacySourceSync,
		},
	}
}

func studentWJInput(row *oracleStudentWJ) *accountSyncInput {
	if row == nil {
		return &accountSyncInput{}
	}
	return &accountSyncInput{
		number:        row.XH,
		name:          row.XM,
		phone:         row.YDDH,
		statusRaw:     row.XJZT,
		unitCode:      "424",
		containerName: legacyContainerStudentWJ,
		groupName:     legacyContainerStudentWJ,
		detail: accountDetailSyncInput{
			idcard: row.SFZH,
			qq:     row.QQ,
			email:  row.Email,
			source: legacySourceSync,
		},
	}
}

func staffJZGInput(row *oracleStaffJZG) *accountSyncInput {
	if row == nil {
		return &accountSyncInput{}
	}
	return &accountSyncInput{
		number:                 row.GH,
		name:                   row.XM,
		phone:                  row.MobilePhone,
		statusRaw:              row.RYZTM,
		unitCode:               row.DWH,
		containerName:          legacyContainerStaff,
		groupName:              legacyContainerStaff,
		requireNumericUnitCode: true,
		detail: accountDetailSyncInput{
			idcard:   row.SFZJH,
			birthday: dateFromTime(row.CSRQ),
			email:    row.Email,
			gender:   genderValue(row.XBM),
			campus:   row.XQ,
			source:   legacySourceSync,
		},
	}
}

func (s *serviceImpl) upsertOracleUnits(ctx context.Context, tenantID int, rows []*oracleDept, stats *jobRunStats) error {
	codeSet := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row != nil {
			if code := normalizeText(row.DeptCode); code != "" {
				codeSet[code] = struct{}{}
			}
		}
	}
	existing, err := unitsByCodes(ctx, tenantID, stringsFromSet(codeSet))
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		code := normalizeText(row.DeptCode)
		name := normalizeText(row.DeptName)
		if code == "" || name == "" {
			stats.errNum++
			continue
		}
		unit := existing[code]
		if unit == nil {
			id, err := dao.Unit.Ctx(ctx).Data(do.Unit{
				TenantId:  tenantID,
				Name:      name,
				Code:      code,
				CreatedBy: int64(0),
				UpdatedBy: int64(0),
			}).InsertAndGetId()
			if err != nil {
				stats.errNum++
				logger.Warningf(ctx, "uidentity oracle dept create failed code=%s err=%v", code, err)
				continue
			}
			existing[code] = &entity.Unit{Id: id, TenantId: tenantID, Code: code, Name: name}
			stats.createNum++
			continue
		}
		if name != unit.Name {
			_, err := dao.Unit.Ctx(ctx).
				Where(dao.Unit.Columns().TenantId, tenantID).
				Where(dao.Unit.Columns().Id, unit.Id).
				Data(do.Unit{Name: name, UpdatedBy: int64(0)}).
				Update()
			if err != nil {
				stats.errNum++
				logger.Warningf(ctx, "uidentity oracle dept update failed code=%s err=%v", code, err)
				continue
			}
			unit.Name = name
			stats.updateNum++
		}
	}
	return nil
}

func unitsByCodes(ctx context.Context, tenantID int, codes []string) (map[string]*entity.Unit, error) {
	result := make(map[string]*entity.Unit, len(codes))
	if len(codes) == 0 {
		return result, nil
	}
	var units []*entity.Unit
	err := dao.Unit.Ctx(ctx).
		Where(dao.Unit.Columns().TenantId, tenantID).
		WhereIn(dao.Unit.Columns().Code, codes).
		Scan(&units)
	if err != nil {
		return nil, err
	}
	for _, unit := range units {
		if unit != nil {
			result[unit.Code] = unit
		}
	}
	return result, nil
}

func intString(value int) string {
	if value <= 0 {
		return ""
	}
	return strconv.Itoa(value)
}

func dateFromTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.DateOnly)
}

func (s *jobRunStats) add(other jobRunStats) {
	s.createNum += other.createNum
	s.updateNum += other.updateNum
	s.deleteNum += other.deleteNum
	s.errNum += other.errNum
	s.updateAccountCount += other.updateAccountCount
	s.updateAccountDetailCount += other.updateAccountDetailCount
}
