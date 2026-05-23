package security

func HasPermission(p *Principal, permission string) bool {
	if p == nil {
		return false
	}
	for _, candidate := range p.Permissions {
		if candidate == permission {
			return true
		}
	}
	return false
}

func HasAnyPermission(p *Principal, permissions ...string) bool {
	for _, permission := range permissions {
		if HasPermission(p, permission) {
			return true
		}
	}
	return false
}

func HasRole(p *Principal, role string) bool {
	if p == nil {
		return false
	}
	for _, candidate := range p.Roles {
		if candidate == role {
			return true
		}
	}
	return false
}
