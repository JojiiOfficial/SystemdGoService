package main

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
	//Notify notify service
	Notify ServiceType = "notify"
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
	Before              Target `name:"Before"`
	After               Target `name:"After"`
	Wants               Target `name:"Wants"`
	ConditionPathExisis string `name:"ConditionPathExists"`
}

//SService [Service] in .service file
type SService struct {
	Type      ServiceType
	ExecStart string
	User      string
	Group     string
	Restart   string
}

//Install [Install] in .service file
type Install struct {
	WantedBy Target
	Alias    string
}

//NewSimpleService creates a new default service
func NewSimpleService(name, description, execStart string) *Service {
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

//Stop stop
func (service *Service) Stop() {

}

//Start starts a service
func (service *Service) Start() {

}

//Enable a service
func (service *Service) Enable() {

}

func (service *Service) save() {

}
