package avi

const projectileTexture = "projectile"

type projectile struct {
	objectT
}

func (projectile) Texture() string {
	return projectileTexture
}
