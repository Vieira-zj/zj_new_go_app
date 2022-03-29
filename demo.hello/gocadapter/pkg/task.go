package pkg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"demo.hello/utils"
)

//
// Task for unhealth services
//

// removeUnhealthSrvInGocTask removes unhealth services from goc register list.
func removeUnhealthSrvInGocTask(host string) error {
	goc := NewGocAPI(host)
	ctx, cancel := context.WithTimeout(context.Background(), shortWait)
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
			ctx, cancel := context.WithTimeout(context.Background(), shortWait)
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
	GocHost string
	SrvName string
	Address string
}

func getSrvCoverAndCreateReportTask(moduleDir string, param SyncSrvCoverParam) error {
	covFile, isUpdate, err := getSrvCoverTask(moduleDir, param)
	if err != nil {
		return fmt.Errorf("getSrvCoverAndCreateReportTask error: %w", err)
	}
	fmt.Println(covFile, isUpdate)
	// TODO:
	return nil
}

func createSrvCoverReportTask(moduleDir, covFile string, param SyncSrvCoverParam) error {
	repoDir := filepath.Join(moduleDir, "repo")
	cmd := NewShCmd()
	if err := syncSrvRepo(repoDir, param); err != nil {
		return fmt.Errorf("createSrvCoverReportTask error: %w", err)
	}

	if _, err := cmd.GoToolCreateCoverFuncReport(repoDir, covFile); err != nil {
		return fmt.Errorf("createSrvCoverReportTask error: %w", err)
	}
	if _, err := cmd.GoToolCreateCoverHTMLReport(repoDir, covFile); err != nil {
		return fmt.Errorf("createSrvCoverReportTask error: %w", err)
	}
	return nil
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

func getSrvCoverTask(moduleDir string, param SyncSrvCoverParam) (string, bool, error) {
	// TODO:
	return "", false, nil
}

// getSrvCoverTaskDeprecated
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
	ctx, cancel := context.WithTimeout(context.Background(), longWait)
	defer cancel()

	goc := NewGocAPI(param.GocHost)
	b, err := goc.GetServiceProfileByAddr(ctx, param.Address)
	if err != nil {
		return "", fmt.Errorf("getAndSaveSrvCover get service [%s] profile error: %w", param.Address, err)
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
// Task delete history profile files
//

// TODO:

//
// Helper
//

func getSavedCovFileName(param SyncSrvCoverParam) (string, error) {
	ip, err := formatIPAddress(param.Address)
	if err != nil {
		return "", err
	}
	now := getSimpleNowDatetime()
	return fmt.Sprintf("%s_%s_%s.cov", param.SrvName, ip, now), nil
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
	srvMeta := getSrvMetaFromName(name)
	return srvMeta.Branch, srvMeta.Commit
}

func getSrvMetaFromName(name string) SrvCoverMeta {
	// name: staging_th_apa_goc_echoserver_master_518e0a570c
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
		Env:     items[0],
		Region:  items[1],
		AppName: strings.Join(compItems, "_"),
		Branch:  branch,
		Commit:  items[size-1],
	}
}
