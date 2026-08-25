package appmeta

var (
	AppID              = "com.oneclickvirt.goecs"
	AppName            = "goecs"
	Version            = "0.1.183"
	UpstreamECSVersion = "v0.1.179"
)

func ReleaseVersion() string {
	return "v" + Version
}
