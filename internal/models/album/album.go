package album

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"springoff/internal/config"
	"springoff/internal/models/upload"
	"springoff/internal/util"
	"strings"
)

type Album struct {
	db *sql.DB
}

type Albums struct {
	IdAlbum    int
	TitleAlbum string
	NameAlbum  string
	PathCover  string
}

type Images struct {
	IdImage   int
	NameImage string
	NameAlbum string
}

type Response struct {
	Data    Data `json:"data"`
	Success bool `json:"success"`
	Status  int  `json:"status"`
}

type Data struct {
	Id   string `json:"id"`
	Link string `json:"link"`
}

type ErrorResponse struct {
	Error `json:"error"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func New(storage *sql.DB) *Album {
	return &Album{db: storage}
}

func (a *Album) GetAll() ([]Albums, error) {
	rows, err := a.db.Query("SELECT * FROM albums ORDER BY id_album DESC")
	if err != nil {
		return nil, err
	}

	var albums []Albums
	for rows.Next() {
		alb := Albums{}
		err := rows.Scan(&alb.IdAlbum, &alb.TitleAlbum, &alb.NameAlbum, &alb.PathCover)
		if err != nil {
			slog.Error("error mapping for object Album", "error", err)
			continue
		}
		albums = append(albums, alb)
	}

	return albums, nil
}

type ImageAlbum struct {
	Data []struct {
		Id   string `json:"id"`
		Link string `json:"link"`
	} `json:"data"`
}

type Image struct {
	Link string
}

func (a *Album) GetImages(id string, config *config.Config) ([]string, error) {
	rows, err := a.db.Query("SELECT * FROM images WHERE name_album=?", id)
	if err != nil {
		return nil, err
	}

	var imageDB []Images
	for rows.Next() {
		img := Images{}
		err := rows.Scan(&img.IdImage, &img.NameImage, &img.NameAlbum)
		if err != nil {
			slog.Error("Error mapping for object Images", "error", err)
			continue
		}
		imageDB = append(imageDB, img)
	}

	if imageDB == nil {
		return nil, fmt.Errorf("album not found")
	}

	agent := fiber.Get(fmt.Sprintf("https://api.imageban.ru/v1/album/%s/images", id))
	agent.Add("Authorization", config.SecretKey)
	code, body, _ := agent.String()
	if code != 200 {
		slog.Error("error get album", "code", code, "body", body)
		return nil, fmt.Errorf("error get album")
	}

	imageAPI := ImageAlbum{}
	if err := json.Unmarshal([]byte(body), &imageAPI); err != nil {
		return nil, err
	}

	var result []string
	done := true
	for done {
		done = false
		for i := 0; i < len(imageDB); i++ {
			if imageDB[i].NameImage == imageAPI.Data[i].Id {
				result = append(result, imageAPI.Data[i].Link)
				//result[i].Link = append(result, imageAPI.Data[i].Link)
				imageAPI.Data[i].Id = ""
				done = true
			}
		}
	}

	return result, nil
}

func (a *Album) GetTitle(id string) (string, error) {
	row := a.db.QueryRow("SELECT title_album FROM albums WHERE name_album=?", id)

	alb := Albums{}
	err := row.Scan(&alb.TitleAlbum)
	if err != nil {
		return "", err
	}
	return alb.TitleAlbum, nil
}

func (a *Album) Upload(upload *upload.Upload, config *config.Config) error {
	slog.Info("create agent")

	agent := fiber.Post("https://api.imageban.ru/v1/album")
	agent.Add("Authorization", config.SecretKey)

	slog.Info("create album")

	args := fiber.AcquireArgs()
	args.Set("album_name", upload.Title[0])
	agent.MultipartForm(args)
	code, body, _ := agent.String()

	if code != 200 {
		slog.Error("something went wrong", "code", code, "body", body)
		return fmt.Errorf("something went wrong when accessing the api")
	}

	slog.Info("handling an unsuccessful response")

	errorResponse := ErrorResponse{}
	if err := json.Unmarshal([]byte(body), &errorResponse); err != nil {
		slog.Error("error unmarshal response json")
		return err
	}
	if len(errorResponse.Message) > 0 {
		slog.Error("api returned an error", "code", errorResponse.Code, "message", errorResponse.Message)
		return fmt.Errorf("something went wrong in the making of the album")
	}

	slog.Info("processing a successful response")

	album := Response{}
	if err := json.Unmarshal([]byte(body), &album); err != nil {
		slog.Error("error unmarshal response json")
		return err
	}
	if album.Status != 200 {
		slog.Error("album not created", "success", album.Success, "status", album.Status)
		return fmt.Errorf("album not created. status=%s", album.Status)
	}

	slog.Info("album created", "id", album.Data.Id)

	slog.Info("adding cover image")

	cover, err := util.UploadFile(upload.Cover[0], config)
	if err != nil {
		return err
	}
	if cover.Status != 200 {
		slog.Error("Cover not added", "code", cover.Status, "body", cover.Data, "success", cover.Success)
		return err
	}

	slog.Info("adding image")

	pathImage := make([]string, len(upload.Album))
	for i := 0; i < len(upload.Album); i++ {
		albumResp, err := util.UploadFile(upload.Album[i], config)
		if err != nil {
			return err
		}
		if albumResp.Status != 200 {
			slog.Error("Album image not added", "code", albumResp.Status, "body", albumResp.Data, "success", albumResp.Success)
			return err
		}
		pathImage[i] = albumResp.Data.Id
	}

	slog.Info("transaction initialization")

	tx, err := a.db.Begin()

	_, err = tx.Exec("INSERT INTO albums(title_album, name_album, cover) VALUES (?,?,?)", upload.Title[0], album.Data.Id, cover.Data.Link)
	if err != nil {
		tx.Rollback()
		return err
	}

	for i := 0; i < len(pathImage); i++ {
		_, err = tx.Exec("INSERT INTO images(name_image, name_album) VALUES (?,?)", pathImage[i], album.Data.Id)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	slog.Info("adding image in album")
	slog.Info("create agent")

	agent = fiber.Put(fmt.Sprintf("https://api.imageban.ru/v1/album/%s", album.Data.Id))
	agent.Add("Authorization", config.SecretKey)

	args = fiber.AcquireArgs()
	args.Set("images", strings.Join(pathImage, ","))
	agent.MultipartForm(args)
	code, body, _ = agent.String()

	if code != 200 {
		slog.Error("error adding image in album", "code", code, "body", body)
		return fmt.Errorf("error adding image in album")
	}

	return nil
}

type DeleteResponse struct {
	Status int
}

func (a *Album) Delete(id string, config *config.Config) error {

	slog.Info("delete album", "id album", id)

	agent := fiber.Delete(fmt.Sprintf("https://api.imageban.ru/v1/album/%s", id))
	agent.Add("Authorization", config.SecretKey)
	code, body, _ := agent.String()
	if code != 200 {
		slog.Error("something went wrong", "code", code, "body", body)
		return fmt.Errorf("something went wrong when accessing the api")
	}

	resp := DeleteResponse{}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		return err
	}

	if resp.Status != 200 {
		slog.Error("album not deleted", "code", resp.Status)
		return fmt.Errorf("album not deleted, code %d", resp.Status)
	}

	resAlbum, err := a.db.Exec(`
	DELETE FROM albums WHERE name_album=?
	`, id)
	affectAlbum, err := resAlbum.RowsAffected()
	if affectAlbum == 0 {
		slog.Error("album deletion error", "id", id)
		return fmt.Errorf("the album hasn't been deleted")
	}
	if err != nil {
		slog.Error("error when executing a request to delete an album")
		return err
	}

	resImage, err := a.db.Exec(`
	DELETE FROM images WHERE name_album=?
	`, id)
	affectImage, err := resImage.RowsAffected()
	if affectImage == 0 {
		slog.Error("images deletion error", "id", id)
		return fmt.Errorf("the album hasn't been deleted")
	}
	if err != nil {
		slog.Error("error when executing a request to delete an images")
		return err
	}

	slog.Info("album successfully deleted from database")

	return nil
}

func (a *Album) Swap(id int, shift string) error {
	slog.Info("swapping album", "id", id, "shift", shift)

	var albumNearId int
	albumCurrent := new(Albums)
	albumNear := new(Albums)

	switch shift {
	case "left":
		albumNearId = id + 1
	case "right":
		albumNearId = id - 1
	default:
		slog.Error("unknown action", "id", id, "action", shift)
		return fmt.Errorf("unknown action: %s", shift)
	}

	slog.Info("Init query SELECT * FROM albums WHERE id_album=?", "id", id)
	rowCurrent := a.db.QueryRow("SELECT * FROM albums WHERE id_album=?", id)
	slog.Info("Init query SELECT * FROM albums WHERE id_album=?", "id", albumNearId)
	rowNear := a.db.QueryRow("SELECT * FROM albums WHERE id_album=?", albumNearId)

	if err := rowCurrent.Scan(&albumCurrent.IdAlbum, &albumCurrent.TitleAlbum, &albumCurrent.NameAlbum, &albumCurrent.PathCover); err != nil {
		slog.Error("error query row current album", "error", err)
		return err
	}
	err := rowNear.Scan(&albumNear.IdAlbum, &albumNear.TitleAlbum, &albumNear.NameAlbum, &albumNear.PathCover)
	//if they try to move the outermost albums
	if errors.Is(sql.ErrNoRows, err) {
		slog.Info("the shift goes beyond the scope of the album")
		return fmt.Errorf("the shift goes beyond the scope of the album")
	}
	if err != nil {
		slog.Error("error query row near album", "error", err)
		return err
	}

	tempTitle := albumCurrent.TitleAlbum
	tempName := albumCurrent.NameAlbum
	tempPath := albumCurrent.PathCover

	tx, err := a.db.Begin()
	_, err = tx.Exec("UPDATE albums SET title_album=?, name_album=?, cover=? WHERE id_album=?", albumNear.TitleAlbum, albumNear.NameAlbum, albumNear.PathCover, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("UPDATE albums SET title_album=?, name_album=?, cover=? WHERE id_album=?", tempTitle, tempName, tempPath, albumNearId)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
