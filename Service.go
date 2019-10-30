package main

import (
	"errors"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
)

//Target systemd target
type Target string

const (
	//NetworkTarget networktarget
	NetworkTarget Target = "network.target"
	//MultiuserTarget multiuser target
	MultiuserTarget Target = "multi-user.target"
	//SocketTarget socket target
	SocketTarget Target = "socket.target"
)

//ServiceType type of service
type ServiceType string

const (
	//Simple simple service
	Simple ServiceType = "simple"
	//Notify tells the systemd if its initialzed
	Notify ServiceType = "notify"
	//Forking keep active if a fork is running but the parent has exited
	Forking ServiceType = "forking"
	//Dbus a dbus service
	Dbus ServiceType = "dbus"
	//Oneshot wait until the start action has finished until it consideres to be active
	Oneshot ServiceType = "oneshot"
	//Exec similar to simple
	Exec ServiceType = "exec"
)

//Restart when the service should be restarted
type Restart string

const (
	//No don't restart
	No Restart = "no"
	//Always restart always
	Always Restart = "always"
	//OnSuccess restart only on success (exitcode=0 or on SIGHUP, SIGINT, SIGTERM or on SIGPIPE)
	OnSuccess Restart = "on-success"
	//OnFailure restart only on failure (exitcode != 0)
	OnFailure Restart = "on-failure"
	//OnAbnormal restart if the service was terminated by a signal, or an operation timed out
	OnAbnormal Restart = "on-abnormal"
	//OnAbort restart if the service was terminated by an non clean exit signal
	OnAbort Restart = "on-abort"
	//OnWatchdog restart if the watchdog timed out
	OnWatchdog = "on-watchdog"
)

//SystemdBool a bool (true=yes/false=no)
type SystemdBool string

const (
	//True true
	True SystemdBool = "yes"
	//False false
	False SystemdBool = "no"
)

//Service service
type Service struct {
	Name    string
	Unit    Unit
	Service SService
	Install Install
}

//Unit [Unit] in .service file
type Unit struct {
	Description         string `name:"Description"`
	Documentation       string `name:"Documentation"`
	Before              Target `name:"Before"`
	After               Target `name:"After"`
	Wants               Target `name:"Wants"`
	ConditionPathExisis string `name:"ConditionPathExists"`
	Conflicts           string `name:"Conflicts"`
}

//SService [Service] in .service file
type SService struct {
	Type                     ServiceType `name:"Type"`
	ExecStartPre             string      `name:"ExecStartPre"`
	ExecStart                string      `name:"ExecStart"`
	ExecReload               string      `name:"ExecReload"`
	ExecStop                 string      `name:"ExecStop"`
	RestartSec               string      `name:"RestartSec"`
	User                     string      `name:"User"`
	Group                    string      `name:"Group"`
	Restart                  Restart     `name:"Restart"`
	TimeoutStartSec          int         `name:"TimeoutStartSec"`
	TimeoutStopSec           int         `name:"TimeoutStopSec"`
	SuccessExitStatus        string      `name:"SuccessExitStatus"`
	RestartPreventExitStatus string      `name:"RestartPreventExitStatus"`
	PIDFile                  string      `name:"PIDFile"`
	WorkingDirectory         string      `name:"WorkingDirectory"`
	RootDirectory            string      `name:"RootDirectory"`
	LogsDirectory            string      `name:"LogsDirectory"`
	KillMode                 string      `name:"KillMode"`
	ConditionPathExists      string      `name:"ConditionPathExists"`
	RemainAfterExit          SystemdBool `name:"RemainAfterExit"`
}

//Install [Install] in .service file
type Install struct {
	WantedBy Target `name:"WantedBy"`
	Alias    string `name:"Alias"`
}

//NewDefaultService creates a new default service
func NewDefaultService(name, description, execStart string) *Service {
	return &Service{
		Name: name,
		Unit: Unit{
			Description: description,
			After:       NetworkTarget,
		},
		Service: SService{
			Type:      Simple,
			ExecStart: execStart,
		},
		Install: Install{
			WantedBy: MultiuserTarget,
		},
	}
}

//NewService creates a new service
func NewService(unit Unit, service SService, install Install) *Service {
	return &Service{
		Unit:    unit,
		Service: service,
		Install: install,
	}
}

//Stop stop
func (service *Service) Stop() {

}

//Start starts a service
func (service *Service) Start() {

}

//Disable disables a service
func (service *Service) Disable() error {
	return service.setStatus(0)
}

//Enable enables a service
func (service *Service) Enable() error {
	return service.setStatus(1)
}

//set the status of a service (0 = disabled;1=enabled)
func (service *Service) setStatus(newStatus int) error {
	name := nameToServiceFile(service.Name)
	_, err := os.Stat("/etc/systemd/system/" + name)
	if err != nil {
		return err
	}
	newMode := "enable"
	if newStatus == 0 {
		newMode = "disable"
	}
	_, err = runCommand(nil, "systemctl "+newMode+" "+name)
	return err
}

//Create creates a service file
func (service *Service) Create() error {
	if os.Getgid() != 0 {
		return errors.New("you neet to be root")
	}

	content := service.Generate()
	name := nameToServiceFile(service.Name)
	f, err := os.Create("/etc/systemd/system/" + name)
	if err != nil {
		return err
	}
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}

func nameToServiceFile(name string) string {
	if !strings.HasPrefix(name, ".service") {
		return name + ".service"
	}
	return name
}

func runCommand(errorHandler func(error, string), sCmd string) (outb string, err error) {
	out, err := exec.Command("su", "-c", sCmd).Output()
	output := string(out)
	if err != nil {
		if errorHandler != nil {
			errorHandler(err, sCmd)
		}
		return "", err
	}
	return output, nil
}

//Generate generates a service to a .service file
func (service *Service) Generate() string {
	unit := service.Unit
	sservice := service.Service
	install := service.Install
	final := ""
	var part interface{}
	for i := 0; i < 3; i++ {
		if i == 0 {
			part = &unit
		} else if i == 1 {
			part = &sservice
		} else if i == 2 {
			part = &install
		}
		if i == 0 {
			final += "[Unit]\n"
		} else if i == 1 {
			final += "\n[Service]\n"
		} else if i == 2 {
			final += "\n[Install]\n"
		}
		v := reflect.ValueOf(part).Elem()
		for index := 0; index < v.NumField(); index++ {
			value := v.Field(index)
			fieldKey := v.Type().Field(index).Tag.Get("name")

			if v.Field(index).Kind() == reflect.String && len(value.String()) > 0 {
				final += fieldKey + "=" + value.String() + "\n"
			} else if v.Field(index).Kind() == reflect.Int && value.Int() != 0 {
				final += fieldKey + "=" + strconv.FormatInt(value.Int(), 10) + "\n"
			}
		}
	}
	return final
}
