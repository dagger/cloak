package container

type Container {
	rootfs *filesystem.Filesystem
}

func New(rootfs *filesystem.Filesystem) *Container {
	return &Container {
		rootfs: rootfs,
	}
})