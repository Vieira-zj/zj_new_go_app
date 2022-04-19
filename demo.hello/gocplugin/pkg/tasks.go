package pkg

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"demo.hello/utils"
	"gorm.io/gorm"
)

const defaultCover = "0"

//
// Task for unhealth services
//

// removeUnhealthSrvInGocTask removes unhealth services from goc register list.
func removeUnhealthSrvInGocTask() error {
	goc := NewGocAPI()
	ctx, cancel := context.WithTimeout(context.Background(), ShortWait)
	defer cancel()
	services, err := goc.ListRegisterServices(ctx)
	if err != nil {
		return fmt.Errorf("RemoveUnhealthServicesFromGocSvrList error: %w", err)
	}

	for _, addrs := range services {
		for _, addr := range addrs {
			if !isAttachServerOK(addr) {
				if _, err := goc.DeleteRegisterServiceByAddr(ctx, addr); err != nil {
					return fmt.Errorf("RemoveUnhealthServicesFromGocSvrList error: %w", err)
				}
			}
		}
	}
	return nil
}

// isAttachServerOK checks wheather attached server is ok, and retry 4 times by default.
func isAttachServerOK(addr string) bool {
	for i := 1; ; i++ {
		err := func() (err error) {
			ctx, cancel := context.WithTimeout(context.Background(), ShortWait)
			defer cancel()
			_, err = APIGetServiceCoverage(ctx, addr)
			return
		}()
		if err != nil {
			if i >= 4 {
				return false
			}
			log.Printf("IsAttachServerOK err: %s, retry %d", err, i)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			return true
		}
	}
}

//
// Task create cover report
//

// SyncSrvCoverParam .
type SyncSrvCoverParam struct {
	SrvName   string   `json:"srv_name"`
	Addresses []string `json:"addresses"`
}

// GetSrvCoverAndCreateReportTask .
func GetSrvCoverAndCreateReportTask(param SyncSrvCoverParam) (string, error) {
	covFile, isUpdate, err := getSrvCoverTask(param)
	if err != nil {
		return "", fmt.Errorf("getSrvCoverAndCreateReportTask error: %w", err)
	}

	meta := GetSrvMetaFromName(param.SrvName)
	if !isUpdate {
		return reuseLastCoverResults(covFile, meta)
	}

	coverTotal, err := createSrvCoverReportTask(covFile, param)
	if err != nil {
		return "", fmt.Errorf("getSrvCoverAndCreateReportTask error: %w", err)
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
		return "", fmt.Errorf("getSrvCoverAndCreateReportTask save db error: %w", err)
	}
	return coverTotal, nil
}

// reuseLastCoverResults: if coverage data is not changed, reuse the last results
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
		return GocSrvCoverModel{}, fmt.Errorf("updateCovFileOfLastSrvCoverRowInDB get error: %w", err)
	}

	return row, dbInstance.UpdateCovFileOfLatestSrvCoverRowByDB(db, meta, covFilePath)
}

func renameLastSrvCoverResults(src, dst string) error {
	srcName := getFilePathWithoutExt(src)
	dstName := getFilePathWithoutExt(dst)
	for _, ext := range []string{".func", ".html"} {
		fmt.Println("rename:", srcName+ext, dstName+ext)
		if err := os.Rename(srcName+ext, dstName+ext); err != nil {
			return fmt.Errorf("copyLastSrvCoverResults error: %w", err)
		}
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
		return fmt.Errorf("createSrvCoverReportTask save db error: %w", err)
	}
	return nil
}

// createSrvCoverReportTask
// 1. sync repo with specified commit;
// 2. run "go tool cover" to generate .func and .html coverage report.
func createSrvCoverReportTask(covFile string, param SyncSrvCoverParam) (string, error) {
	moduleDir := getModuleDir(param.SrvName)
	repoDir := filepath.Join(moduleDir, "repo")
	cmd := NewShCmd()
	if err := syncSrvRepo(repoDir, param); err != nil {
		return defaultCover, fmt.Errorf("createSrvCoverReportTask error: %w", err)
	}

	coverTotal, err := cmd.GoToolCreateCoverFuncReport(repoDir, covFile)
	if err != nil {
		return defaultCover, fmt.Errorf("createSrvCoverReportTask error: %w", err)
	}

	if _, err = cmd.GoToolCreateCoverHTMLReport(repoDir, covFile); err != nil {
		return defaultCover, fmt.Errorf("createSrvCoverReportTask error: %w", err)
	}
	return coverTotal, nil
}

func syncSrvRepo(workingDir string, param SyncSrvCoverParam) error {
	if !utils.IsDirExist(workingDir) {
		if err := func() error {
			url, err := getRepoURLFromSrvName(param.SrvName)
			if err != nil {
				return fmt.Errorf("syncSrvRepo get repo url error: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			head, err := GitClone(ctx, url, workingDir)
			if err != nil {
				return err
			}
			log.Printf("Git clone repo [%s] with head [%s]", param.SrvName, head)
			return nil
		}(); err != nil {
			return fmt.Errorf("syncSrvRepo error: %w", err)
		}
		return nil
	}

	repo := NewGitRepo(workingDir)
	branch, commitID := getBranchAndCommitFromSrvName(param.SrvName)
	IsExist, err := repo.IsBranchExist(branch)
	if err != nil {
		return fmt.Errorf("syncSrvRepo error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	head := ""
	if IsExist {
		head, err = repo.Pull(ctx, branch)
		if err != nil {
			return fmt.Errorf("syncSrvRepo error: %w", err)
		}
	} else {
		head, err = repo.CheckoutRemoteBranch(ctx, branch)
		if err != nil {
			return fmt.Errorf("syncSrvRepo error: %w", err)
		}
	}

	if head != commitID {
		if err := repo.CheckoutToCommit(commitID); err != nil {
			return fmt.Errorf("syncSrvRepo error: %w", err)
		}
	}
	return nil
}

//
// Task sync service coverage/profile
//

// getSrvCoverTask
// 1. get service coverage from goc, and save to file;
// 2. compare current coverage data with last one.
func getSrvCoverTask(param SyncSrvCoverParam) (string, bool, error) {
	moduleDir := getModuleDir(param.SrvName)
	coverDir := filepath.Join(moduleDir, "cover_data")
	if !utils.IsDirExist(coverDir) {
		if err := utils.MakeDir(coverDir); err != nil {
			return "", false, fmt.Errorf("getSrvCoverTask create cover dir error: %w", err)
		}
	}

	savedPath, err := getAndSaveSrvCover(coverDir, param)
	if err != nil {
		return "", false, fmt.Errorf("getSrvCoverTask error: %w", err)
	}

	isCovUpdated := true
	meta := GetSrvMetaFromName(param.SrvName)
	dbInstance := NewGocSrvCoverDBInstance()
	row, err := dbInstance.GetLatestSrvCoverRow(meta)
	if err != nil {
		if errors.Is(err, ErrSrvCoverLatestRowNotFound) {
			return savedPath, isCovUpdated, nil
		}
		return "", false, fmt.Errorf("getSrvCoverTask error: %w", err)
	}

	isEqual, err := utils.IsFileContentEqual(row.CovFilePath, savedPath)
	if err != nil {
		return "", false, fmt.Errorf("getSrvCoverTask compare cover files content error: %w", err)
	}
	if isEqual {
		isCovUpdated = false
	}
	return savedPath, isCovUpdated, nil
}

// getSrvCoverTask (Deprecated)
// 1. get service coverage from goc, and save to file;
// 2. move last cov file to history dir.
func getSrvCoverTaskDeprecated(moduleDir string, param SyncSrvCoverParam) (string, bool, error) {
	coverFiles, err := utils.GetFilesBySuffix(moduleDir, ".cov")
	if err != nil {
		return "", false, fmt.Errorf("getSrvCoverTask get exsting profile file error: %w", err)
	}
	if len(coverFiles) > 1 {
		return "", false, fmt.Errorf("getSrvCoverTask get more than one existing cover file from dir: %s", moduleDir)
	}

	savedPath, err := getAndSaveSrvCover(moduleDir, param)
	if err != nil {
		return "", false, fmt.Errorf("getSrvCoverTask error: %w", err)
	}

	isCovUpdated := true
	if len(coverFiles) == 0 {
		return savedPath, isCovUpdated, nil
	}

	historyDir := filepath.Join(moduleDir, "history")
	if !utils.IsDirExist(historyDir) {
		if err := utils.MakeDir(historyDir); err != nil {
			return "", false, fmt.Errorf("getSrvCoverTask create history dir error: %w", err)
		}
	}

	lastCoverFilePath := coverFiles[0]
	lastCoverFileName := filepath.Base(lastCoverFilePath)
	historyCoverFilePath := filepath.Join(historyDir, lastCoverFileName)
	if err := utils.MoveFile(lastCoverFilePath, historyCoverFilePath); err != nil {
		return "", false, fmt.Errorf("getSrvCoverTask move exiting cover file to history error: %w", err)
	}

	isEqual, err := utils.IsFileContentEqual(historyCoverFilePath, savedPath)
	if err != nil {
		return "", false, fmt.Errorf("getSrvCoverTask compare cover files content error: %w", err)
	}
	if isEqual {
		isCovUpdated = false
	}
	return savedPath, isCovUpdated, nil
}

func getAndSaveSrvCover(savedDir string, param SyncSrvCoverParam) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), LongWait)
	defer cancel()

	// TODO: merge cover for addresses
	goc := NewGocAPI()
	b, err := goc.GetServiceProfileByAddr(ctx, param.Addresses[0])
	if err != nil {
		return "", fmt.Errorf("getAndSaveSrvCover get service [%s] profile error: %w", param.Addresses[0], err)
	}

	fileName, err := getSavedCovFileName(param)
	if err != nil {
		return "", fmt.Errorf("getAndSaveSrvCover get saved cov file name error: %w", err)
	}

	savedPath := filepath.Join(savedDir, fileName)
	f, err := os.Create(savedPath)
	if err != nil {
		return "", fmt.Errorf("getAndSaveSrvCover open file [%s] error: %w", savedPath, err)
	}
	defer f.Close()

	buf := bytes.NewBuffer(b)
	if _, err := f.Write(buf.Bytes()); err != nil {
		return "", fmt.Errorf("getAndSaveSrvCover write file [%s] error: %w", savedPath, err)
	}
	return savedPath, nil
}

//
// Task delete history profile files/db rows
// TODO:
//

//
// Helper
//

func getSavedCovFileName(param SyncSrvCoverParam) (string, error) {
	now := getSimpleNowDatetime()
	return fmt.Sprintf("%s_%s.cov", param.SrvName, now), nil
}

func getModuleFromSrvName(name string) (string, error) {
	for mod := range ModuleToRepoMap {
		if strings.Contains(name, mod) {
			return mod, nil
		}
	}
	return "", fmt.Errorf("Module is not found for service: %s", name)
}

// ErrURLNotFound .
var ErrURLNotFound = errors.New("Repo URL not found")

func getRepoURLFromSrvName(name string) (string, error) {
	for k, v := range ModuleToRepoMap {
		if strings.Contains(name, k) {
			return v, nil
		}
	}
	return "", ErrURLNotFound
}

func getBranchAndCommitFromSrvName(name string) (string, string) {
	srvMeta := GetSrvMetaFromName(name)
	return srvMeta.GitBranch, srvMeta.GitCommit
}

func getModuleDir(srvName string) string {
	meta := GetSrvMetaFromName(srvName)
	return filepath.Join(AppConfig.RootDir, meta.AppName)
}

// GetSrvMetaFromName .
func GetSrvMetaFromName(name string) SrvCoverMeta {
	// example: staging_th_apa_goc_echoserver_master_518e0a570c
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

// GetSrvTotalFromGoc .
func GetSrvTotalFromGoc(addr string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ShortWait)
	defer cancel()
	cover, err := APIGetServiceCoverage(ctx, addr)
	if err != nil {
		return defaultCover, fmt.Errorf("GetSrvTotalFromGoc error: %w", err)
	}

	total, err := formatCoverPercentage(cover)
	if err != nil {
		return defaultCover, fmt.Errorf("GetSrvTotalFromGoc error: %w", err)
	}
	return total, nil
}
