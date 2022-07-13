package launcher

type LaunchTask struct {
	JavaBin     string
	JVMOptions  []string
	GameOptions []string
}

func CreateLaunchTask() LaunchTask {
	return LaunchTask{}
}

func (t *LaunchTask) AddJVMOption(option string) *LaunchTask {
	t.JVMOptions = append(t.JVMOptions, "-"+option)
	return t
}

func (t *LaunchTask) AddGameOption(option string) *LaunchTask {
	t.GameOptions = append(t.GameOptions, "--"+option)
	return t
}

func (t *LaunchTask) SetJavaBinPath(option string) *LaunchTask {
	t.JavaBin = option
	return t
}
