package pkg

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"demo.hello/utils"
)

//
// Task removes unhealth services from goc register list at fixed interval.
//

// ScheduleTaskRemoveUnhealthSrv .
func ScheduleTaskRemoveUnhealthSrv(ctx context.Context, interval time.Duration, host string) {
	go func() {
		tick := time.Tick(interval)
		for {
			select {
			case <-tick:
				func() {
					if err := removeUnhealthServicesFromGoc(host); err != nil {
						localCtx, cancel := context.WithTimeout(context.Background(), wait)
						defer cancel()
						errText := fmt.Sprintln("TaskRemoveUnhealthServices remove unhealth service failed:", err)
						notify.SendMessageToDefaultUser(localCtx, errText)
					}
				}()
			case <-ctx.Done():
				log.Println("Task exit: remove unhealth services from goc register list")
				return
			}
		}
	}()
}

// removeUnhealthServicesFromGoc removes unhealth services from goc register list.
func removeUnhealthServicesFromGoc(host string) error {
	goc := NewGocAPI(host)
	ctx, cancel := context.WithTimeout(context.Background(), shortWait)
	defer cancel()
	services, err := goc.ListRegisterServices(ctx)
	if err != nil {
		return fmt.Errorf("RemoveUnhealthServicesFromGocSvrList get goc register service list failed: %w", err)
	}

	for _, addrs := range services {
		for _, addr := range addrs {
			if !isAttachServerOK(addr) {
				if _, err := goc.DeleteRegisterServiceByAddr(ctx, addr); err != nil {
					return fmt.Errorf("RemoveUnhealthServicesFromGocSvrList remove goc register service [%s] failed: %w", addr, err)
				}
			}
		}
	}
	return nil
}

// isAttachServerOK checks wheather attached server is ok, and retry 3 times by default.
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
			log.Printf("IsAttachServerOK get service [%s] coverage failed: %s, retry %d", addr, err, i)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			return true
		}
	}
}

//
// Task sync service coverage/profile from goc center.
//

// SyncSrvCoverParam .
type SyncSrvCoverParam struct {
	Host    string
	SrvName string
	Address string
}

// ScheduleTaskSyncSrvCoverFromGoc .
func ScheduleTaskSyncSrvCoverFromGoc(param SyncSrvCoverParam, intervals []time.Duration) error {
	return nil
}

func createSrvCoverReport(param SyncSrvCoverParam) error {
	moduleName, err := getModuleFromSrvName(param.SrvName)
	if err != nil {
		return fmt.Errorf("createSrvCoverReport get module name failed: %w", err)
	}
	moduleDir := filepath.Join(WorkingRootDir, moduleName)
	if !utils.IsDirExist(moduleDir) {
		if err := utils.MakeDir(moduleDir); err != nil {
			return fmt.Errorf("createSrvCoverReport mkdir failed: %w", err)
		}
	}

	isUpdated, err := getSrvCoverProcess(param, moduleDir)
	if err != nil {
		return err
	}
	if !isUpdated {
		return nil
	}

	repoDir := filepath.Join(moduleDir, "repo")
	cmd := NewShCmd()
	if !utils.IsDirExist(repoDir) {
		uri := ModuleToRepoMap[moduleName]
		if err := cmd.CloneProject(uri, moduleDir); err != nil {
			return err
		}
	}

	if err := syncSrvRepo(param, repoDir); err != nil {
		return err
	}
	if err := cmd.GoToolCreateCoverFuncReport(repoDir); err != nil {
		return err
	}
	if err := cmd.GoToolCreateCoverHTMLReport(repoDir); err != nil {
		return err
	}
	return nil
}

func syncSrvRepo(param SyncSrvCoverParam, repoDir string) error {
	// TODO:
	return nil
}

// getSrvCoverProcess 1) get coverage from goc, and save to file; 2) move last cov file to history dir.
func getSrvCoverProcess(param SyncSrvCoverParam, moduleDir string) (bool, error) {
	coverFiles, err := utils.GetFilesBySuffix(moduleDir, ".cov")
	if err != nil {
		return false, fmt.Errorf("getSrvCoverProcess get exsting profile file failed: %w", err)
	}
	if len(coverFiles) > 1 {
		return false, fmt.Errorf("getSrvCoverProcess get more than one existing profile file from dir: %s", moduleDir)
	}

	savedPath, err := getSrvCoverAndSave(param, moduleDir)
	if err != nil {
		return false, fmt.Errorf("getSrvCoverProcess save service profile file failed: %w", err)
	}

	isCovUpdated := true
	if len(coverFiles) == 0 {
		return isCovUpdated, nil
	}
	lastCoverFilePath := coverFiles[0]
	lastCoverFileName := filepath.Base(lastCoverFilePath)
	historyCoverFilePath := filepath.Join(moduleDir, "history", lastCoverFileName)
	if err := utils.MoveFile(lastCoverFilePath, historyCoverFilePath); err != nil {
		return false, fmt.Errorf("getSrvCoverProcess move exiting profile file to history failed: %w", err)
	}

	isEqual, err := utils.IsFileContentEqual(historyCoverFilePath, savedPath)
	if err != nil {
		return false, fmt.Errorf("getSrvCoverProcess compare profile files content failed: %w", err)
	}
	if isEqual {
		isCovUpdated = false
	}
	return isCovUpdated, nil
}

func getSrvCoverAndSave(param SyncSrvCoverParam, savedDir string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), longWait)
	defer cancel()

	goc := NewGocAPI(param.Host)
	b, err := goc.GetServiceProfileByAddr(ctx, param.Address)
	if err != nil {
		return "", fmt.Errorf("saveSrvCoverFile get service [%s] profile failed: %w", param.Address, err)
	}

	fileName, err := getSavedCovFileName(param)
	if err != nil {
		return "", fmt.Errorf("saveSrvCoverFile get saved cov file name failed: %w", err)
	}

	savedPath := filepath.Join(savedDir, fileName)
	f, err := os.Create(savedPath)
	if err != nil {
		return "", fmt.Errorf("saveSrvCoverFile open file [%s] failed: %w", savedPath, err)
	}
	defer f.Close()

	buf := bytes.NewBuffer(b)
	if _, err := f.Write(buf.Bytes()); err != nil {
		return "", fmt.Errorf("saveSrvCoverFile write file [%s] failed: %w", savedPath, err)
	}
	return savedPath, nil
}

func getSavedCovFileName(param SyncSrvCoverParam) (string, error) {
	ip, err := getIPfromSrvAddress(param.Address)
	if err != nil {
		return "", err
	}
	now := getSimpleNowDatetime()
	return fmt.Sprintf("%s_%s_%s_%s.cov", param.SrvName, param.Address, ip, now), nil
}

//
// Task delete history profile files.
//

// TODO:
