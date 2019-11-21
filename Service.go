package SystemdGoService

import (
	"bufio"
	"errors"
	"fmt"
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

//ServiceRestart when the service should be restarted
type ServiceRestart string

const (
	//No don't restart
	No ServiceRestart = "no"
	//Always restart always
	Always ServiceRestart = "always"
	//OnSuccess restart only on success (exitcode=0 or on SIGHUP, SIGINT, SIGTERM or on SIGPIPE)
	OnSuccess ServiceRestart = "on-success"
	//OnFailure restart only on failure (exitcode != 0)
	OnFailure ServiceRestart = "on-failure"
	//OnAbnormal restart if the service was terminated by a signal, or an operation timed out
	OnAbnormal ServiceRestart = "on-abnormal"
	//OnAbort restart if the service was terminated by an non clean exit signal
	OnAbort ServiceRestart = "on-abort"
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
	Name    string   `name:"name"`
	Unit    Unit     `name:"[Unit]"`
	Service SService `name:"[Service]"`
	Install Install  `name:"[Install]"`
}

//Unit [Unit] in .service file
type Unit struct {
	Description         string
	Documentation       string
	Before              Target
	After               Target
	Wants               Target
	ConditionPathExists string
	Conflicts           string
}

//SService [Service] in .service file
type SService struct {
	Type                     ServiceType
	ExecStartPre             string
	ExecStart                string
	ExecReload               string
	ExecStop                 string
	RestartSec               string
	User                     string
	Group                    string
	Restart                  ServiceRestart
	TimeoutStartSec          int
	TimeoutStopSec           int
	SuccessExitStatus        string
	RestartPreventExitStatus string
	PIDFile                  string
	WorkingDirectory         string
	RootDirectory            string
	EnvironmentFile          string
	RuntimeDirectory         string
	RuntimeDirectoryMode     string
	LogsDirectory            string
	KillMode                 string
	ConditionPathExists      string
	RemainAfterExit          SystemdBool
}

//Install [Install] in .service file
type Install struct {
	WantedBy Target
	Alias    string
	Also     string
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
func (service *Service) Stop() error {
	return service.setStatus(Stop)
}

//Start starts a service
func (service *Service) Start() error {
	return service.setStatus(Start)
}

//Disable disables a service
func (service *Service) Disable() error {
	return service.setStatus(Disable)
}

//Enable enables a service
func (service *Service) Enable() error {
	return service.setStatus(Enable)
}

//SystemdCommand a command for systemd
type SystemdCommand int

const (
	//Stop stops a running service
	Stop SystemdCommand = 0
	//Start starts a stopped service
	Start SystemdCommand = 1
	//Enable enables a service to auto start
	Enable SystemdCommand = 2
	//Disable disables a service to auto start
	Disable SystemdCommand = 3
	//Restart restarts a service
	Restart SystemdCommand = 4
)

//SetServiceStatus sets new status for service
func SetServiceStatus(service string, newStatus SystemdCommand) error {
	name := NameToServiceFile(service)
	if !SystemfileExists(name) {
		return errors.New("service not found")
	}
	newMode := ""
	switch newStatus {
	case Stop:
		{
			newMode = "stop"
		}
	case Start:
		{
			newMode = "start"
		}
	case Enable:
		{
			newMode = "enable"
		}
	case Disable:
		{
			newMode = "disable"
		}
	case Restart:
		{
			newMode = "restart"
		}
	default:
		{
			return errors.New("no matching command available")
		}
	}
	_, err := runCommand(nil, "systemctl "+newMode+" "+name)
	return err
}

//set the status of a service (0 = disabled;1=enabled)
func (service *Service) setStatus(newStatus SystemdCommand) error {
	return SetServiceStatus(service.Name, newStatus)
}

//SystemfileExists returns true if service exists
func SystemfileExists(name string) bool {
	name = NameToServiceFile(name)
	file := "/etc/systemd/system/" + name
	_, err := os.Stat(file)
	if err != nil {
		return false
	}
	return true
}

//Create creates a service file
func (service *Service) Create() error {
	if os.Getgid() != 0 {
		return errors.New("you need to be root")
	}

	content := service.Generate()
	name := NameToServiceFile(service.Name)
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

//NameToServiceFile returns the name of the servicefile
func NameToServiceFile(name string) string {
	if !strings.HasSuffix(name, ".service") {
		return name + ".service"
	}
	return name
}

//DaemonReload reloads the daemon
func DaemonReload() error {
	_, err := runCommand(nil, "systemctl daemon-reload")
	return err
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
			fieldKey := v.Type().Field(index).Name
			if v.Field(index).Kind() == reflect.String && len(value.String()) > 0 {
				final += fieldKey + "=" + value.String() + "\n"
			} else if v.Field(index).Kind() == reflect.Int && value.Int() != 0 {
				final += fieldKey + "=" + strconv.FormatInt(value.Int(), 10) + "\n"
			}
		}
	}
	return final
}

//Parse a service file to a Service scruct
func Parse(fileName string) *Service {
	if !SystemfileExists(NameToServiceFile(fileName)) {
		return nil
	}
	file, err := os.Open("/etc/systemd/system/" + fileName)
	if err != nil {
		return nil
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var unitPart, servicePart, installPart string
	var currentPart *string
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), "")
		if len(strings.Trim(line, " ")) == 0 {
			continue
		}
		if strings.HasPrefix(line, "[") {
			if line == "[Unit]" {
				currentPart = &unitPart
			} else if line == "[Service]" {
				currentPart = &servicePart
			} else if line == "[Install]" {
				currentPart = &installPart
			}
		} else {
			if currentPart != nil {
				*currentPart += line + "\n"
			}
		}
	}
	unitLines := strings.Split(unitPart, "\n")
	unit := Unit{}
	vu := reflect.ValueOf(&unit).Elem()
	for _, line := range unitLines {
		data := strings.Split(line, "=")
		if len(data) != 2 {
			continue
		}
		key := data[0]
		val := data[1]
		fi := vu.FieldByName(key)
		if fi != (reflect.Value{}) {
			fi.SetString(val)
		} else {
			fmt.Println(key)
		}
	}

	serviceLines := strings.Split(servicePart, "\n")
	sservice := SService{}
	vs := reflect.ValueOf(&sservice).Elem()
	for _, line := range serviceLines {
		data := strings.Split(line, "=")
		if len(data) != 2 {
			continue
		}
		key := data[0]
		val := data[1]
		fi := vs.FieldByName(key)
		if fi != (reflect.Value{}) {
			fi.SetString(val)
		} else {
			fmt.Println(key)
		}
	}
	installLines := strings.Split(installPart, "\n")
	install := Install{}
	vi := reflect.ValueOf(&install).Elem()
	for _, line := range installLines {
		data := strings.Split(line, "=")
		if len(data) != 2 {
			continue
		}
		key := data[0]
		val := data[1]
		fi := vi.FieldByName(key)
		if fi != (reflect.Value{}) {
			fi.SetString(val)
		} else {
			fmt.Println(key)
		}
	}
	return &Service{
		Name:    fileName,
		Unit:    unit,
		Service: sservice,
		Install: install,
	}
}
