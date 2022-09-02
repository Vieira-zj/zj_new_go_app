package googleapi

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

//
// Google Driver
//

var (
	gDriver     *GDriver
	gDriverOnce sync.Once
)

type GDriver struct {
	srv *drive.Service
}

func NewGDriver() *GDriver {
	gDriverOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		client := getAuthClient(drive.DriveFileScope)
		srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			log.Fatalf("Unable to retrieve Drive client: %v", err)
		}

		gDriver = &GDriver{
			srv: srv,
		}
	})

	return gDriver
}

func (gDriver *GDriver) ListFilesSample(ctx context.Context) error {
	resp, err := gDriver.srv.Files.List().PageSize(5).Fields("nextPageToken,files(id,name,permissions)").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("Unable to retrieve files: %v", err)
	}
	if len(resp.Files) == 0 {
		return fmt.Errorf("No files found")
	}

	for _, f := range resp.Files {
		fmt.Printf("name=%s,id=%s\n", f.Name, f.Id)
	}
	return nil
}

func (gDriver *GDriver) GetFilesSample(ctx context.Context, fileId string) error {
	file, err := gDriver.srv.Files.Get(fileId).Fields("id,name,permissions").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("Unable to retrieve files: %v", err)
	}

	fmt.Printf("name=[%s],id=%s\n", file.Name, file.Id)
	for _, p := range file.Permissions {
		fmt.Printf("\tid=%s,role=%s,type=%s\n", p.Id, p.Role, p.Type)
	}
	return nil
}

// AddEditPermissionForAnyone: refer https://developers.google.com/drive/api/guides/manage-sharing#python
func (gDriver *GDriver) AddEditPermissionForAnyone(ctx context.Context, fileId string) error {
	p := &drive.Permission{
		Role: "writer",
		Type: "anyone",
	}
	_, err := gDriver.srv.Permissions.Create(fileId, p).Context(ctx).Do()
	return err
}
