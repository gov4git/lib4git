package git

func IsAlreadyUpToDate(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "already up-to-date"
}

func IsRemoteRepoIsEmpty(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "remote repository is empty"
}
