package service

import (
	json2 "encoding/json"
	"fmt"
	"time"

	model2 "github.com/IceWhaleTech/CasaOS/model"
	"github.com/IceWhaleTech/CasaOS/model/notify"
	"github.com/IceWhaleTech/CasaOS/service/model"
	"github.com/IceWhaleTech/CasaOS/types"
	"github.com/ambelovsky/gosf"
	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/mem"
	"gorm.io/gorm"
)

var NotifyMsg chan notify.Message
var ClientCount int = 0

type NotifyServer interface {
	GetLog(id string) model.AppNotify
	AddLog(log model.AppNotify)
	UpdateLog(log model.AppNotify)
	UpdateLogByCustomId(log model.AppNotify)
	DelLog(id string)
	GetList(c int) (list []model.AppNotify)
	MarkRead(id string, state int)
	//	SendText(m model.AppNotify)
	SendUninstallAppBySocket(app notify.Application)
	SendNetInfoBySocket(netList []model2.IOCountersStat)
	SendCPUInfoBySocket(cpu map[string]interface{})
	SendMemInfoBySocket(mem *mem.VirtualMemoryStat)
	SendUSBInfoBySocket(list []model2.DriveUSB)
	SendDiskInfoBySocket(disk model2.Summary)
	SendPersonStatusBySocket(status notify.Person)
	SendFileOperateNotify(nowSend bool)
	SendInstallAppBySocket(app notify.Application)
	SendAllHardwareStatusBySocket(disk model2.Summary, list []model2.DriveUSB, mem map[string]interface{}, cpu map[string]interface{}, netList []model2.IOCountersStat)
}

type notifyServer struct {
	db *gorm.DB
}

func (i *notifyServer) SendAllHardwareStatusBySocket(disk model2.Summary, list []model2.DriveUSB, mem map[string]interface{}, cpu map[string]interface{}, netList []model2.IOCountersStat) {

	body := make(map[string]interface{})
	body["sys_disk"] = disk

	body["sys_usb"] = list

	body["sys_mem"] = mem

	body["sys_cpu"] = cpu

	body["sys_net"] = netList

	msg := gosf.Message{}
	msg.Body = body
	msg.Success = true
	msg.Text = "sys_hardware_status"

	notify := notify.Message{}
	notify.Path = "sys_hardware_status"
	notify.Msg = msg

	NotifyMsg <- notify

}

// Send periodic broadcast messages
func (i *notifyServer) SendFileOperateNotify(nowSend bool) {

	if nowSend {

		len := 0
		FileQueue.Range(func(k, v interface{}) bool {
			len++
			return true
		})

		model := notify.NotifyModel{}
		listMsg := make(map[string]interface{})
		if len == 0 {
			model.Data = []string{}

			listMsg["file_operate"] = model
			msg := gosf.Message{}
			msg.Success = true
			msg.Body = listMsg
			msg.Text = "file_operate"

			notify := notify.Message{}
			notify.Path = "file_operate"
			notify.Msg = msg
			NotifyMsg <- notify
			return
		}

		model.State = "NORMAL"
		list := []notify.File{}
		OpStrArrbak := OpStrArr

		for _, v := range OpStrArrbak {
			tempItem, ok := FileQueue.Load(v)
			temp := tempItem.(model2.FileOperate)
			if !ok {
				continue
			}
			task := notify.File{}
			task.Id = v
			task.ProcessedSize = temp.ProcessedSize
			task.TotalSize = temp.TotalSize
			task.To = temp.To
			task.Type = temp.Type
			if task.ProcessedSize == 0 {
				task.Status = "STARTING"
			} else {
				task.Status = "PROCESSING"
			}

			if temp.Finished || temp.ProcessedSize >= temp.TotalSize {

				task.Finished = true
				task.Status = "FINISHED"
				FileQueue.Delete(v)
				OpStrArr = OpStrArr[1:]
				go ExecOpFile()
				list = append(list, task)
				continue
			}
			for _, v := range temp.Item {
				if v.Size != v.ProcessedSize {
					task.ProcessingPath = v.From
					break
				}
			}

			list = append(list, task)
		}
		model.Data = list

		listMsg["file_operate"] = model

		msg := gosf.Message{}
		msg.Success = true
		msg.Body = listMsg
		msg.Text = "file_operate"

		notify := notify.Message{}
		notify.Path = "file_operate"
		notify.Msg = msg
		NotifyMsg <- notify
	} else {
		for {

			len := 0
			FileQueue.Range(func(k, v interface{}) bool {
				len++
				return true
			})
			if len == 0 {
				return
			}
			listMsg := make(map[string]interface{})
			model := notify.NotifyModel{}
			model.State = "NORMAL"
			list := []notify.File{}
			OpStrArrbak := OpStrArr

			for _, v := range OpStrArrbak {
				tempItem, ok := FileQueue.Load(v)
				temp := tempItem.(model2.FileOperate)
				if !ok {
					continue
				}
				task := notify.File{}
				task.Id = v
				task.ProcessedSize = temp.ProcessedSize
				task.TotalSize = temp.TotalSize
				task.To = temp.To
				task.Type = temp.Type
				if task.ProcessedSize == 0 {
					task.Status = "STARTING"
				} else {
					task.Status = "PROCESSING"
				}
				if temp.Finished || temp.ProcessedSize >= temp.TotalSize {

					task.Finished = true
					task.Status = "FINISHED"
					FileQueue.Delete(v)
					OpStrArr = OpStrArr[1:]
					go ExecOpFile()
					list = append(list, task)
					continue
				}
				for _, v := range temp.Item {
					if v.Size != v.ProcessedSize {
						task.ProcessingPath = v.From
						break
					}
				}

				list = append(list, task)
			}
			model.Data = list

			listMsg["file_operate"] = model

			msg := gosf.Message{}
			msg.Success = true
			msg.Body = listMsg
			msg.Text = "file_operate"

			notify := notify.Message{}
			notify.Path = "file_operate"
			notify.Msg = msg
			NotifyMsg <- notify
			time.Sleep(time.Second * 3)
		}
	}

}

func (i *notifyServer) SendPersonStatusBySocket(status notify.Person) {
	body := make(map[string]interface{})
	body["data"] = status

	msg := gosf.Message{}
	msg.Body = body
	msg.Success = true
	msg.Text = "person_status"

	notify := notify.Message{}
	notify.Path = "person_status"
	notify.Msg = msg

	NotifyMsg <- notify
}

func (i *notifyServer) SendDiskInfoBySocket(disk model2.Summary) {
	body := make(map[string]interface{})
	body["data"] = disk

	msg := gosf.Message{}
	msg.Body = body
	msg.Success = true
	msg.Text = "sys_disk"

	notify := notify.Message{}
	notify.Path = "sys_disk"
	notify.Msg = msg

	NotifyMsg <- notify
}

func (i *notifyServer) SendUSBInfoBySocket(list []model2.DriveUSB) {
	body := make(map[string]interface{})
	body["data"] = list

	msg := gosf.Message{}
	msg.Body = body
	msg.Success = true
	msg.Text = "sys_usb"

	notify := notify.Message{}
	notify.Path = "sys_usb"
	notify.Msg = msg

	NotifyMsg <- notify
}

func (i *notifyServer) SendMemInfoBySocket(mem *mem.VirtualMemoryStat) {
	body := make(map[string]interface{})
	body["data"] = mem

	msg := gosf.Message{}
	msg.Body = body
	msg.Success = true
	msg.Text = "sys_mem"

	notify := notify.Message{}
	notify.Path = "sys_mem"
	notify.Msg = msg

	NotifyMsg <- notify
}

func (i *notifyServer) SendInstallAppBySocket(app notify.Application) {
	body := make(map[string]interface{})
	body["data"] = app

	msg := gosf.Message{}
	msg.Body = body
	msg.Success = true
	msg.Text = "app_install"

	notify := notify.Message{}
	notify.Path = "app_install"
	notify.Msg = msg

	NotifyMsg <- notify
}

func (i *notifyServer) SendCPUInfoBySocket(cpu map[string]interface{}) {
	body := make(map[string]interface{})
	body["data"] = cpu

	msg := gosf.Message{}
	msg.Body = body
	msg.Success = true
	msg.Text = "sys_cpu"

	notify := notify.Message{}
	notify.Path = "sys_cpu"
	notify.Msg = msg

	NotifyMsg <- notify
}
func (i *notifyServer) SendNetInfoBySocket(netList []model2.IOCountersStat) {
	body := make(map[string]interface{})
	body["data"] = netList

	msg := gosf.Message{}
	msg.Body = body
	msg.Success = true
	msg.Text = "sys_net"

	notify := notify.Message{}
	notify.Path = "sys_net"
	notify.Msg = msg

	NotifyMsg <- notify
}

func (i *notifyServer) SendUninstallAppBySocket(app notify.Application) {
	body := make(map[string]interface{})
	body["data"] = app

	msg := gosf.Message{}
	msg.Body = body
	msg.Success = true
	msg.Text = "app_uninstall"

	notify := notify.Message{}
	notify.Path = "app_uninstall"
	notify.Msg = msg

	NotifyMsg <- notify
}

func (i *notifyServer) SSR() {
	server := socketio.NewServer(nil)
	fmt.Println(server)
}

func (i notifyServer) GetList(c int) (list []model.AppNotify) {
	i.db.Where("class = ?", c).Where(i.db.Where("state = ?", types.NOTIFY_DYNAMICE).Or("state = ?", types.NOTIFY_UNREAD)).Find(&list)
	return
}

func (i *notifyServer) AddLog(log model.AppNotify) {
	i.db.Create(&log)
}

func (i *notifyServer) UpdateLog(log model.AppNotify) {
	i.db.Save(&log)
}
func (i *notifyServer) UpdateLogByCustomId(log model.AppNotify) {
	if len(log.CustomId) == 0 {
		return
	}
	i.db.Model(&model.AppNotify{}).Select("*").Where("custom_id = ? ", log.CustomId).Updates(log)
}
func (i *notifyServer) GetLog(id string) model.AppNotify {
	var log model.AppNotify
	i.db.Where("custom_id = ? ", id).First(&log)
	return log
}
func (i *notifyServer) MarkRead(id string, state int) {
	if id == "0" {
		i.db.Model(&model.AppNotify{}).Where("1 = ?", 1).Update("state", state)
		return
	}
	i.db.Model(&model.AppNotify{}).Where("id = ? ", id).Update("state", state)
}
func (i *notifyServer) DelLog(id string) {
	var log model.AppNotify
	i.db.Where("custom_id = ?", id).Delete(&log)
}

func SendMeg() {
	// for {
	// 	mt, message, err := ws.ReadMessage()
	// 	if err != nil {
	// 		break
	// 	}
	// 	notify := model.NotifyMssage{}
	// 	json2.Unmarshal(message, &notify)
	// 	if notify.Type == "read" {
	// 		service.MyService.Notify().MarkRead(notify.Data, types.NOTIFY_READ)
	// 	}
	// 	if notify.Type == "app" {
	//		go func(ws *websocket.Conn) {

	for {
		list := MyService.Notify().GetList(types.NOTIFY_APP)
		json, _ := json2.Marshal(list)

		if len(list) > 0 {
			var temp []*websocket.Conn
			for _, v := range WebSocketConns {

				err := v.WriteMessage(1, json)
				if err == nil {
					temp = append(temp, v)
				}
			}
			WebSocketConns = temp
			for _, v := range list {
				MyService.Notify().MarkRead(v.Id, types.NOTIFY_READ)
			}
		}

		if len(WebSocketConns) == 0 {
			SocketRun = false
		}
		time.Sleep(time.Second * 2)
	}
	// 	}(ws)
	// }
	//	}
}

// func (i notifyServer) SendText(m model.AppNotify) {
// 	list := []model.AppNotify{}
// 	list = append(list, m)
// 	json, _ := json2.Marshal(list)
// 	var temp []*websocket.Conn
// 	for _, v := range WebSocketConns {

// 		err := v.WriteMessage(1, json)
// 		if err == nil {
// 			temp = append(temp, v)
// 		}
// 	}
// 	WebSocketConns = temp

// 	if len(WebSocketConns) == 0 {
// 		SocketRun = false
// 	}

// }

func NewNotifyService(db *gorm.DB) NotifyServer {
	return &notifyServer{db: db}
}
