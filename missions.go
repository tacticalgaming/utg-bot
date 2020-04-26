package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Missions struct {
	config *Config
}

func (m *Missions) Init(config *Config) error {
	if config == nil {
		return fmt.Errorf("nil config")
	}

	m.config = config
	log.Infof("Initializing missions subsystem")
	missions, err := m.list()
	if err != nil {
		return err
	}
	log.Infof("%d missions found on server", len(missions))

	return nil
}

// checkPermissions returns true if user allowed to upload
func (m *Missions) checkPermissions(uid string) bool {
	for _, u := range m.config.Uploaders {
		if u == uid {
			return true
		}
	}

	return false
}

func (m *Missions) list() ([]string, error) {
	path := m.directory()
	var result []string
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		result = append(result, f.Name())
	}
	return result, nil
}

func (m *Missions) checkFilename(filename string) bool {
	if strings.Index(filename, ".pbo") > 0 {
		return true
	}
	return false
}

func (m *Missions) handle(uid string, params []string, attachments []*discordgo.MessageAttachment) (string, error) {
	log.Infof("New command from %s. Params %+v", uid, params)
	if len(params) == 0 {
		return "", nil
	}

	if !m.checkPermissions(uid) {
		return "Тебе нельзя загружать миссии", nil
	}

	response := ""
	if params[0] == "upload" {
		if len(attachments) == 0 {
			return "Нет файлов для загрузки", nil
		}
		for _, attachment := range attachments {
			log.Infof("Attachment: %+v", attachment)
			if !m.checkFilename(attachment.Filename) {
				response += attachment.Filename + " не загружаю, потому что не PBO"
				continue
			}
			err := m.upload(attachment.URL, attachment.ProxyURL, attachment.Filename)
			if err != nil {
				response += attachment.Filename + " " + err.Error() + "\n"
			} else {
				response += attachment.Filename + " успешно загружена\n"
			}
		}
	}
	return response, nil
}

func (m *Missions) upload(url, proxy, filename string) error {
	missions, err := m.list()
	if err != nil {
		return err
	}
	for _, mission := range missions {
		if mission == filename {
			return fmt.Errorf("Такая миссия уже есть")
		}
	}
	log.Infof("Attempting download")
	uploaded, err := m.download(url)
	if err != nil {
		log.Errorf("Download failed. Attempting download using proxy")
		uploaded, err = m.download(proxy)
		if err != nil {
			return err
		}
	}

	if uploaded == "" {
		return fmt.Errorf("Неизвестная ошибка")
	}

	err = m.move(uploaded, filename)
	if err != nil {
		return fmt.Errorf("Не удалось перенести файл: %s", err.Error())
	}

	err = m.chown(m.directory() + "/" + filename)
	if err != nil {
		return fmt.Errorf("Не смог изменить владельца, но миссия загружена")
	}

	return nil
}

func (m *Missions) chown(path string) error {
	err := os.Chown(path, m.config.Uid, m.config.Gid)
	return err
}

func (m *Missions) move(uploaded, filename string) error {
	log.Infof("Moving %s to %s as %s", uploaded, m.directory(), filename)
	err := os.Rename(uploaded, m.directory()+"/"+filename)
	if err != nil {
		return err
	}
	return nil
}

func (m *Missions) download(url string) (string, error) {
	tokens := strings.Split(url, "/")
	fileName := "/tmp/" + tokens[len(tokens)-1]
	log.Infof("Downloading %s to %s", url, fileName)

	// TODO: check file existence first with io.IsExist
	output, err := os.Create(fileName)
	if err != nil {
		log.Errorf("Couldn't create tmp file: %s, %s", fileName, err.Error())
		return "", fmt.Errorf("Не смог создать tmp файл: %s %s", fileName, err.Error())
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		log.Errorf("Couldn't download file from %s: %s", url, err.Error())
		return "", fmt.Errorf("Не смог скачать файл %s: %s", url, err.Error())
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		log.Errorf("Download from %s failed: %s", url, err.Error())
		return "", fmt.Errorf("Во время загрузки произошла ошибка: %s", err.Error())
	}

	log.Infof("%d bytes downloaded", n)
	return fileName, nil
}

func (m *Missions) directory() string {
	return m.config.Directory + "/" + m.config.ServerDir + "/mpmissions"
}
