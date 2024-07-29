package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"springoff/internal/config"
	"springoff/internal/util"
	"springoff/internal/util/validate"
)

type Service struct {
	db *sql.DB
}

func New(storage *sql.DB) *Service {
	return &Service{db: storage}
}

type Services struct {
	Id          int
	Title       string
	Description string
	Price       int
	PathImage   string
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

func (s *Service) GetServices() (*[]Services, error) {
	slog.Info("Get service")

	rows, err := s.db.Query("SELECT * FROM service ORDER BY id_service DESC")
	if err != nil {
		return nil, err
	}

	var services []Services
	for rows.Next() {
		srv := Services{}
		err := rows.Scan(&srv.Id, &srv.Title, &srv.Description, &srv.Price, &srv.PathImage)
		if err != nil {
			slog.Error("error mapping for object Services", "error", err)
			continue
		}
		services = append(services, srv)
	}

	return &services, nil
}

func (s *Service) UploadService(form *multipart.Form, config *config.Config) error {

	title := form.Value["title"]
	description := form.Value["description"]
	price := form.Value["price"]
	image := form.File["image-service"]
	err := validate.Service(title, description, price, image)
	if err != nil {
		slog.Error("error validate form object", "message", err)
		return err
	}
	slog.Info("upload form", "title", title[0], "description", description[0], "price", price[0], "image", image[0].Filename)

	imgService, err := util.UploadFile(image[0], config)
	if err != nil {
		return err
	}
	if imgService.Status != 200 {
		slog.Error("Cover not added", "code", imgService.Status, "body", imgService.Data, "success", imgService.Success)
		return err
	}

	_, err = s.db.Exec("INSERT INTO service(title, description, price, path_image) VALUES (?,?,?,?)", title[0], description[0], price[0], imgService.Data.Link)
	if err != nil {
		slog.Error("error execute query", "error", err, "query", "INSERT INTO service(title, description, price, path_image) VALUES (?,?,?,?)")
		return err
	}

	return nil
}

func (s *Service) Delete(id string) error {

	slog.Info("delete service", "id album", id)

	resAlbum, err := s.db.Exec(`
	DELETE FROM service WHERE id_service=?
	`, id)
	affectAlbum, err := resAlbum.RowsAffected()
	if affectAlbum == 0 {
		slog.Error("service deletion error", "id", id)
		return fmt.Errorf("the service hasn't been deleted")
	}
	if err != nil {
		slog.Error("error when executing a request to delete an service")
		return err
	}

	slog.Info("service successfully deleted from database")

	return nil
}

func (s *Service) Swap(id int, shift string) error {
	slog.Info("swapping album", "id", id, "shift", shift)

	var serviceNearId int
	serviceCurrent := new(Services)
	serviceNear := new(Services)

	switch shift {
	case "left":
		serviceNearId = id + 1
	case "right":
		serviceNearId = id - 1
	default:
		slog.Error("unknown action", "id", id, "action", shift)
		return fmt.Errorf("unknown action: %s", shift)
	}

	slog.Info("Init query SELECT * FROM service WHERE id_service=?", "id", id)
	rowCurrent := s.db.QueryRow("SELECT * FROM service WHERE id_service=?", id)
	slog.Info("Init query SELECT * FROM service WHERE id_service=?", "id", serviceNearId)
	rowNear := s.db.QueryRow("SELECT * FROM service WHERE id_service=?", serviceNearId)

	if err := rowCurrent.Scan(&serviceCurrent.Id, &serviceCurrent.Title, &serviceCurrent.Description, &serviceCurrent.Price, &serviceCurrent.PathImage); err != nil {
		slog.Error("error query row current album", "error", err)
		return err
	}
	err := rowNear.Scan(&serviceNear.Id, &serviceNear.Title, &serviceNear.Description, &serviceNear.Price, &serviceNear.PathImage)
	//if they try to move the outermost service
	if errors.Is(sql.ErrNoRows, err) {
		slog.Info("the shift goes beyond the scope of the service")
		return fmt.Errorf("the shift goes beyond the scope of the service")
	}
	if err != nil {
		slog.Error("error query row near service", "error", err)
		return err
	}

	tempTitle := serviceCurrent.Title
	tempDescription := serviceCurrent.Description
	tempPrice := serviceCurrent.Price
	tempPathImage := serviceCurrent.PathImage

	tx, err := s.db.Begin()
	_, err = tx.Exec("UPDATE service SET title=?, description=?, price=?, path_image=? WHERE id_service=?", serviceNear.Title, serviceNear.Description, serviceNear.Price, serviceNear.PathImage, id)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = tx.Exec("UPDATE service SET title=?, description=?, price=?, path_image=? WHERE id_service=?", tempTitle, tempDescription, tempPrice, tempPathImage, serviceNearId)
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
