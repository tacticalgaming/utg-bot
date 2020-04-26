package main

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
	//"syscall"
)

type Mod struct {
	URL   string
	Title string
	Name  string
}

type Mods struct {
	config           *Config
	TableFile        string
	List             []Mod
	UpdateInProgress bool
	Users            []int // List of user IDs (from discord) who can update mods
	LastUpdateOutput []byte
	LastUpdateError  error
}

func (m *Mods) Init(config *Config) error {
	if config == nil {
		return fmt.Errorf("nil config")
	}

	m.config = config

	m.UpdateInProgress = false
	log.Infof("Initializing mods subsystem")
	m.TableFile = config.Directory + "/" + config.ScriptsDir + "/mods_table"

	file, err := os.Open(m.TableFile)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("Failed to open: %s", err.Error())
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		d := strings.Split(scanner.Text(), ";")
		if len(d) != 3 {
			continue
		}
		var nm Mod
		nm.URL = d[0]
		nm.Title = d[1]
		nm.Name = d[2]
		m.List = append(m.List, nm)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%s", err.Error())
	}

	log.Infof("%d mods loaded from file", len(m.List))

	return nil
}

func (m *Mods) handle(uid string, params []string) (string, error) {
	log.Infof("New command from %s. Params %+v", uid, params)
	if len(params) >= 1 {
		if params[0] == "help" {
			return getModsHelp(), nil
		} else if params[0] == "list" {
			if len(params) >= 2 {
				if params[1] == "links" {
					return m.listLinks(), nil
				} else if params[1] == "names" {
					return m.listNames(), nil
				} else {
					return "Непонятная команда", nil
				}
			}
			return m.listTitles(), nil
		} else if params[0] == "update" {
			if len(params) == 1 {
				if m.UpdateInProgress {
					return "Предыдущее обновление модов еще не завершилось", nil
				}

				if !m.validatePermission(uid) {
					return "Не достаточно прав для обновления модов", nil
				}

				m.UpdateInProgress = true
				go m.update()

				return "Запущена процедура обновления модов", nil
			} else if len(params) >= 2 {
				if params[1] == "status" {
					if m.UpdateInProgress {
						return "Обновление еще не завершено", nil
					}
					output := "Обновление завершено. Проверьте вывод на наличие ошибок:\n"
					if m.LastUpdateError != nil {
						output = "Обновить не удалось: " + m.LastUpdateError.Error() + "\n"
						return output, nil
					}
					output += string(m.LastUpdateOutput)
					return output, nil
				}
			}
		}
	}

	return getModsHelp(), nil
}

func (m *Mods) update() {
	m.LastUpdateError = nil
	m.LastUpdateOutput = m.LastUpdateOutput[:0]
	scriptPath := m.config.Directory + "/" + m.config.ScriptsDir + "/update_mods.txt"
	// cmd := exec.Command(m.config.SteamCmd, "+login", m.config.SteamUsername, "+runscript", scriptPath)
	cmd := exec.Command(m.config.SteamCmd)
	log.Infof("Executing: %s %s %s %s %s as %d:%d", m.config.SteamCmd, "+login", m.config.SteamUsername, "+runscript", scriptPath, m.config.Uid, m.config.Gid)
	/* cmd.SysProcAttr = &syscall.SysProcAttr{} */
	/* cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(m.config.Uid), Gid: uint32(m.config.Gid)} */
	m.LastUpdateOutput, m.LastUpdateError = cmd.Output()
	m.UpdateInProgress = false
	log.Infof("Output: %s", m.LastUpdateOutput)
	if m.LastUpdateError != nil {
		log.Errorf("Update completed with error: %s", m.LastUpdateError.Error())
		return
	}
	log.Infof("Update completed. No errors reported")
}

// returns false if user have no rights to update server
func (m *Mods) validatePermission(uid string) bool {
	for _, eid := range m.config.Admins {
		if eid == uid {
			return true
		}
	}
	for _, eid := range m.config.ModAdmins {
		if eid == uid {
			return true
		}
	}
	return false
}

func (m *Mods) listNames() string {
	buffer := "Названия модов на сервере:\n"
	i := 1
	for _, mod := range m.List {
		buffer += fmt.Sprintf("%d. %s\n", i, mod.Name)
		i++
	}
	return buffer
}

func (m *Mods) listLinks() string {
	buffer := "Страницы модов на сервере:\n"
	i := 1
	for _, mod := range m.List {
		buffer += fmt.Sprintf("%d. %s\n", i, mod.URL)
		i++
	}
	return buffer
}

func (m *Mods) listTitles() string {
	buffer := "Наименования модов на сервере:\n"
	i := 1
	for _, mod := range m.List {
		buffer += fmt.Sprintf("%d. %s\n", i, mod.Title)
		i++
	}
	return buffer
}

func getModsHelp() string {
	text := "Управление модами\n"
	text += "list - список модов (Названия)\n"
	text += "list links - список модов (Ссылки)\n"
	text += "list names - список модов (Имена на сервере, например @ace)\n"
	text += "update - запустить процедуру обновления модов\n"
	text += "update status - вернуть вывод процесса обновления\n"
	text += "Процесс обновления модов может занят от 3 до 20 минут. Зависит не известно от чего, но всегда по разному. Можно выполнять эту команду даже если никаких обновлений не было - это безопасно. Не забывайте перезапускать сервер после обновления.\n"
	return text
}
