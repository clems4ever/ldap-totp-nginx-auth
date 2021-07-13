package authentication

import "fmt"

func ldapBuildGroupsFilterFromGroupsAttribute(groups []string, distinguishedNameAttribute string) (filter string) {
	for _, group := range groups {
		filter += fmt.Sprintf("(%s=%s)", distinguishedNameAttribute, group)
	}

	if filter == "" {
		return filter
	}

	return fmt.Sprintf("(|%s)", filter)
}
