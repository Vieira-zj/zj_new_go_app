package pkg

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"demo.hello/utils"
	"gorm.io/gorm"
)

const (
	// ZeroCoverTotal .
	ZeroCoverTotal = "0.0"
	// CovFilePathNullValue .
	CovFilePathNullValue = "null"
)

//
// Task: list and remove unhealth services
//

// SyncAndListRegisterSrvsTask .
func SyncAndListRegisterSrvsTask() (map[string][]string, error) {
	if err := RemoveUnhealthSrvInGocTask(); err != nil {
		return nil, fmt.Errorf("SyncAndListRegisterSrvs error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), Wait)
	defer cancel()

	gocAPI := NewGocAPI()
	srvs, err := gocAPI.ListRegisterServices(ctx)
	if err != nil {
		return nil, fmt.Errorf("SyncAndListRegisterSrvs error: %w", err)
	}
	return srvs, nil
}

// RemoveUnhealthSrvInGocTask removes unhealth services from goc register list.
func RemoveUnhealthSrvInGocTask() error {
	goc := NewGocAPI()
	ctx, cancel := context.WithTimeout(context.Background(), ShortWait)
	defer cancel()
	services, err := goc.ListRegisterServices(ctx)
	if err != nil {
		return fmt.Errorf("RemoveUnhealthSrvInGoc error: %w", err)
	}

	for _, addrs := range services {
		for _, addr := range addrs {
			ok, err := isSrvOK(addr)
			if err != nil {
				return fmt.Errorf("RemoveUnhealthSrvInGoc error: %w", err)
			}
			if !ok {
				if _, err := goc.DeleteRegisterServiceByAddr(ctx, addr); err != nil {
					return fmt.Errorf("RemoveUnhealthSrvInGoc error: %w", err)
				}
				log.Printf("Remove unhealth service from goc list: srv_ip=%s", addr)
			}
		}
	}
	return nil
}

type podStatusReq struct {
	IPs []string `json:"ips"`
}

type podStatusResp struct {
	Total int `json:"total"`
	Data  []struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
		IP        string `json:"ip"`
		Value     string `json:"value"`
	}
}

func isSrvOK(addr string) (bool, error) {
	if AppConfig.Cluster == clusterK8s {
		return isPodOK(addr)
	}
	return isAttachSrvOK(addr)
}

// isPodOK checks pod status from pod monitor service.
func isPodOK(addr string) (bool, error) {
	addr = formatURLAddress(addr)
	req := podStatusReq{
		IPs: []string{addr},
	}
	body, err := json.Marshal(&req)
	if err != nil {
		return false, fmt.Errorf("isPodOK Marshal error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), Wait)
	defer cancel()

	client := utils.NewDefaultHTTPUtils()
	url := AppConfig.PodMonitorHost + "/list/pods/filter"
	resp, err := client.Post(ctx, url, map[string]string{}, string(body))
	podStatus := podStatusResp{}
	if err := json.Unmarshal(resp, &podStatus); err != nil {
		return false, fmt.Errorf("isPodOK Unmarshal error: %w", err)
	}

	if podStatus.Total == 0 {
		return false, nil
	}
	for _, status := range podStatus.Data {
		if strings.ToLower(status.Value) == "running" {
			return true, nil
		}
	}
	return false, nil
}

// formatURLAddress removes url schema and port from address.
func formatURLAddress(addr string) string {
	const httpSchema = "http://"
	if strings.HasPrefix(addr, httpSchema) {
		addr = strings.Replace(addr, httpSchema, "", 1)
	}

	if strings.Contains(addr, ":") {
		addr = strings.Split(addr, ":")[0]
	}
	return addr
}

// isAttachSrvOK checks wheather attached server is ok, and retry 4 times by default. (used in k8s cluster)
func isAttachSrvOK(addr string) (bool, error) {
	for i := 1; ; i++ {
		err := func() (err error) {
			ctx, cancel := context.WithTimeout(context.Background(), ShortWait)
			defer cancel()
			_, err = APIGetServiceCoverage(ctx, addr)
			return
		}()
		if err != nil {
			if i >= 4 {
				return false, err
			}
			log.Printf("IsAttachServerOK err: %s, retry %d", err, i)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			return true, nil
		}
	}
}

//
// Task: sync service cover and create report
//

// SyncSrvCoverParam .
type SyncSrvCoverParam struct {
	SrvName   string   `json:"srv_name" binding:"required"`
	Addresses []string `json:"addresses"`
}

// GetSrvCoverAndCreateReportTask .
func GetSrvCoverAndCreateReportTask(param SyncSrvCoverParam) (string, error) {
	covFile, isUpdate, err := getSrvCoverTask(param.SrvName)
	if err != nil {
		return "", fmt.Errorf("GetSrvCoverAndCreateReportTask error: %w", err)
	}

	meta := GetSrvMetaFromName(param.SrvName)
	if !isUpdate {
		return reuseLastCoverResults(covFile, meta)
	}

	coverTotal, err := createSrvCoverReportTask(covFile, param.SrvName)
	if err != nil {
		return "", fmt.Errorf("GetSrvCoverAndCreateReportTask error: %w", err)
	}

	meta.Addrs = strings.Join(param.Addresses, ",")
	newRow := GocSrvCoverModel{
		SrvCoverMeta: meta,
		IsLatest:     true,
		CovFilePath:  covFile,
		CoverTotal: sql.NullString{
			String: coverTotal,
			Valid:  true,
		},
	}
	if err := saveSrvCoverInDB(newRow); err != nil {
		return "", fmt.Errorf("GetSrvCoverAndCreateReportTask save db error: %w", err)
	}
	return coverTotal, nil
}

// reuseLastCoverResults: if coverage data is not changed, reuse the last results.
func reuseLastCoverResults(covFile string, meta SrvCoverMeta) (string, error) {
	dbInstance := NewGocSrvCoverDBInstance()
	transaction := dbInstance.BeginTransaction()
	row, err := updateCovFileOfLastSrvCoverRowInDB(transaction, covFile, meta)
	if err != nil {
		transaction.Rollback()
		return "", fmt.Errorf("reuseLastCoverResults error: %w", err)
	}

	oldCovFile := row.CovFilePath
	if err := renameLastSrvCoverResults(oldCovFile, covFile); err != nil {
		transaction.Rollback()
		return "", fmt.Errorf("reuseLastCoverResults error: %w", err)
	}
	if err := renameLastSrvHTMLCoverResults(meta.AppName, oldCovFile, covFile); err != nil {
		transaction.Rollback()
		return "", fmt.Errorf("reuseLastCoverResults error: %w", err)
	}

	if result := transaction.Commit(); result.Error != nil {
		return "", fmt.Errorf("reuseLastCoverResults commit error: %w", result.Error)
	}
	return row.CoverTotal.String, nil
}

// updateCovFileOfLastSrvCoverRowInDB updates cov file of latest cover row, and returns the old one.
func updateCovFileOfLastSrvCoverRowInDB(db *gorm.DB, covFilePath string, meta SrvCoverMeta) (GocSrvCoverModel, error) {
	dbInstance := NewGocSrvCoverDBInstance()
	row, err := dbInstance.GetLatestSrvCoverRowByDB(db, meta)
	if err != nil {
		return GocSrvCoverModel{}, fmt.Errorf("updateCovFileOfLastSrvCoverRowInDB error: %w", err)
	}

	err = dbInstance.UpdateCovFileOfLatestSrvCoverRowByDB(db, meta, covFilePath)
	return row, err
}

func renameLastSrvCoverResults(src, dst string) error {
	srcPath := getFilePathWithoutExt(src)
	dstPath := getFilePathWithoutExt(dst)
	for _, ext := range [1]string{".func"} {
		log.Println("rename:", srcPath+ext, dstPath+ext)
		if err := os.Rename(srcPath+ext, dstPath+ext); err != nil {
			return fmt.Errorf("renameLastSrvCoverResults error: %w", err)
		}
	}
	return nil
}

func renameLastSrvHTMLCoverResults(appName, src, dst string) error {
	const ext = ".html"
	srcName := filepath.Base(getFilePathWithoutExt(src)) + ext
	dstName := filepath.Base(getFilePathWithoutExt(dst)) + ext
	srcPath := filepath.Join(AppConfig.PublicDir, appName, srcName)
	dstPath := filepath.Join(AppConfig.PublicDir, appName, dstName)
	if err := os.Rename(srcPath, dstPath); err != nil {
		return fmt.Errorf("renameLastSrvHTMLCoverResults error: %w", err)
	}
	return nil
}

func saveSrvCoverInDB(row GocSrvCoverModel) error {
	dbInstance := NewGocSrvCoverDBInstance()
	if _, err := dbInstance.GetLatestSrvCoverRow(row.SrvCoverMeta); err != nil {
		if errors.Is(err, ErrSrvCoverLatestRowNotFound) {
			return dbInstance.InsertSrvCoverRow(row)
		}
		return fmt.Errorf("saveSrvCoverInDB get error: %w", err)
	}

	if err := dbInstance.AddLatestSrvCoverRow(row); err != nil {
		return fmt.Errorf("saveSrvCoverInDB save db error: %w", err)
	}
	return nil
}

//
// Task: create service cover report
//

// createSrvCoverReportTask
// 1. sync repo with specified commit;
// 2. run "go tool cover" to generate .func and .html coverage report.
func createSrvCoverReportTask(covFile, srvName string) (string, error) {
	meta := GetSrvMetaFromName(srvName)
	repoDir := filepath.Join(AppConfig.RootDir, meta.AppName, "repo")
	cmd := NewShCmd()
	if err := checkoutSrvRepo(repoDir, srvName); err != nil {
		return ZeroCoverTotal, fmt.Errorf("createSrvCoverReportTask error: %w", err)
	}

	coverTotal, err := cmd.GoToolCreateCoverFuncReport(repoDir, covFile)
	if err != nil {
		return ZeroCoverTotal, fmt.Errorf("createSrvCoverReportTask error: %w", err)
	}

	if _, err = cmd.GoToolCreateCoverHTMLReport(repoDir, meta.AppName, covFile); err != nil {
		return ZeroCoverTotal, fmt.Errorf("createSrvCoverReportTask error: %w", err)
	}
	return coverTotal, nil
}

func checkoutSrvRepo(workingDir, srvName string) error {
	meta := GetSrvMetaFromName(srvName)
	if !utils.IsDirExist(workingDir) {
		if err := func() error {
			repoURL, ok := ModuleToRepoMap[meta.AppName]
			if !ok {
				return fmt.Errorf("checkoutSrvRepo repo url not found: [%s]", meta.AppName)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			head, err := GitClone(ctx, repoURL, workingDir)
			if err != nil {
				return err
			}
			log.Printf("Git clone service [%s] with head [%s]", meta.AppName, head)
			return nil
		}(); err != nil {
			return fmt.Errorf("checkoutSrvRepo error: %w", err)
		}
	}

	repo := NewGitRepo(workingDir)
	IsExist, err := repo.IsBranchExist(meta.GitBranch)
	if err != nil {
		return fmt.Errorf("checkoutSrvRepo error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	head := ""
	if IsExist {
		head, err = repo.Pull(ctx, meta.GitBranch)
		if err != nil {
			return fmt.Errorf("checkoutSrvRepo error: %w", err)
		}
	} else {
		head, err = repo.CheckoutRemoteBranch(ctx, meta.GitBranch)
		if err != nil {
			return fmt.Errorf("checkoutSrvRepo error: %w", err)
		}
	}

	if head != meta.GitCommit {
		log.Printf("Checkout service [%s] with commit [%s]", meta.AppName, meta.GitCommit)
		if err := repo.CheckoutToCommit(meta.GitCommit); err != nil {
			return fmt.Errorf("checkoutSrvRepo error: %w", err)
		}
	}
	return nil
}

//
// Task: sync service coverage/profile
//

// getSrvCoverTask
// 1. fetch service coverage from goc, and save to file;
// 2. compare current coverage data with last one.
func getSrvCoverTask(srvName string) (string, bool, error) {
	moduleDir := GetSrvModuleDir(srvName)
	coverDir := filepath.Join(moduleDir, ReportCoverDataDirName)
	if !utils.IsDirExist(coverDir) {
		if err := utils.MakeDir(coverDir); err != nil {
			return "", false, fmt.Errorf("getSrvCoverTask make cover data dir error: %w", err)
		}
	}

	savedPath, err := FetchAndSaveSrvCover(coverDir, srvName)
	if err != nil {
		return "", false, fmt.Errorf("getSrvCoverTask error: %w", err)
	}

	isCovUpdated := true
	meta := GetSrvMetaFromName(srvName)
	dbInstance := NewGocSrvCoverDBInstance()
	row, err := dbInstance.GetLatestSrvCoverRow(meta)
	if err != nil {
		if errors.Is(err, ErrSrvCoverLatestRowNotFound) {
			return savedPath, isCovUpdated, nil
		}
		return "", false, fmt.Errorf("getSrvCoverTask error: %w", err)
	}
	if row.CovFilePath == CovFilePathNullValue {
		return savedPath, isCovUpdated, nil
	}

	isEqual, err := utils.IsFilesEqual(row.CovFilePath, savedPath)
	if err != nil {
		return "", false, fmt.Errorf("getSrvCoverTask compare cover files error: %w", err)
	}
	if isEqual {
		isCovUpdated = false
	}
	return savedPath, isCovUpdated, nil
}

// FetchAndSaveSrvCover .
func FetchAndSaveSrvCover(savedDir, srvName string) (string, error) {
	b, err := func() ([]byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), LongWait)
		defer cancel()
		goc := NewGocAPI()
		b, err := goc.GetServiceProfileByName(ctx, srvName)
		if err != nil {
			log.Println("FetchAndSaveSrvCover error:", err)
			return getSrvCoverFromWatcher()
		}
		return b, nil
	}()
	if err != nil {
		return "", fmt.Errorf("FetchAndSaveSrvCover get srv profile error: %w", err)
	}

	fileName := getSavedCovFileNameWithSuffix(srvName, "")
	savedPath := filepath.Join(savedDir, fileName)
	if err := utils.CreateFile(savedPath, b); err != nil {
		return "", fmt.Errorf("FetchAndSaveSrvCover error: %w", err)
	}
	return savedPath, nil
}

// FetchAndSaveSrvCoverByAddr get service cover by addresses and do merge.
func FetchAndSaveSrvCoverByAddr(savedDir string, param SyncSrvCoverParam) (string, error) {
	savedPaths := make([]string, 0, len(param.Addresses))
	goc := NewGocAPI()
	for idx, addr := range param.Addresses {
		b, err := func() ([]byte, error) {
			ctx, cancel := context.WithTimeout(context.Background(), LongWait)
			defer cancel()
			b, err := goc.GetServiceProfileByAddr(ctx, addr)
			if err != nil {
				log.Println("FetchAndSaveSrvCoverByAddr error: %w", err)
				return getSrvCoverFromWatcher()
			}
			return b, nil
		}()
		if err != nil {
			return "", fmt.Errorf("FetchAndSaveSrvCoverByAddr get srv profile error: %w", err)
		}

		fileName := getSavedCovFileNameWithSuffix(param.SrvName, strconv.Itoa(idx))
		savedPath := filepath.Join(savedDir, fileName)

		if err := utils.CreateFile(savedPath, b); err != nil {
			return "", fmt.Errorf("FetchAndSaveSrvCoverByAddr error: %w", err)
		}
		savedPaths = append(savedPaths, savedPath)
	}

	if len(savedPaths) == 1 {
		return savedPaths[0], nil
	}

	cmd := NewShCmd()
	mergeFileName := getSavedCovFileNameWithSuffix(param.SrvName, "merge")
	mergeFilePath := filepath.Join(savedDir, mergeFileName)
	if err := cmd.GocToolMergeSrvCovers(savedPaths, mergeFilePath); err != nil {
		return "", fmt.Errorf("FetchAndSaveSrvCoverByAddr merge srv cov files error: %w", err)
	}
	return mergeFilePath, nil
}

func getSrvCoverFromWatcher() ([]byte, error) {
	// TODO:
	return nil, nil
}

// GetCoverTotalFromAttachSrv gets service cover total from attached server directly.
func GetCoverTotalFromAttachSrv(addr string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ShortWait)
	defer cancel()

	cover, err := APIGetServiceCoverage(ctx, addr)
	if err != nil {
		return ZeroCoverTotal, fmt.Errorf("GetCoverTotalFromAttachSrv error: %w", err)
	}

	total, err := formatCoverPercentage(cover)
	if err != nil {
		return ZeroCoverTotal, fmt.Errorf("GetCoverTotalFromAttachSrv error: %w", err)
	}
	return total, nil
}

//
// Helper
//

// GetSrvModuleDir .
func GetSrvModuleDir(srvName string) string {
	meta := GetSrvMetaFromName(srvName)
	return filepath.Join(AppConfig.RootDir, meta.AppName)
}

// GetSrvMetaFromName .
func GetSrvMetaFromName(name string) SrvCoverMeta {
	// input name example:
	// staging_th_apa_goc_echoserver_master_518e0a570c
	items := strings.Split(name, "_")
	size := len(items)
	compItems := make([]string, 0, 4)
	for i := 2; i < size-2; i++ {
		compItems = append(compItems, items[i])
	}

	branch := items[size-2]
	if strings.Contains(branch, "/") {
		brItems := strings.Split(branch, "/")
		branch = brItems[len(brItems)-1]
	}

	return SrvCoverMeta{
		Env:       items[0],
		Region:    items[1],
		AppName:   strings.Join(compItems, "_"),
		GitBranch: branch,
		GitCommit: items[size-1],
	}
}

func getSavedCovFileNameWithSuffix(srvName, suffix string) string {
	now := getSimpleNowDatetime()
	if len(suffix) > 0 {
		return fmt.Sprintf("%s_%s_%s.cov", srvName, now, suffix)
	}
	return fmt.Sprintf("%s_%s.cov", srvName, now)
}
