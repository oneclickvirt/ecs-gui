package appmeta

const (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.139"
	UpstreamECSVersion = "v0.1.139"
)

func ReleaseVersion() string {
	return "v" + Version
}
