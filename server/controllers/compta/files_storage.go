package compta

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	ca "registro/sql/compta"
	"registro/sql/files"
	"registro/utils"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type googleDriveAPI struct {
	KeyFile  string
	FolderId string // id of the main storage folder
}

func (ga googleDriveAPI) newService() (*drive.Service, error) {
	return drive.NewService(context.Background(), option.WithAuthCredentialsFile(option.ServiceAccount, ga.KeyFile), option.WithScopes(drive.DriveScope))
}

// check emit an API call to check for credentials/connection issues
func (ga googleDriveAPI) check() error {
	ser, err := ga.newService()
	if err != nil {
		return err
	}

	// validate the given folder early
	file, err := ser.Files.Get(ga.FolderId).Do()
	if err != nil {
		return err
	}
	const folderType = "application/vnd.google-apps.folder"
	if file.MimeType != folderType {
		return fmt.Errorf("Compta Google Drive folder is %s", file.MimeType)
	}
	return nil
}

func (ga googleDriveAPI) createFile(fileName string, content []byte) (ID, miniatureID string, _ error) {
	min, err := files.ComputeMiniature(fileName, bytes.NewReader(content))
	if err != nil {
		return "", "", err
	}

	ser, err := ga.newService()
	if err != nil {
		return "", "", err
	}
	created, err := ser.Files.Create(&drive.File{Name: fileName, Parents: []string{ga.FolderId}}).Media(bytes.NewReader(content)).Do()
	if err != nil {
		return "", "", err
	}
	miniature, err := ser.Files.Create(&drive.File{Name: "min_" + fileName, Parents: []string{ga.FolderId}}).Media(bytes.NewReader(min)).Do()
	if err != nil {
		return "", "", err
	}

	return created.Id, miniature.Id, nil
}

func (ga googleDriveAPI) remove(fileId, miniatureId string) error {
	ser, err := ga.newService()
	if err != nil {
		return err
	}

	err = ser.Files.Delete(fileId).Do()
	if err != nil {
		return err
	}

	err = ser.Files.Delete(miniatureId).Do()
	return err
}

func (ga googleDriveAPI) removeMany(files []string) error {
	ser, err := ga.newService()
	if err != nil {
		return err
	}

	for _, file := range files {
		err = ser.Files.Delete(file).Do()
		if err != nil {
			return err
		}
	}
	return nil
}

func (ga googleDriveAPI) download(fileId string) ([]byte, error) {
	ser, err := ga.newService()
	if err != nil {
		return nil, err
	}

	resp, err := ser.Files.Get(fileId).Download()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	return b, err
}

func (ga googleDriveAPI) downloadToZip(files []idAndName, metas ca.ComptaFiles, out *zip.Writer) error {
	ser, err := ga.newService()
	if err != nil {
		return err
	}

	for _, file := range files {
		meta := metas[file.id]

		ext := filepath.Ext(meta.Name)
		outName := file.newName + ext

		resp, err := ser.Files.Get(meta.DriveID).Download()
		if err != nil {
			return err
		}

		fileW, err := out.Create(outName)
		if err != nil {
			return err
		}
		_, err = io.Copy(fileW, resp.Body)
		if err != nil {
			return err
		}

		err = resp.Body.Close()
		if err != nil {
			return err
		}

		err = out.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}

type fileSystemAPI struct {
	// Folder is the root directory
	// where files are stored
	Folder string
}

func (ga fileSystemAPI) createFile(meta ca.ComptaFile, content []byte) error {
	miniature, err := files.ComputeMiniature(meta.Name, bytes.NewReader(content))
	if err != nil {
		return err
	}

	filePath, miniaturePath := ga.paths(meta.Id)

	err = os.WriteFile(filePath, content, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Enregistrement impossible: %w", err)
	}
	err = os.WriteFile(miniaturePath, miniature, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Enregistrement impossible: %w", err)
	}

	return nil
}

func (ga fileSystemAPI) paths(id ca.IdComptaFile) (filePath, miniaturePath string) {
	return filepath.Join(ga.Folder, fmt.Sprintf("file_%d", id)), filepath.Join(ga.Folder, fmt.Sprintf("file_%d_min", id))
}

func (ga fileSystemAPI) remove(id ca.IdComptaFile) error {
	filePath, miniaturePath := ga.paths(id)
	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	err = os.Remove(miniaturePath)
	return err
}

func (ga fileSystemAPI) removeMany(files ca.ComptaFiles) error {
	for file := range files {
		if err := ga.remove(file); err != nil {
			return err
		}
	}
	return nil
}

func (ga fileSystemAPI) download(id ca.IdComptaFile, isMiniature bool) ([]byte, error) {
	filePath, miniaturePath := ga.paths(id)

	pathToOpen := filePath
	if isMiniature {
		pathToOpen = miniaturePath
	}
	b, err := os.ReadFile(pathToOpen)

	return b, err
}

type idAndName struct {
	id      ca.IdComptaFile
	newName string // without extension
}

func (ga fileSystemAPI) downloadToZip(files []idAndName, metas ca.ComptaFiles, out *zip.Writer) error {
	for _, file := range files {
		meta := metas[file.id]

		ext := filepath.Ext(meta.Name)
		outName := file.newName + ext

		filePath, _ := ga.paths(file.id)
		f, err := os.Open(filePath)

		fileW, err := out.Create(outName)
		if err != nil {
			return err
		}
		_, err = io.Copy(fileW, f)
		if err != nil {
			return err
		}

		err = f.Close()
		if err != nil {
			return err
		}

		err = out.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}

type filesAPI struct {
	googleDrive googleDriveAPI
	fileSystem  fileSystemAPI
}

func (ga filesAPI) createFile(db ca.DB, mode ca.FileStatus, fileName string, fileContent io.Reader) (_ ca.ComptaFile, _ error) {
	content, err := io.ReadAll(fileContent)
	if err != nil {
		return ca.ComptaFile{}, fmt.Errorf("Erreur interne: %w", err)
	}

	meta, err := ca.ComptaFile{Name: fileName}.Insert(db)
	if err != nil {
		return ca.ComptaFile{}, utils.SQLError(err)
	}

	switch mode {
	case ca.FileSystem:
		err = ga.fileSystem.createFile(meta, content)
		return meta, err
	case ca.GoogleDrive:
		id, idMin, err := ga.googleDrive.createFile(fileName, content)
		if err != nil {
			return ca.ComptaFile{}, err
		}
		meta.DriveID, meta.DriveIDMiniature = id, idMin
		meta, err = meta.Update(db)
		if err != nil {
			return ca.ComptaFile{}, utils.SQLError(err)
		}
		return meta, nil
	case ca.Archived:
		return ca.ComptaFile{}, fmt.Errorf("Enregistrement impossible: dossier archivé.")
	default:
		panic("exhaustive switch")
	}
}

func (ga filesAPI) remove(db ca.DB, mode ca.FileStatus, id ca.IdComptaFile) error {
	meta, err := ca.DeleteComptaFileById(db, id)
	if err != nil {
		return utils.SQLError(err)
	}

	switch mode {
	case ca.FileSystem:
		return ga.fileSystem.remove(id)
	case ca.GoogleDrive:
		return ga.googleDrive.remove(meta.DriveID, meta.DriveIDMiniature)
	case ca.Archived:
		// nothing to do
		return nil
	default:
		panic("exhaustive switch")
	}
}

func (ga filesAPI) removeMany(db ca.DB, mode ca.FileStatus, files ca.ComptaFiles) error {
	_, err := ca.DeleteComptaFilesByIDs(db, files.IDs()...)
	if err != nil {
		return utils.SQLError(err)
	}

	switch mode {
	case ca.FileSystem:
		return ga.fileSystem.removeMany(files)
	case ca.GoogleDrive:
		var ids []string
		for _, file := range files {
			ids = append(ids, file.DriveID, file.DriveIDMiniature)
		}
		return ga.googleDrive.removeMany(ids)
	case ca.Archived:
		// nothing to do
		return nil
	default:
		panic("exhaustive switch")
	}
}

func (ga filesAPI) download(db ca.DB, mode ca.FileStatus, fileId ca.IdComptaFile, isMiniature bool) ([]byte, string, error) {
	meta, err := ca.SelectComptaFile(db, fileId)
	if err != nil {
		return nil, "", utils.SQLError(err)
	}

	var b []byte
	switch mode {
	case ca.FileSystem:
		b, err = ga.fileSystem.download(fileId, isMiniature)
	case ca.GoogleDrive:
		if isMiniature {
			b, err = ga.googleDrive.download(meta.DriveIDMiniature)
		} else {
			b, err = ga.googleDrive.download(meta.DriveID)
		}
	case ca.Archived:
		err = fmt.Errorf("Téléchargement impossible: Dossier archivé.")
	default:
		panic("exhaustive switch")
	}

	return b, meta.Name, err
}

func (ga filesAPI) downloadToZip(db ca.DB, mode ca.FileStatus, files []idAndName, out *zip.Writer) error {
	var ids []ca.IdComptaFile
	for _, file := range files {
		ids = append(ids, file.id)
	}
	metas, err := ca.SelectComptaFiles(db, ids...)
	if err != nil {
		return utils.SQLError(err)
	}

	switch mode {
	case ca.FileSystem:
		return ga.fileSystem.downloadToZip(files, metas, out)
	case ca.GoogleDrive:
		return ga.googleDrive.downloadToZip(files, metas, out)
	case ca.Archived:
		return fmt.Errorf("Téléchargement impossible: dossier archivé.")
	default:
		panic("exhaustive switch")
	}
}
